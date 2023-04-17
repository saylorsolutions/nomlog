package iterator

import (
	"context"
	"errors"
	"github.com/saylorsolutions/slog/pkg/entries"
	"golang.org/x/sync/semaphore"
)

var (
	ErrStopIteration = errors.New("stop iterating")
)

type Iterator interface {
	// Next returns the next LogEntry and its offset in the stream.
	// May return ErrStopIteration if the end of the stream is reached.
	Next() (entries.LogEntry, int, error)
	// Iterate will progress through all LogEntry items in the stream, calling iter for each one along with the offset.
	// If iter returns ErrStopIteration, then iteration will cease, returning nil.
	// If any other error is returned, then iteration will cease, and the error will be returned.
	Iterate(iter func(entry entries.LogEntry, i int) error) error
}

func FromSlice(entries []entries.LogEntry) Iterator {
	return &entrySlice{entries: entries}
}

func FromChannel(entries <-chan entries.LogEntry) Iterator {
	return &entryChannel{ch: entries}
}

func AsChannel(iter Iterator) <-chan entries.LogEntry {
	if chi, ok := iter.(*entryChannel); ok {
		return chi.ch
	}
	if chs, ok := iter.(*entrySlice); ok {
		ch := make(chan entries.LogEntry, len(chs.entries))
		defer close(ch)
		for i := 0; i < len(chs.entries); i++ {
			ch <- chs.entries[i]
		}
		return ch
	}
	ch := make(chan entries.LogEntry)
	go func() {
		defer close(ch)
		_ = iter.Iterate(func(entry entries.LogEntry, i int) error {
			ch <- entry
			return nil
		})
	}()
	return ch
}

// Merge will take over the passed in LogIterators and forward all LogEntry elements to the new Iterator.
// It's advised not to read from an iterator that has been passed to Merge.
func Merge(a, b Iterator) Iterator {
	aCh := AsChannel(a)
	bCh := AsChannel(b)

	outCh := make(chan entries.LogEntry)
	out := FromChannel(outCh)

	go func() {
		defer close(outCh)
		for aCh != nil || bCh != nil {
			select {
			case ae, ok := <-aCh:
				if !ok {
					aCh = nil
					continue
				}
				outCh <- ae
			case be, ok := <-bCh:
				if !ok {
					bCh = nil
					continue
				}
				outCh <- be
			}
		}
	}()
	return out
}

// Dupe will take control of and branch the duplicate Iterator into two identical LogIterators.
// Any LogEntry posted to the source Iterator will be sent to both of the new LogIterators.
// This is useful in a case similar to when you want to print messages as well as write them to a file.
// It's not advised to read from a Iterator that has been passed to Dupe, use oen of the returned LogIterators instead.
func Dupe(iter Iterator) (Iterator, Iterator) {
	if iter == nil {
		ch := make(chan entries.LogEntry)
		close(ch)
		return FromChannel(ch), FromChannel(ch)
	}

	a := make(chan entries.LogEntry)
	b := make(chan entries.LogEntry)
	aiter := FromChannel(a)
	biter := FromChannel(b)

	go func() {
		sem := semaphore.NewWeighted(2)
		ctx := context.Background()

		defer func() {
			_ = sem.Acquire(ctx, 2)
			close(a)
			close(b)
		}()
		_ = iter.Iterate(func(entry entries.LogEntry, i int) error {
			_ = sem.Acquire(ctx, 1)
			go func() {
				defer sem.Release(1)
				a <- entry
			}()
			_ = sem.Acquire(ctx, 1)
			go func() {
				defer sem.Release(1)
				b <- entry
			}()
			return nil
		})
	}()
	return aiter, biter
}

// Drain will drain all entries from a Iterator in a new goroutine.
// This can be useful as an error fallback in case of an iteration error to prevent upstream blocking.
func Drain(iter Iterator) {
	ch := AsChannel(iter)
	go func() {
		for range ch {
		}
	}()
}

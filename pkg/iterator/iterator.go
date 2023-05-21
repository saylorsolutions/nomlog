package iterator

import (
	"context"
	"errors"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"golang.org/x/sync/semaphore"
)

var (
	ErrAtEnd = errors.New("end of iterator")
)

type Iterator interface {
	// Next returns the next LogEntry and its offset in the stream.
	// May return ErrAtEnd if the end of the stream is reached.
	Next() (entries.LogEntry, int, error)
	// Iterate will progress through all LogEntry items in the stream, calling iter for each one along with the offset.
	// If iter returns ErrAtEnd, then iteration will cease, returning nil.
	// If any other error is returned, then iteration will cease, and the error will be returned.
	Iterate(iter func(entry entries.LogEntry, i int) error) error
}

var _ Iterator = (Func)(nil)

// Func provides a quicker way to implement an Iterator.
// A Func implements Iterator.Next, and implicitly provides a base Iterate implementation.
type Func func() (entries.LogEntry, int, error)

func (f Func) Next() (entries.LogEntry, int, error) {
	return f()
}

func (f Func) Iterate(iter func(entry entries.LogEntry, i int) error) error {
	for {
		entry, i, err := f.Next()
		if err != nil {
			if errors.Is(err, ErrAtEnd) {
				return nil
			}
			return err
		}
		if err := iter(entry, i); err != nil {
			if errors.Is(err, ErrAtEnd) {
				Drain(f)
				return nil
			}
			return err
		}
	}
}

// Err makes it easier to return the standard response for Iterator.Next errors.
func Err(err error) (entries.LogEntry, int, error) {
	return nil, -1, err
}

func FromSlice(slice []entries.LogEntry) Iterator {
	cur := 0
	return Func(func() (entries.LogEntry, int, error) {
		if cur >= len(slice) {
			return nil, -1, ErrAtEnd
		}
		e := slice[cur]
		i := cur
		cur++
		return e, i, nil
	})
}

func FromChannel(entries <-chan entries.LogEntry) Iterator {
	if entries == nil {
		return Empty()
	}
	return &entryChannel{ch: entries}
}

func AsChannel(iter Iterator) <-chan entries.LogEntry {
	if chi, ok := iter.(*entryChannel); ok {
		return chi.ch
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

// Merge will take over the passed in iterators and forward all log entries elements to the new Iterator.
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

// Dupe will take control of and branch the duplicate Iterator into two identical iterators.
// Any LogEntry posted to the source Iterator will be sent to both of the new iterators.
// This is useful in a case similar to when you want to print messages as well as write them to a file.
// It's not advised to read from an Iterator that has been passed to Dupe, use one of the returned iterators instead.
func Dupe(iter Iterator) (Iterator, Iterator) {
	if iter == nil {
		return Empty(), Empty()
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

// Fanout will take control of the input Iterator and output entries received from the input Iterator to one of the output Iterators.
// It's not advised to read from the input Iterator after passing it to Fanout.
// If an error occurs during iteration, then Drain will be called on the input.
func Fanout(iter Iterator) (Iterator, Iterator) {
	if iter == nil {
		return Empty(), Empty()
	}

	a := make(chan entries.LogEntry)
	b := make(chan entries.LogEntry)
	go func() {
		defer func() {
			close(a)
			close(b)
		}()
		err := iter.Iterate(func(entry entries.LogEntry, i int) error {
			select {
			case a <- entry:
			case b <- entry:
			}
			return nil
		})
		if err != nil {
			Drain(iter)
		}
	}()
	return FromChannel(a), FromChannel(b)
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

// Empty returns an empty Iterator that immediately returns from a call to Iterate.
func Empty() Iterator {
	return FromSlice(nil)
}

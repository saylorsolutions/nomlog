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
	// Should return ErrAtEnd if the end of the stream is reached.
	Next() (entries.LogEntry, int, error)
	// Iterate will progress through all LogEntry items in the stream, calling iter for each one along with the offset.
	// If iter returns ErrAtEnd, then iteration will cease, returning a nil error.
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
			if IsEnd(err) {
				return nil
			}
			return err
		}
		if err := iter(entry, i); err != nil {
			if IsEnd(err) {
				return nil
			}
			Drain(f)
			return err
		}
	}
}

// Err makes it easier to return the standard response for Iterator.Next errors.
func Err(err error) (entries.LogEntry, int, error) {
	return nil, -1, err
}

// End signifies that the end of the stream has been reached.
func End() (entries.LogEntry, int, error) {
	return Err(ErrAtEnd)
}

// IsEnd returns whether the error indicates the end of a stream.
func IsEnd(err error) bool {
	return errors.Is(err, ErrAtEnd)
}

func FromSlice(slice []entries.LogEntry) Iterator {
	cur := 0
	return Func(func() (entries.LogEntry, int, error) {
		if cur >= len(slice) {
			return End()
		}
		e := slice[cur]
		i := cur
		cur++
		return e, i, nil
	})
}

// FromChannel will create a new Iterator from a channel of entries.LogEntry.
func FromChannel(_entries <-chan entries.LogEntry) Iterator {
	var next int
	return Func(func() (entries.LogEntry, int, error) {
		entry, ok := <-_entries
		if !ok {
			return End()
		}
		cur := next
		next++
		return entry, cur, nil
	})
}

// AsChannel will create a channel that is populated from the Iterator by a new goroutine.
// If bufferSize is populated, then the returned channel will be a buffered channel.
func AsChannel(iter Iterator, bufferSize ...int) <-chan entries.LogEntry {
	var ch chan entries.LogEntry
	if len(bufferSize) == 0 {
		ch = make(chan entries.LogEntry)
	} else {
		ch = make(chan entries.LogEntry, bufferSize[0])
	}
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

package iterator

import (
	"context"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"sync"
)

// Filter wraps an Iterator with a function that - when it returns true - will allow the return values of Next through.
// If the wrapped Iterator returns a non-nil error, then all values will be passed through regardless
func Filter(iter Iterator, filter func(entry entries.LogEntry, i int, err error) bool) Iterator {
	return Func(func() (entries.LogEntry, int, error) {
		for {
			entry, idx, err := iter.Next()
			if err != nil {
				return entry, idx, err
			}
			if filter(entry, idx, err) {
				return entry, idx, err
			}
		}
	})
}

// Cancellable wraps an iterator and makes it cancellable by context.
// When the context is cancelled and Next is called, all LogEntries will be forwarded to Drain.
func Cancellable(ctx context.Context, iter Iterator) Iterator {
	var (
		cancelled bool
		drain     sync.Once
	)
	go func() {
		<-ctx.Done()
		cancelled = true
	}()
	return Func(func() (entries.LogEntry, int, error) {
		if cancelled {
			drain.Do(func() {
				Drain(iter)
			})
			return End()
		}
		return iter.Next()
	})
}

// Concat will return entries from next after base has been exhausted.
func Concat(base, next Iterator) Iterator {
	var idx int
	return Func(func() (entries.LogEntry, int, error) {
		e, i, err := base.Next()
		if err != nil {
			if IsEnd(err) {
				e, i, err := next.Next()
				if err != nil {
					return e, i, err
				}
				return e, i + idx, err
			}
			return e, i, err
		}
		idx++
		return e, i, err
	})
}

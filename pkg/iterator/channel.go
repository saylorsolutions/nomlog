package iterator

import (
	"errors"
	"github.com/saylorsolutions/nomlog/pkg/entries"
)

var _ Iterator = (*entryChannel)(nil)

type entryChannel struct {
	ch   <-chan entries.LogEntry
	next int
}

func (e *entryChannel) Next() (entries.LogEntry, int, error) {
	entry, ok := <-e.ch
	if !ok {
		return nil, -1, ErrStopIteration
	}
	cur := e.next
	e.next += 1
	return entry, cur, nil
}

func (e *entryChannel) Iterate(iter func(entry entries.LogEntry, i int) error) error {
	for {
		entry, i, err := e.Next()
		if err != nil {
			if errors.Is(err, ErrStopIteration) {
				return nil
			}
			return err
		}
		if err := iter(entry, i); err != nil {
			return err
		}
	}
}

package iterator

import (
	"errors"
	"github.com/saylorsolutions/nomlog/pkg/entries"
)

var _ Iterator = (*entrySlice)(nil)

type entrySlice struct {
	entries []entries.LogEntry
	next    int
}

func (e *entrySlice) Next() (entries.LogEntry, int, error) {
	cur := e.next
	if len(e.entries) > cur {
		e.next += 1
		return e.entries[cur], cur, nil
	}
	return nil, -1, ErrStopIteration
}

func (e *entrySlice) Iterate(iter func(entry entries.LogEntry, i int) error) error {
	entry, i, err := e.Next()
	for ; err == nil; entry, i, err = e.Next() {
		entry := entry
		i := i
		err = iter(entry, i)
		if err != nil {
			break
		}
	}
	if errors.Is(err, ErrStopIteration) {
		return nil
	}
	return err
}

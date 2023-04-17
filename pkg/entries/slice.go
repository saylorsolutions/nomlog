package entries

import "errors"

var _ LogIterator = (*entrySlice)(nil)

type entrySlice struct {
	entries []LogEntry
	next    int
}

func (e *entrySlice) Next() (LogEntry, int, error) {
	cur := e.next
	if len(e.entries) > cur {
		e.next += 1
		return e.entries[cur], cur, nil
	}
	return nil, -1, ErrStopIteration
}

func (e *entrySlice) Iterate(iter func(entry LogEntry, i int) error) error {
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

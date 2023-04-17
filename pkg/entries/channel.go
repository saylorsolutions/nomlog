package entries

var _ LogIterator = (*entryChannel)(nil)

type entryChannel struct {
	ch   <-chan LogEntry
	next int
}

func (e *entryChannel) Next() (LogEntry, int, error) {
	entry, ok := <-e.ch
	if !ok {
		return nil, -1, ErrStopIteration
	}
	cur := e.next
	e.next += 1
	return entry, cur, nil
}

func (e *entryChannel) Iterate(iter func(entry LogEntry, i int) error) error {
	for {
		entry, i, err := e.Next()
		if err != nil {
			return err
		}
		if err := iter(entry, i); err != nil {
			return err
		}
	}
}

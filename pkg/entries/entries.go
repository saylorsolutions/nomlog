package entries

import "errors"

// LogEntry is a single entry in a log, with potentially many fields.
type LogEntry map[string]any

func (e LogEntry) HasField(name string) bool {
	_, ok := e[name]
	return ok
}

var (
	ErrStopIteration = errors.New("stop iterating")
)

type LogIterator interface {
	// Next returns the next LogEntry and its offset in the stream.
	// May return ErrStopIteration if the end of the stream is reached.
	Next() (LogEntry, int, error)
	// Iterate will progress through all LogEntry items in the stream, calling iter for each one along with the offset.
	// If iter returns ErrStopIteration, then iteration will cease, returning nil.
	// If any other error is returned, then iteration will cease, and the error will be returned.
	Iterate(iter func(entry LogEntry, i int) error) error
}

func NewSliceIterator(entries []LogEntry) LogIterator {
	return &entrySlice{entries: entries}
}

func NewChannelIterator(entries <-chan LogEntry) LogIterator {
	return &entryChannel{ch: entries}
}

func NewIterationChannel(iter LogIterator) <-chan LogEntry {
	if chi, ok := iter.(*entryChannel); ok {
		return chi.ch
	}
	if chs, ok := iter.(*entrySlice); ok {
		ch := make(chan LogEntry, len(chs.entries))
		defer close(ch)
		for i := 0; i < len(chs.entries); i++ {
			ch <- chs.entries[i]
		}
		return ch
	}
	ch := make(chan LogEntry)
	go func() {
		defer close(ch)
		_ = iter.Iterate(func(entry LogEntry, i int) error {
			ch <- entry
			return nil
		})
	}()
	return ch
}

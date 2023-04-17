package file

import (
	"encoding/json"
	"github.com/nxadm/tail"
	"github.com/saylorsolutions/slog/pkg/entries"
	"time"
)

const (
	defaultMessageField = "@message"
	readTimeField       = "@read_timestamp"
	readLineField       = "@read_line_number"
)

// Source will create an entries.LogIterator that contains lines from the provided log file.
// If the file is structured as JSON data, then the individual fields of the line will be merged into the entries.LogEntry.
// Otherwise, a @message field will be populated with the entire line.
func Source(filename string) (entries.LogIterator, error) {
	_, i, err := source(filename)
	if err != nil {
		return nil, err
	}
	return i, err
}

func source(filename string) (*tail.Tail, entries.LogIterator, error) {
	t, err := tail.TailFile(filename, tail.Config{
		ReOpen:    true,
		MustExist: true,
		Follow:    true,
	})
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan entries.LogEntry)
	go func() {
		defer close(ch)
		for l := range t.Lines {
			entry := entries.LogEntry{
				readTimeField: l.Time.Format(time.RFC3339),
				readLineField: l.Num,
			}
			err := json.Unmarshal([]byte(l.Text), &entry)
			if err != nil {
				entry[defaultMessageField] = l.Text
			}
			ch <- entry
		}
	}()
	return t, entries.NewChannelIterator(ch), nil
}

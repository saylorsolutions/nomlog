package file

import (
	"context"
	"encoding/json"
	"github.com/nxadm/tail"
	"github.com/saylorsolutions/slog/pkg/entries"
	"os"
	"time"
)

const (
	defaultMessageField = "@message"
	readTimeField       = "@read_timestamp"
	readLineField       = "@read_line_number"
)

// Source behaves the same as CtxSource, except that it will use context.Background as the context.
func Source(filename string) (entries.LogIterator, error) {
	_, i, err := ctxSource(context.Background(), filename)
	return i, err
}

// CtxSource will create an entries.LogIterator that contains lines from the provided log file.
// If the file is structured as JSON data, then the individual fields of the line will be merged into the entries.LogEntry.
// Otherwise, a @message field will be populated with the entire line.
func CtxSource(ctx context.Context, filename string) (entries.LogIterator, error) {
	_, i, err := ctxSource(ctx, filename)
	return i, err
}

func ctxSource(ctx context.Context, filename string) (*tail.Tail, entries.LogIterator, error) {
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
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-t.Lines:
				if !ok {
					return
				}
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
		}
	}()
	return t, entries.NewChannelIterator(ch), nil
}

// Sink will append each entry in the entries.LogIterator to the specified file, creating it if necessary.
// If Sink is called asynchronously, it's recommended to wait until it returns to close down the application.
// This can be done with CtxSource by cancelling the provided context and waiting on the goroutine calling Sink to exit.
func Sink(iter entries.LogIterator, filename string, perms os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return iter.Iterate(func(entry entries.LogEntry, _ int) error {
		data, err := json.Marshal(entry)
		if err != nil {
			// Shouldn't ever happen, given the data type.
			return err
		}
		_, err = f.Write(data)
		if err != nil {
			return err
		}
		return nil
	})
}

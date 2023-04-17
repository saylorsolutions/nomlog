package file

import (
	"context"
	"encoding/json"
	"github.com/nxadm/tail"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"os"
	"time"
)

const (
	readTimeField = "@read_timestamp"
	readLineField = "@read_line_number"
)

// Source behaves the same as CtxSource, except that it will use context.Background as the context.
func Source(filename string) (iterator.Iterator, error) {
	_, i, err := ctxSource(context.Background(), filename)
	return i, err
}

// CtxSource will create an entries.Iterator that contains lines from the provided log file.
// If the file is structured as JSON data, then the individual fields of the line will be merged into the entries.LogEntry.
// Otherwise, a @message field will be populated with the entire line.
func CtxSource(ctx context.Context, filename string) (iterator.Iterator, error) {
	_, i, err := ctxSource(ctx, filename)
	return i, err
}

func ctxSource(ctx context.Context, filename string) (*tail.Tail, iterator.Iterator, error) {
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
				_ = t.Stop()
				return
			case l, ok := <-t.Lines:
				if !ok {
					return
				}
				entry := entries.FromString(l.Text)
				entry[readTimeField] = l.Time.UTC().Format(time.RFC3339)
				entry[readLineField] = l.Num
				ch <- entry
			}
		}
	}()
	return t, iterator.FromChannel(ch), nil
}

// Sink will append each entry in the entries.Iterator to the specified file, creating it if necessary.
// If Sink is called asynchronously, it's recommended to wait until it returns to close down the application.
// This can be done with CtxSource by cancelling the provided context and waiting on the goroutine calling Sink to exit.
// In case of an error, Sink will drain the entries.Iterator to prevent upstream blocking.
func Sink(iter iterator.Iterator, filename string, perms os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		iterator.Drain(iter)
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	err = iter.Iterate(func(entry entries.LogEntry, _ int) error {
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
	if err != nil {
		iterator.Drain(iter)
		return err
	}
	return nil
}
package file

import (
	"bufio"
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

// TailSource behaves the same as CtxTailSource, except that it will use context.Background as the context.
// This means that the goroutine will be tailing the file for the entire life of the program.
func TailSource(filename string) (iterator.Iterator, error) {
	_, i, err := ctxTailSource(context.Background(), filename)
	return i, err
}

// CtxTailSource will create an iterator.Iterator that contains lines from the provided log file.
// If the file is structured as JSON data, then the individual fields of the line will be merged into the entries.LogEntry.
// Otherwise, a @message field will be populated with the entire line.
func CtxTailSource(ctx context.Context, filename string) (iterator.Iterator, error) {
	_, i, err := ctxTailSource(ctx, filename)
	return i, err
}

func ctxTailSource(ctx context.Context, filename string) (*tail.Tail, iterator.Iterator, error) {
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

// Source operates the same as CtxSource, except that it will use the background context for cancellation.
func Source(filename string) (iterator.Iterator, error) {
	return CtxSource(context.Background(), filename)
}

// CtxSource will create an iterator.Iterator from all lines in the specified file in a new goroutine.
// If there is an error opening the file, then it will be reported from this method.
// If the given context is cancelled while the goroutine is reading, then it will stop and close the file and iterator.
func CtxSource(ctx context.Context, filename string) (iterator.Iterator, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	_ctx, cancel := context.WithCancel(ctx)
	ch := make(chan entries.LogEntry)
	iter := iterator.FromChannel(ch)

	go func() {
		scanner := bufio.NewScanner(f)
		defer func() {
			cancel()
			close(ch)
			_ = f.Close()
		}()

		var (
			hasClosed bool
			num       int
		)
		go func() {
			<-_ctx.Done()
			hasClosed = true
		}()

		for scanner.Scan() {
			if hasClosed {
				return
			}
			line := scanner.Text()
			entry := entries.FromString(line)
			entry[readTimeField] = time.Now().UTC().Format(time.RFC3339)
			entry[readLineField] = num
			num++
			ch <- entry
		}
	}()
	return iter, nil
}

// Sink will append each entry in the iterator.Iterator to the specified file, creating it if necessary.
// If Sink is called asynchronously, it's recommended to wait until it returns to close down the application.
// This can be done with CtxTailSource by cancelling the provided context and waiting on the goroutine calling Sink to exit.
// In case of an error, Sink will drain the iterator.Iterator to prevent upstream blocking.
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

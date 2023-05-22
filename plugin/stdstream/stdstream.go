package stdstream

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/saylorsolutions/nomlog/plugin"
	"os"
)

var _ plugin.Plugin = (*stdplugin)(nil)

type stdplugin struct {
}

func (s *stdplugin) ID() string {
	return "std"
}

func (s *stdplugin) Register(reg *plugin.Registration) {
	reg.RegisterSource("std", "In", SourceIn)
	reg.DocumentSource("std", "In", `std.In

Reads each line of STDIN as a log entry. The input may be a valid JSON object, or completely unstructured.`)
	reg.RegisterSink("std", "Out", SinkOut)
	reg.DocumentSink("std", "Out", `std.Out

Writes each log entry as a line to STDOUT.`)
	reg.RegisterSink("std", "Err", SinkErr)
	reg.DocumentSink("std", "Err", `std.Err

Writes each log entry as a line to STDERR.`)
}

func (s *stdplugin) Stopping() error {
	return nil
}

func SourceIn(ctx context.Context, _ ...*dsl.Arg) (iterator.Iterator, error) {
	ch := make(chan entries.LogEntry)
	go func() {
		defer func() {
			close(ch)
		}()
		scanner := bufio.NewScanner(os.Stdin)

		var hasClosed bool
		go func() {
			<-ctx.Done()
			hasClosed = true
		}()

		for scanner.Scan() {
			if hasClosed {
				return
			}
			line := scanner.Text()
			entry := entries.FromString(line)
			ch <- entry
		}
	}()
	return iterator.FromChannel(ch), nil
}

func jsonify(entry entries.LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SinkOut(ctx context.Context, src iterator.Iterator, _ ...*dsl.Arg) error {
	var hasCancelled bool
	go func() {
		<-ctx.Done()
		hasCancelled = true
	}()
	err := src.Iterate(func(entry entries.LogEntry, i int) error {
		if hasCancelled {
			return iterator.ErrAtEnd
		}
		str, err := jsonify(entry)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "%s\n", str)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		iterator.Drain(src)
		return err
	}
	return nil
}
func SinkErr(ctx context.Context, src iterator.Iterator, _ ...*dsl.Arg) error {
	var hasCancelled bool
	go func() {
		<-ctx.Done()
		hasCancelled = true
	}()
	err := src.Iterate(func(entry entries.LogEntry, i int) error {
		if hasCancelled {
			return iterator.ErrAtEnd
		}
		str, err := jsonify(entry)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stderr, "%s\n", str)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		iterator.Drain(src)
		return err
	}
	return nil
}

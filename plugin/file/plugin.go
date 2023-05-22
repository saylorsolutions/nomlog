package file

import (
	"context"
	"fmt"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/saylorsolutions/nomlog/plugin"
	"github.com/saylorsolutions/nomlog/runtime/dsl"
	"os"
	"strconv"
)

func Plugin() plugin.Plugin {
	return new(filePlugin)
}

type filePlugin struct{}

func (*filePlugin) ID() string {
	return "file"
}

func (*filePlugin) Stopping() error {
	return nil
}

func (*filePlugin) Register(reg *plugin.Registration) {
	reg.RegisterSource("file", "Tail", func(ctx context.Context, args ...*dsl.Arg) (iterator.Iterator, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("%w: requires 1 argument", plugin.ErrArgs)
		}
		return CtxTailSource(ctx, args[0].String)
	})
	reg.DocumentSource("file", "Tail", `file.Tail FILE_NAME

This source will watch the file specified by FILE_NAME for changes, producing a new log entry for each new line.
Just like the file.File source, structured or unstructured data may be read.`)
	reg.RegisterSource("file", "File", func(ctx context.Context, args ...*dsl.Arg) (iterator.Iterator, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("%w: requires 1 argument", plugin.ErrArgs)
		}
		return CtxSource(ctx, args[0].String)
	})
	reg.DocumentSource("file", "File", `file.File FILE_NAME

This source will read each line of the file specified by FILE_NAME, emitting a log entry for each one.
If the line represents a valid JSON document, then it will be emitted as-is except with additional fields specifying read timing.
Otherwise, the line is added as-is to a log entry with a field "@message" containing the original line.`)
	reg.RegisterSink("file", "File", func(_ context.Context, src iterator.Iterator, args ...*dsl.Arg) error {
		if len(args) < 1 {
			return fmt.Errorf("%w: requires 1 or 2 arguments", plugin.ErrArgs)
		}

		if len(args) >= 2 {
			perms, err := strconv.ParseUint(args[1].String, 8, 32)
			if err != nil {
				return fmt.Errorf("%w: invalid file permission argument", plugin.ErrArgs)
			}
			return Sink(src, args[0].String, os.FileMode(perms))
		}
		return Sink(src, args[0].String, 0600)
	})
	reg.DocumentSink("file", "File", `file.File FILE_NAME [FILE_MODE]

This sink will append each log entry as a JSON document on a single line to a file specified by FILE_NAME, creating it if necessary.
If FILE_MODE is specified, and it's a string representing a valid octal file mode like "644", then this mode will be used to create the file if it doesn't already exist.
If FILE_MODE is specified but invalid, then the sink operation will fail.
If FILE_MODE is not specified, then a value of "600" will be assumed.
The file's permissions will not be modified if it already exists.'`)
}

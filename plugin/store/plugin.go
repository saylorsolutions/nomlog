package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/saylorsolutions/nomlog/plugin"
	"strings"
)

func Plugin() plugin.Plugin {
	return &sqlitePlugin{
		storeCache: map[string]*SqliteStore{},
	}
}

type sqlitePlugin struct {
	storeCache map[string]*SqliteStore
}

func (p *sqlitePlugin) Closing() error {
	closeErrors := make(chan error, len(p.storeCache))
	for file, store := range p.storeCache {
		if err := store.Close(); err != nil {
			closeErrors <- fmt.Errorf("%w: file: %s", err, file)
		}
	}
	close(closeErrors)
	var buf strings.Builder
	buf.WriteString("error closing SQLite plugin: ")

	i := 0
	for err := range closeErrors {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(err.Error())
		i++
	}
	return errors.New(buf.String())
}

func (p *sqlitePlugin) Register(reg *plugin.Registration) {
	reg.RegisterSource("sqlite", "Table", func(ctx context.Context, args ...dsl.Arg) (iterator.Iterator, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("%w: requires 2 argument", plugin.ErrArgs)
		}
		file := args[0].String
		if file == "" {
			return nil, fmt.Errorf("%w: file name string must be specified as first argument", plugin.ErrArgs)
		}
		table := args[1].String
		if table == "" {
			return nil, fmt.Errorf("%w: table name string must be specified as second argument", plugin.ErrArgs)
		}
		store, ok := p.storeCache[file]
		if !ok {
			_store, err := NewStore(hclog.Default(), file)
			if err != nil {
				return nil, err
			}
			store = _store
			p.storeCache[file] = _store
		}
		return store.CtxQueryEntries(ctx, table)
	})
	reg.DocumentSource("sqlite", "Table", `sqlite.Table FILE_NAME TABLE_NAME

This source will query all rows from a table and return each row as a log entry.
It may not return continuously added rows, so it should be used for tables that represent a static snapshot of log entries.`)
	reg.RegisterSink("sqlite", "Table", func(ctx context.Context, src iterator.Iterator, args ...dsl.Arg) error {
		if len(args) < 2 {
			return fmt.Errorf("%w: requires 2 argument", plugin.ErrArgs)
		}
		file := args[0].String
		if file == "" {
			return fmt.Errorf("%w: file name string must be specified as first argument", plugin.ErrArgs)
		}
		table := args[1].String
		if table == "" {
			return fmt.Errorf("%w: table name string must be specified as second argument", plugin.ErrArgs)
		}
		store, ok := p.storeCache[file]
		if !ok {
			_store, err := NewStore(hclog.Default(), file)
			if err != nil {
				return err
			}
			store = _store
			p.storeCache[file] = _store
		}
		return store.CtxSink(ctx, src, table)
	})
	reg.DocumentSink("sqlite", "Table", `sqlite.Table FILE_NAME TABLE_NAME

This sink will land all log entries into the SQLite database table specified. The TABLE_NAME argument may be prefixed with a schema name like "my_schema.my_table".
If the table does not exist, then it will be created with an integer primary key column called evt_id. Table columns will be created as needed, one for each log entry field.
This means that the table may trend toward being sparsely populated if the input entries are largely heterogeneous.`)
}

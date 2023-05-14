package store

import (
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
	reg.RegisterSource("sqlite", "Table", func(args ...dsl.Arg) (iterator.Iterator, error) {
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
		return store.QueryEntries(table)
	})
	reg.RegisterSink("sqlite", "Table", func(src iterator.Iterator, args ...dsl.Arg) error {
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
		return store.Sink(src, table)
	})
}

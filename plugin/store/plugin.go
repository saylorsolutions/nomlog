package store

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/saylorsolutions/nomlog/plugin"
)

var (
	storeCache = map[string]*SqliteStore{}
)

func Register(reg *plugin.Registration) {
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
		store, ok := storeCache[file]
		if !ok {
			_store, err := NewStore(hclog.Default(), file)
			if err != nil {
				return nil, err
			}
			store = _store
			storeCache[file] = _store
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
		store, ok := storeCache[file]
		if !ok {
			_store, err := NewStore(hclog.Default(), file)
			if err != nil {
				return err
			}
			store = _store
			storeCache[file] = _store
		}
		return store.Sink(src, table)
	})
}

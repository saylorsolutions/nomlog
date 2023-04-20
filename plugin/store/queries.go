package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
)

const (
	createTable = `
create table if not exists %s (
	evt_id integer primary key
)`
)

var (
	ErrUnexpectedColumnType = errors.New("unexpected column type")
)

func newQueryIterator(log hclog.Logger, rows *sql.Rows) (iterator.Iterator, error) {
	cols, err := rows.Columns()
	if err != nil {
		log.Error("Failed to query parameters", "error", err)
		return nil, err
	}
	var rowNum int

	if len(cols) == 0 {
		return iterator.Func(func() (entries.LogEntry, int, error) {
			return nil, -1, iterator.ErrAtEnd
		}), nil
	}

	return iterator.Func(func() (entries.LogEntry, int, error) {
		if !rows.Next() {
			_ = rows.Close()
			return nil, -1, iterator.ErrAtEnd
		}
		rowNum++
		var rowID int
		vals := make([]any, len(cols))
		if cols[0] == "evt_id" {
			vals[0] = &rowID
			for i := 1; i < len(vals); i++ {
				vals[i] = &sql.NullString{}
			}
		} else {
			for i := range vals {
				vals[i] = &sql.NullString{}
			}
		}
		if err := rows.Scan(vals...); err != nil {
			_ = rows.Close()
			return nil, -1, err
		}

		entry := entries.LogEntry{}
		for i, v := range vals {
			switch s := v.(type) {
			case *sql.NullString:
				if s.Valid {
					entry[cols[i]] = s.String
				}
			case *int:
				if s != nil {
					entry[cols[i]] = *s
				}
			default:
				return nil, -1, fmt.Errorf("%w: %T", ErrUnexpectedColumnType, v)
			}
		}
		return entry, rowNum, nil
	}), nil
}

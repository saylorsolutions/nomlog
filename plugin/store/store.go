package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	_ "modernc.org/sqlite"
	"regexp"
	"strings"
)

// SqliteStore is a store for LogEntries using Sqlite3 as a storage engine.
type SqliteStore struct {
	db  *sql.DB
	log hclog.Logger
}

func NewStore(log hclog.Logger, filename string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, err
	}
	log = log.Named("sqlite-entry-store")
	return &SqliteStore{
		db:  db,
		log: log,
	}, nil
}

func (s *SqliteStore) QueryEntries(table string) (iterator.Iterator, error) {
	if !tablePattern.MatchString(table) {
		return nil, fmt.Errorf("%w: %s", ErrBadTable, table)
	}
	rows, err := s.db.Query("select * from " + table)
	if err != nil {
		return nil, err
	}
	return newQueryIterator(s.log, rows)
}

func (s *SqliteStore) Sink(iter iterator.Iterator, table string) error {
	return s.SinkCtx(context.Background(), iter, table)
}

var (
	tablePattern = regexp.MustCompile(`^[\w\d]+(\.[\w\d]+)?$`)
	ErrBadTable  = errors.New("invalid table name")
)

func (s *SqliteStore) SinkCtx(ctx context.Context, iter iterator.Iterator, table string) error {
	if !tablePattern.MatchString(table) {
		return fmt.Errorf("%w: %s", ErrBadTable, table)
	}
	s.log.Debug("Establishing connection")
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	s.log.Debug("Ensuring the specified table is present")
	if err := s.ensureTable(ctx, conn, table); err != nil {
		iterator.Drain(iter)
		_ = conn.Close()
		return err
	}
	s.log.Debug("Getting table columns")
	cols, err := s.getTableColumns(ctx, conn, table)
	if err != nil {
		iterator.Drain(iter)
		_ = conn.Close()
		return err
	}
	colMap := map[string]bool{}
	for _, c := range cols {
		colMap[c] = true
	}

	s.log.Debug("Starting sink operation")
	s.sink(ctx, conn, table, iter, colMap)
	return nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

func (s *SqliteStore) ensureTable(ctx context.Context, conn *sql.Conn, table string) error {
	stmt, err := conn.PrepareContext(ctx, fmt.Sprintf(createTable, table))
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.ExecContext(ctx, table)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteStore) getTableColumns(ctx context.Context, conn *sql.Conn, table string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, "select * from "+table)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	return rows.Columns()
}

func (s *SqliteStore) sink(ctx context.Context, conn *sql.Conn, table string, iter iterator.Iterator, colMap map[string]bool) {
	log := s.log.With("table", table).Named("sink")
	cancelled := false

	defer func() {
		_ = conn.Close()
		log.Debug("DB connection closed")
	}()

	go func() {
		<-ctx.Done()
		log.Debug("Context cancelled")
		cancelled = true
	}()

	err := iter.Iterate(func(entry entries.LogEntry, i int) error {
		log.Debug("Received log entry", "entry", entry)
		if cancelled {
			return iterator.ErrAtEnd
		}

		var intoFields []string
		for k := range entry {
			if !colMap[k] {
				log.Debug("New field discovered, adding to table", "field", k)
				err := s.addColumn(ctx, conn, table, k)
				if err != nil {
					log.Error("Failed to add field to table", "field", k, "error", err)
					return err
				}
				colMap[k] = true
			}
			intoFields = append(intoFields, k)
		}

		var intoStr strings.Builder
		var params strings.Builder
		for i, f := range intoFields {
			if i > 0 {
				intoStr.WriteString(",")
				params.WriteString(",")
			}
			intoStr.WriteString("\"" + f + "\"")
			params.WriteString("?")
		}
		query := fmt.Sprintf("insert into %s (%s) values (%s)", table, intoStr.String(), params.String())
		log.Debug("Inserting log entry into table", "query", query)
		stmt, err := conn.PrepareContext(ctx, query)
		if err != nil {
			log.Error("Failed to prepare statement", "error", err)
			return err
		}
		defer func() {
			_ = stmt.Close()
		}()
		args := make([]any, len(intoFields))
		for i, f := range intoFields {
			str, ok := entry.AsString(f)
			if !ok {
				args[i] = ""
				log.Warn("Field not able to be coerced to string", "field", f)
				continue
			}
			args[i] = str
		}
		_, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			log.Error("Failed to insert into table", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("Error sinking to DB, draining iterator", "error", err)
		iterator.Drain(iter)
		return
	}
}

func (s *SqliteStore) addColumn(ctx context.Context, conn *sql.Conn, table string, colName string) error {
	_, err := conn.ExecContext(ctx, fmt.Sprintf("alter table %s add column \"%s\" text null", table, colName))
	if err != nil {
		return err
	}
	return nil
}

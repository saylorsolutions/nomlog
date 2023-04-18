package store

import (
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestSqliteStore_Sink(t *testing.T) {
	iter := iterator.FromSlice([]entries.LogEntry{
		{
			"A":           "A",
			"other-field": "value",
		},
		{
			"A": "A",
			"B": "B",
		},
		{
			"A": "A",
			"B": "B",
			"C": "C",
		},
	})
	log := hclog.Default()
	log.SetLevel(hclog.Debug)
	store, cleanup := _tempStore(t, log)
	defer cleanup()
	err := store.Sink(iter, "test")
	assert.NoError(t, err)
}

func _tempStore(t *testing.T, log hclog.Logger) (*SqliteStore, func()) {
	td, err := os.MkdirTemp("", "_tempStore-*")
	require.NoError(t, err)
	t.Log("Using temp store:", td)
	store, err := NewStore(log, filepath.Join(td, "store.db"))
	if err != nil {
		_ = os.RemoveAll(td)
		t.Fatal("Failed to create new store:", err)
	}

	return store, func() {
		if err := store.Close(); err != nil {
			t.Error("Failed to close DB")
		} else {
			t.Log("SqliteStore closed")
		}
		if err := os.RemoveAll(td); err != nil {
			t.Error("Failed to remove temp dir:", err)
		} else {
			t.Log("Removed temp dir")
		}
	}
}

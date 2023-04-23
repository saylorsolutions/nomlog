package file

import (
	"context"
	"encoding/json"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestSource_Structured(t *testing.T) {
	_tail, iter, err := ctxSource(context.Background(), "structured.log")
	require.NoError(t, err)
	require.NotNil(t, _tail)
	require.NotNil(t, iter)

	count := 0
	err = iter.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		assert.True(t, entry.HasField(entries.StandardMessageField), "Entry should have '@message' field")
		assert.True(t, entry.HasField(entries.StandardTimestampField), "Entry should have '@timestamp' field")
		assert.True(t, entry.HasField("@read_timestamp"), "Entry should have '@read_timestamp' field")
		assert.True(t, entry.HasField("@read_line_number"), "Entry should have '@read_line_number' field")
		switch count {
		case 1:
			assert.Equal(t, "A", entry[entries.StandardMessageField], "'A' should have been parsed as the message")
		case 2:
			assert.Equal(t, "B", entry[entries.StandardMessageField], "'B' should have been parsed as the message")
		case 3:
			assert.Equal(t, "C", entry[entries.StandardMessageField], "'C' should have been parsed as the message")
		default:
			t.Error("Should not have consumed 4+ entries")
		}
		if count == 3 {
			err := _tail.Stop()
			if err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSource_Unstructured(t *testing.T) {
	_tail, iter, err := ctxSource(context.Background(), "unstructured.log")
	require.NoError(t, err)
	require.NotNil(t, _tail)
	require.NotNil(t, iter)

	count := 0
	err = iter.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		assert.True(t, entry.HasField(entries.StandardMessageField), "Entry should have '@message' field")
		assert.True(t, entry.HasField("@read_timestamp"), "Entry should have '@read_timestamp' field")
		assert.True(t, entry.HasField("@read_line_number"), "Entry should have '@read_line_number' field")
		switch count {
		case 1:
			assert.Equal(t, "A", entry[entries.StandardMessageField], "'A' should have been parsed as the message")
		case 2:
			assert.Equal(t, "B", entry[entries.StandardMessageField], "'B' should have been parsed as the message")
		case 3:
			assert.Equal(t, "C", entry[entries.StandardMessageField], "'C' should have been parsed as the message")
		default:
			t.Error("Should not have consumed 4+ entries")
		}
		if count == 3 {
			err := _tail.Stop()
			if err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSink(t *testing.T) {
	td, err := os.MkdirTemp("", "TestSink-*")
	require.NoError(t, err)
	t.Log("Using temp directory:", td)
	defer func() {
		err := os.RemoveAll(td)
		if err != nil {
			t.Error("Failed to remove temp directory:", td)
		} else {
			t.Log("Removed temp directory")
		}
	}()

	iter := iterator.FromSlice([]entries.LogEntry{
		{
			"A": "A",
		},
	})
	err = Sink(iter, filepath.Join(td, "test.log"), 0600)
	assert.NoError(t, err)

	f, err := os.Open(filepath.Join(td, "test.log"))
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
	}()
	entry := entries.LogEntry{}
	assert.NoError(t, json.NewDecoder(f).Decode(&entry))
	assert.True(t, entry.HasField("A"), "Log entry wasn't written")
}

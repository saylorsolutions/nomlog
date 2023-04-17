package file

import (
	"github.com/saylorsolutions/slog/pkg/entries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSource_Structured(t *testing.T) {
	_tail, iter, err := source("structured.log")
	require.NoError(t, err)
	require.NotNil(t, _tail)
	require.NotNil(t, iter)

	count := 0
	err = iter.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		assert.True(t, entry.HasField("@message"), "Entry should have '@message' field")
		assert.True(t, entry.HasField("@timestamp"), "Entry should have '@timestamp' field")
		assert.True(t, entry.HasField("@read_timestamp"), "Entry should have '@read_timestamp' field")
		assert.True(t, entry.HasField("@read_line_number"), "Entry should have '@read_line_number' field")
		switch count {
		case 1:
			assert.Equal(t, "A", entry["@message"], "'A' should have been parsed as the message")
		case 2:
			assert.Equal(t, "B", entry["@message"], "'B' should have been parsed as the message")
		case 3:
			assert.Equal(t, "C", entry["@message"], "'C' should have been parsed as the message")
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
	_tail, iter, err := source("unstructured.log")
	require.NoError(t, err)
	require.NotNil(t, _tail)
	require.NotNil(t, iter)

	count := 0
	err = iter.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		assert.True(t, entry.HasField("@message"), "Entry should have '@message' field")
		assert.True(t, entry.HasField("@read_timestamp"), "Entry should have '@read_timestamp' field")
		assert.True(t, entry.HasField("@read_line_number"), "Entry should have '@read_line_number' field")
		switch count {
		case 1:
			assert.Equal(t, "A", entry["@message"], "'A' should have been parsed as the message")
		case 2:
			assert.Equal(t, "B", entry["@message"], "'B' should have been parsed as the message")
		case 3:
			assert.Equal(t, "C", entry["@message"], "'C' should have been parsed as the message")
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

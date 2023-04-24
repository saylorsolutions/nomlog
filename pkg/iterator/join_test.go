package iterator

import (
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoiner(t *testing.T) {
	iter := FromSlice([]entries.LogEntry{
		entries.FromString("start entry"),
		entries.FromString("another entry"),
		entries.FromString("start complete"),
	})
	iter = Joiner(iter, `^start`)

	first, _, err := iter.Next()
	msg, ok := first.AsString(entries.StandardMessageField)
	assert.NoError(t, err)
	assert.True(t, ok, "Message should be defined on first log event")
	assert.Equal(t, "start entry\nanother entry", msg)

	second, _, err := iter.Next()
	msg, ok = second.AsString(entries.StandardMessageField)
	assert.NoError(t, err)
	assert.True(t, ok, "Message should be defined on second log event")
	assert.Equal(t, "start complete", msg)

	_, _, err = iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
}

func TestJoiner_Midstream_read(t *testing.T) {
	iter := FromSlice([]entries.LogEntry{
		entries.FromString("another entry"),
		entries.FromString("start complete"),
	})
	iter = Joiner(iter, `^start`)

	first, _, err := iter.Next()
	msg, ok := first.AsString(entries.StandardMessageField)
	assert.NoError(t, err)
	assert.True(t, ok, "Message should be defined on first log event")
	assert.Equal(t, "another entry", msg)

	second, _, err := iter.Next()
	msg, ok = second.AsString(entries.StandardMessageField)
	assert.NoError(t, err)
	assert.True(t, ok, "Message should be defined on second log event")
	assert.Equal(t, "start complete", msg)

	_, _, err = iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
}

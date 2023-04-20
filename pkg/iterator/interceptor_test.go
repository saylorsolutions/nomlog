package iterator

import (
	"context"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	iter := FromSlice([]entries.LogEntry{
		{
			"A": "A",
		},
		{
			"B": "B",
		},
		{
			"C": "C",
		},
	})
	iter = Filter(iter, func(entry entries.LogEntry, i int, err error) bool {
		return entry.HasField("C")
	})

	el, _, err := iter.Next()
	assert.NoError(t, err)
	s, ok := el.AsString("C")
	assert.True(t, ok, "Field 'C' should exist in this entry")
	assert.Equal(t, "C", s)

	el, _, err = iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
}

func TestCancellable(t *testing.T) {
	iter := FromSlice([]entries.LogEntry{
		{
			"A": "A",
		},
		{
			"B": "B",
		},
		{
			"C": "C",
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	iter = Cancellable(ctx, iter)

	el, _, err := iter.Next()
	assert.NoError(t, err)
	s, ok := el.AsString("A")
	assert.True(t, ok, "Field 'A' should exist in this entry")
	assert.Equal(t, "A", s)

	cancel()
	time.Sleep(200 * time.Millisecond)

	el, _, err = iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
}

func TestConcat(t *testing.T) {
	iter1 := FromSlice([]entries.LogEntry{
		{
			"A": "A",
		},
	})
	iter2 := FromSlice([]entries.LogEntry{
		{
			"B": "B",
		},
	})
	iter := Filter(Concat(iter1, iter2), func(entry entries.LogEntry, i int, err error) bool {
		t.Log(entry)
		return true
	})

	el, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, i)
	s, ok := el.AsString("A")
	assert.True(t, ok, "Field 'A' should exist in this entry")
	assert.Equal(t, "A", s)

	el, i, err = iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	s, ok = el.AsString("B")
	assert.True(t, ok, "Field 'B' should exist in this entry")
	assert.Equal(t, "B", s)

	el, _, err = iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
}

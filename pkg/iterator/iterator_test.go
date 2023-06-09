package iterator

import (
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntrySlice_Next(t *testing.T) {
	iter := FromSlice(_testEntries())
	a, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, "A", a["message"])

	b, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, "B", b["message"])

	c, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 2, i)
	assert.Equal(t, "C", c["message"])

	z, i, err := iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
	assert.Equal(t, -1, i)
	assert.Nil(t, z)
}

func TestEntryChannel_Next(t *testing.T) {
	iter := FromChannel(_testEntryChannel())
	a, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, "A", a["message"])

	b, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, "B", b["message"])

	c, i, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, 2, i)
	assert.Equal(t, "C", c["message"])

	z, i, err := iter.Next()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAtEnd)
	assert.Equal(t, -1, i)
	assert.Nil(t, z)
}

func TestEntrySlice_Iterate(t *testing.T) {
	iter := FromSlice(_testEntries())
	count := 0

	err := iter.Iterate(func(entry entries.LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestEntryChannel_Iterate(t *testing.T) {
	iter := FromChannel(_testEntryChannel())
	count := 0

	err := iter.Iterate(func(entry entries.LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestMerge_FromSlices(t *testing.T) {
	a := FromSlice(_testEntries())
	b := FromSlice(_testEntries())
	c := Merge(a, b)
	count := 0

	err := c.Iterate(func(entry entries.LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 6, count)
}

func TestMerge_FromChannels(t *testing.T) {
	a := FromChannel(_testEntryChannel())
	b := FromChannel(_testEntryChannel())
	c := Merge(a, b)
	count := 0

	err := c.Iterate(func(entry entries.LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 6, count)
}

func TestDupe(t *testing.T) {
	base := FromSlice(_testEntries())
	a, b := Dupe(base)
	merged := Merge(a, b)
	count := 0
	err := merged.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 6, count)
}

func TestFanout(t *testing.T) {
	base := FromSlice(_testEntries())
	a, b := Fanout(base)
	merged := Merge(a, b)
	count := 0
	err := merged.Iterate(func(entry entries.LogEntry, i int) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func _testEntries() []entries.LogEntry {
	return []entries.LogEntry{
		{
			"message": "A",
		},
		{
			"message": "B",
		},
		{
			"message": "C",
		},
	}
}

func _testEntryChannel() <-chan entries.LogEntry {
	slice := _testEntries()
	ch := make(chan entries.LogEntry, len(slice))
	for _, s := range slice {
		ch <- s
	}
	close(ch)
	return ch
}

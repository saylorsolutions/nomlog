package entries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntrySlice_Next(t *testing.T) {
	iter := NewSliceIterator(_testEntries())
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
	assert.ErrorIs(t, err, ErrStopIteration)
	assert.Equal(t, -1, i)
	assert.Nil(t, z)
}

func TestEntryChannel_Next(t *testing.T) {
	iter := NewChannelIterator(_testEntryChannel())
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
	assert.ErrorIs(t, err, ErrStopIteration)
	assert.Equal(t, -1, i)
	assert.Nil(t, z)
}

func TestEntrySlice_Iterate(t *testing.T) {
	iter := NewSliceIterator(_testEntries())
	count := 0

	err := iter.Iterate(func(entry LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestEntryChannel_Iterate(t *testing.T) {
	iter := NewChannelIterator(_testEntryChannel())
	count := 0

	err := iter.Iterate(func(entry LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestMerge_FromSlices(t *testing.T) {
	a := NewSliceIterator(_testEntries())
	b := NewSliceIterator(_testEntries())
	c := Merge(a, b)
	count := 0

	err := c.Iterate(func(entry LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 6, count)
}

func TestMerge_FromChannels(t *testing.T) {
	a := NewChannelIterator(_testEntryChannel())
	b := NewChannelIterator(_testEntryChannel())
	c := Merge(a, b)
	count := 0

	err := c.Iterate(func(entry LogEntry, i int) error {
		count += 1
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 6, count)
}

func _testEntries() []LogEntry {
	return []LogEntry{
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

func _testEntryChannel() <-chan LogEntry {
	slice := _testEntries()
	ch := make(chan LogEntry, len(slice))
	for _, s := range slice {
		ch <- s
	}
	close(ch)
	return ch
}

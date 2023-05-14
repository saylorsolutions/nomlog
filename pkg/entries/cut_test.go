package entries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCut(t *testing.T) {
	entry := LogEntry{
		StandardMessageField: "a b c",
	}
	entry, err := Cut(entry,
		CutDelim(" "),
		CutField(StandardMessageField),
		CutCollector(
			NewCutCollectSpec().
				Map("a", 0).
				Map("b", 1).
				Map("c", 2).
				Collector(),
		),
		RemoveSource(),
	)
	assert.NoError(t, err)

	a, ok := entry.AsString("a")
	assert.True(t, ok, "Should have a new field 'a'")
	assert.Equal(t, "a", a)
	b, ok := entry.AsString("b")
	assert.True(t, ok, "Should have a new field 'b'")
	assert.Equal(t, "b", b)
	c, ok := entry.AsString("c")
	assert.True(t, ok, "Should have a new field 'c'")
	assert.Equal(t, "c", c)
	assert.False(t, entry.HasField(StandardMessageField), "Standard message field should be removed after cutting")
}

func TestCut_DefaultOpts(t *testing.T) {
	entry := LogEntry{
		StandardMessageField: "a b c d",
	}
	entry, err := Cut(entry)
	assert.NoError(t, err)

	a, ok := entry.AsString("0")
	assert.True(t, ok, "Should have a new field '0'")
	assert.Equal(t, "a", a)
	b, ok := entry.AsString("1")
	assert.True(t, ok, "Should have a new field '1'")
	assert.Equal(t, "b", b)
	c, ok := entry.AsString("2")
	assert.True(t, ok, "Should have a new field '2'")
	assert.Equal(t, "c", c)
	msg, ok := entry.AsString(StandardMessageField)
	assert.True(t, ok, "Standard message field should NOT be removed after cutting")
	assert.Equal(t, "", msg)
}

func TestCut_WithRemainder(t *testing.T) {
	entry := LogEntry{
		StandardMessageField: "a b c d e",
	}
	entry, err := Cut(entry,
		CutDelim(" "),
		CutField(StandardMessageField),
		CutCollector(
			NewCutCollectSpec().
				Map("a", 0).
				Map("b", 1).
				Map("c", 2).
				Collector(),
		),
	)
	assert.NoError(t, err)

	a, ok := entry.AsString("a")
	assert.True(t, ok, "Should have a new field 'a'")
	assert.Equal(t, "a", a)
	b, ok := entry.AsString("b")
	assert.True(t, ok, "Should have a new field 'b'")
	assert.Equal(t, "b", b)
	c, ok := entry.AsString("c")
	assert.True(t, ok, "Should have a new field 'c'")
	assert.Equal(t, "c", c)
	msg, ok := entry.AsString(StandardMessageField)
	assert.True(t, ok, "Standard message field should be present")
	assert.Equal(t, "d e", msg, "Message remainder should be present in the field")
}

func TestCut_WithNegativeIndex(t *testing.T) {
	entry := LogEntry{
		StandardMessageField: "a b c d e",
	}
	entry, err := Cut(entry,
		CutDelim(" "),
		CutField(StandardMessageField),
		CutCollector(
			NewCutCollectSpec().
				Map("a", -5).
				Map("b", -4).
				Map("c", -3).
				Map("e", -1).
				Collector(),
		),
	)
	assert.NoError(t, err)

	a, ok := entry.AsString("a")
	assert.True(t, ok, "Should have a new field 'a'")
	assert.Equal(t, "a", a)
	b, ok := entry.AsString("b")
	assert.True(t, ok, "Should have a new field 'b'")
	assert.Equal(t, "b", b)
	c, ok := entry.AsString("c")
	assert.True(t, ok, "Should have a new field 'c'")
	assert.Equal(t, "c", c)
	e, ok := entry.AsString("e")
	assert.True(t, ok, "Should have a new field 'e'")
	assert.Equal(t, "e", e)
	msg, ok := entry.AsString(StandardMessageField)
	assert.True(t, ok, "Standard message field should be present")
	assert.Equal(t, "d", msg, "Message remainder should be present in the field")
}

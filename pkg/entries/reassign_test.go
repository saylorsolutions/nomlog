package entries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReassign(t *testing.T) {
	entry := LogEntry{
		"a": 5,
	}
	spec := NewReassignSpec().
		Move("a", "b").
		Move("doesn't exist", "z")
	entry = Reassign(entry, spec)
	b, ok := entry.AsString("b")
	assert.True(t, ok, "'b' should exist")
	assert.Equal(t, "5", b)
	assert.False(t, entry.HasField("a"), "'a' should have moved to 'b'")
	assert.False(t, entry.HasField("z"), "Should not be a 'z' field")
	assert.False(t, entry.HasField("doesn't exist"), "Should not be a 'doesn't exist' field")
}

package entries

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLogEntry_AsFloat(t *testing.T) {
	tests := map[string]struct {
		val      any
		expected float64
		exists   bool
	}{
		"float64": {
			val:      float64(5),
			expected: 5,
			exists:   true,
		},
		"float32": {
			val:      float32(5),
			expected: 5,
			exists:   true,
		},
		"float string": {
			val:      "5.0",
			expected: 5,
			exists:   true,
		},
		"something else": {
			val:      'a',
			expected: 0,
			exists:   false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			entry := LogEntry{
				"val": tc.val,
			}
			got, ok := entry.AsFloat("val")
			t.Log("Expected:", tc.expected)
			assert.Equal(t, tc.expected, got)
			assert.Equal(t, tc.exists, ok)
		})
	}
}

func TestLogEntry_AsInt(t *testing.T) {
	tests := map[string]struct {
		val      any
		expected int64
		exists   bool
	}{
		"int": {
			val:      5,
			expected: 5,
			exists:   true,
		},
		"int64": {
			val:      int64(5),
			expected: 5,
			exists:   true,
		},
		"int32": {
			val:      int32(5),
			expected: 5,
			exists:   true,
		},
		"int string": {
			val:      "5",
			expected: 5,
			exists:   true,
		},
		"something else": {
			val:      "blah",
			expected: 0,
			exists:   false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			entry := LogEntry{
				"val": tc.val,
			}
			got, ok := entry.AsInt("val")
			t.Log("Expected:", tc.expected)
			assert.Equal(t, tc.expected, got)
			assert.Equal(t, tc.exists, ok)
		})
	}
}

func TestLogEntry_AsUint(t *testing.T) {
	tests := map[string]struct {
		val      any
		expected uint64
		exists   bool
	}{
		"uint64": {
			val:      uint64(5),
			expected: 5,
			exists:   true,
		},
		"uint32": {
			val:      uint32(5),
			expected: 5,
			exists:   true,
		},
		"int string": {
			val:      "5",
			expected: 5,
			exists:   true,
		},
		"something else": {
			val:      "blah",
			expected: 0,
			exists:   false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			entry := LogEntry{
				"val": tc.val,
			}
			got, ok := entry.AsUint("val")
			t.Log("Expected:", tc.expected)
			assert.Equal(t, tc.expected, got)
			assert.Equal(t, tc.exists, ok)
		})
	}
}

func TestLogEntry_AsTime(t *testing.T) {
	var none time.Time
	now, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	now822, err := time.Parse(time.RFC822, time.Now().UTC().Format(time.RFC822))
	assert.NoError(t, err)
	tests := map[string]struct {
		val      any
		expected time.Time
		exists   bool
	}{
		"Time": {
			val:      now,
			expected: now,
			exists:   true,
		},
		"Time string RFC 3339": {
			val:      now.Format(time.RFC3339),
			expected: now,
			exists:   true,
		},
		"Time string RFC 822": {
			val:      now822.Format(time.RFC822),
			expected: now822,
			exists:   true,
		},
		"something else": {
			val:      "blah",
			expected: none,
			exists:   false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			entry := LogEntry{
				"val": tc.val,
			}
			got, ok := entry.AsTime("val", time.RFC3339, time.RFC822)
			t.Log("Expected:", tc.expected)
			assert.Equal(t, tc.expected, got)
			assert.Equal(t, tc.exists, ok)
		})
	}
}

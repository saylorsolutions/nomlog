package entries

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestTransform(t *testing.T) {
	tests := map[string]struct {
		value      any
		expected   any
		transforms []transFunc
	}{
		"No transformation": {
			value:    "5",
			expected: "5",
		},
		"trim transform": {
			value:    " \t\nwith spaces \t\n",
			expected: "with spaces",
			transforms: []transFunc{
				TransformType(func(s string) string {
					return strings.TrimSpace(s)
				}),
			},
		},
		"Nil transform": {
			value:      "5",
			expected:   "5",
			transforms: []transFunc{nil},
		},
		"String from number": {
			value:    5,
			expected: 5,
			transforms: []transFunc{
				TransformType(func(s string) string {
					t.Error("String transform shouldn't run on non-string value")
					return s
				}),
			},
		},
		"Number transform": {
			value:    5,
			expected: 10,
			transforms: []transFunc{
				func(val any) any {
					if i, ok := val.(int); ok {
						return i * 2
					}
					return val
				},
			},
		},
		"Nil value": {
			value:    nil,
			expected: nil,
			transforms: []transFunc{
				func(val any) any {
					t.Error("Transforms should not be run on nil values")
					return val
				},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			entry := LogEntry{
				"value": tc.value,
			}
			spec := NewTransformSpec()
			spec.Transform("not-found", func(val any) any {
				t.Error("Unrelated transform ran")
				return val
			})

			for _, tf := range tc.transforms {
				spec.Transform("value", tf)
			}
			t.Logf("Initial value: '%v' with type %[1]T", entry["value"])
			result := Transform(entry, spec)
			t.Logf("Transformed value: '%v' with type %[1]T", entry["value"])
			assert.Equal(t, tc.expected, result["value"])
		})
	}
}

package dsl

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseString_ShortString(t *testing.T) {
	text := `source as someFile file.File "someFile"`
	nodes, err := ParseString(text)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	expected := &Source{
		ID: "someFile",
		Class: &SourceClass{
			Qualifier:   "file",
			SourceClass: "File",
		},
		Args: []*Arg{{String: "someFile"}},
	}
	source, ok := nodes[0].(*Source)
	assert.True(t, ok, "Expected to parse Source AstNode")
	assert.Equal(t, expected.ID, source.ID)
	assert.Equal(t, expected.Class.Qualifier, source.Class.Qualifier)
	assert.Equal(t, expected.Class.SourceClass, source.Class.SourceClass)
	assert.Len(t, source.Args, len(expected.Args))
	assert.Equal(t, expected.Args[0].String, source.Args[0].String)
}

func TestEscapeString(t *testing.T) {
	given := `" \t\r\n\abc\\n\r\t "`
	expected := " \t\r\n\\abc\\\n\r\t "
	assert.Equal(t, expected, escapeString(given))
}

func TestParseString(t *testing.T) {
	nodes, err := ParseString(string(script))
	assert.NoError(t, err)
	expectedTypes := []AstType{SOURCE, SOURCE, MERGE, DUPE, APPEND, CUT, FANOUT, TAG, JOIN, SINK, ASYNC_SINK}
	assert.Len(t, nodes, len(expectedTypes))

	for i := 0; i < len(expectedTypes); i++ {
		if len(nodes) > i+1 {
			assert.Equal(t, expectedTypes[i], nodes[i].Type())
		}
	}
}

func TestAst_MarshalJSON(t *testing.T) {
	nodes, err := ParseString(string(script))
	assert.NoError(t, err)
	data, err := json.MarshalIndent(nodes, "", "  ")
	assert.NoError(t, err)
	t.Log(string(data))
}

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
	expectedTypes := []AstType{SOURCE, SOURCE, MERGE, DUPE, APPEND, CUT, FANOUT, TAG, EOL, JOIN, SINK, ASYNC_SINK}
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

func TestParseString_Multiline(t *testing.T) {
	script := `source as src file.File "a really long string
that is even multi-line"`
	nodes, err := ParseString(script)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, SOURCE, nodes[0].Type())
	src, ok := nodes[0].(*Source)
	assert.True(t, ok, "SOURCE node should be *Source")
	assert.Len(t, src.Args, 1)
	assert.Equal(t, "\"a really long string\nthat is even multi-line\"", src.Args[0].Text())
}

func TestParse_ClassNoArgs(t *testing.T) {
	script := `source as a std.In
sink a to std.Out`
	nodes, err := ParseString(script)
	assert.NoError(t, err)

	expected := []AstType{SOURCE, SINK}
	assert.Len(t, nodes, 2)
	for i, n := range nodes {
		assert.NotNil(t, n)
		assert.Equal(t, expected[i], n.Type())
		switch n := n.(type) {
		case *Source:
			assert.Len(t, n.Args, 0)
		case *Sink:
			assert.Len(t, n.Args, 0)
		default:
			t.Errorf("Unexpected AstNode type '%T'", n)
		}
	}
}

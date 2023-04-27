package dsl

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestLexer_Lex(t *testing.T) {
	text := `source as blah file.Tail "file.log"
`
	l := lexString(text)
	go func() {
		l.lex()
	}()
	tokens := consume(l.tokens)
	for i, tok := range tokens {
		if tok.Type == tErr {
			t.Errorf("Error at position %d: '%v'", tok.Pos, tok.Text)
		}
		switch i {
		case 0:
			assert.Equal(t, tSource, tok.Type)
		case 1:
			assert.Equal(t, tAs, tok.Type)
		case 2:
			assert.Equal(t, tIdentifier, tok.Type)
		case 3:
			assert.Equal(t, tIdentifier, tok.Type)
		case 4:
			assert.Equal(t, tDot, tok.Type)
		case 5:
			assert.Equal(t, tIdentifier, tok.Type)
		case 6:
			assert.Equal(t, tString, tok.Type)
		case 7:
			assert.Equal(t, tEol, tok.Type)
		case 8:
			assert.Equal(t, tEof, tok.Type)
		default:
			t.Error("Should not have got this many tokens")
		}
	}
}

//go:embed test
var script []byte

func TestLexFile(t *testing.T) {
	tmp, err := os.MkdirTemp("", "TestLexFile-*")
	require.NoError(t, err)

	defer func() {
		err := os.RemoveAll(tmp)
		assert.NoError(t, err, "Failed to remove temp dir", tmp)
	}()

	path := filepath.Join(tmp, "script")
	require.NoError(t, os.WriteFile(path, script, 0600))

	l, err := lexFile(path)
	assert.NoError(t, err)
	go l.lex()
	tokens := consume(l.tokens)

	expected := []lexType{
		tSource, tAs, tIdentifier, tIdentifier, tDot, tIdentifier, tString, tEol,
		tSource, tAs, tIdentifier, tIdentifier, tDot, tIdentifier, tString, tEol,
		tMerge, tIdentifier, tAnd, tIdentifier, tAs, tIdentifier, tEol,
		tDupe, tIdentifier, tAs, tIdentifier, tAnd, tIdentifier, tEol,
		tAppend, tIdentifier, tTo, tIdentifier, tEol,
		tCut, tWith, tString, tIdentifier, tSet, tLpar, tIdentifier, tEq, tInt, tComma, tIdentifier, tEq, tInt, tRpar, tEol,
		tFanout, tIdentifier, tAs, tIdentifier, tAnd, tIdentifier, tEol,
		tSink, tIdentifier, tTo, tIdentifier, tDot, tIdentifier, tString, tEol,
		tSink, tIdentifier, tAsync, tAs, tIdentifier, tTo, tIdentifier, tDot, tIdentifier, tString, tEol,
		tEof,
	}

	for i, tok := range tokens {
		require.Equalf(t, expected[i], tok.Type, "token %d mismatch: %s", i, tok.Text)
	}
	assert.Len(t, tokens, len(expected))
}

func consume(ch <-chan token) []token {
	var tokens []token
	for t := range ch {
		tokens = append(tokens, t)
	}
	return tokens
}

package dsl

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
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

func TestLexer_ReadNumber(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
		found    bool
		typ      lexType
	}{
		"Integer": {
			input:    "123",
			expected: "123",
			found:    true,
			typ:      tInt,
		},
		"Negative Integer": {
			input:    "-123",
			expected: "-123",
			found:    true,
			typ:      tInt,
		},
		"Number": {
			input:    "123.01",
			expected: "123.01",
			found:    true,
			typ:      tNumber,
		},
		"Negative Number": {
			input:    "-123.01",
			expected: "-123.01",
			found:    true,
			typ:      tNumber,
		},
		"Integer with suffix": {
			input:    "123abc",
			expected: "123",
			found:    true,
			typ:      tInt,
		},
		"Negative Integer with suffix": {
			input:    "-123abc",
			expected: "-123",
			found:    true,
			typ:      tInt,
		},
		"Number with suffix": {
			input:    "123.01abc",
			expected: "123.01",
			found:    true,
			typ:      tNumber,
		},
		"Negative Number with suffix": {
			input:    "-123.01abc",
			expected: "-123.01",
			found:    true,
			typ:      tNumber,
		},
	}

	for name, tc := range tests {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			l := lexString(tc.input)
			var (
				err    error
				wg     sync.WaitGroup
				tokens []token
			)
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer close(l.tokens)
				err = l.readNumber()
			}()
			go func() {
				defer wg.Done()
				tokens = consume(l.tokens)
			}()
			wg.Wait()
			assert.Equal(t, tc.found, err == nil)
			if err == nil {
				assert.Len(t, tokens, 1)
				if len(tokens) > 0 {
					assert.Equal(t, tc.expected, tokens[0].Text)
					assert.Equal(t, tc.typ, tokens[0].Type)
				}
			}
		})
	}
}

func consume(ch <-chan token) []token {
	var tokens []token
	for t := range ch {
		tokens = append(tokens, t)
	}
	return tokens
}

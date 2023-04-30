package dsl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type lexType int

const (
	tEof lexType = iota + 1
	tErr
	tEol
	tString
	tNumber
	tLpar
	tRpar
	tInt
	tEq
	tComma
	tDot
	tAs
	tAnd
	tTo
	tSource
	tVar
	tSink
	tAsync
	tIdentifier
	tMerge
	tDupe
	tAppend
	tCut
	tSet
	tWith
	tFanout
)

var (
	digitRegex             = regexp.MustCompile(`^\d$`)
	idRegex                = regexp.MustCompile(`^\w(\w|\d)*$`)
	ErrNoDigitAfterDecimal = errors.New("missing digit(s) after decimal")
	ErrUnknownToken        = errors.New("unknown token")
)

type token struct {
	Pos  int
	Line int
	Text string
	Type lexType
}

type lexFn func(int, *lexer) lexFn

type lexer struct {
	*lexBuf
	tokens chan token
	err    error
}

func lexFile(filename string) (*lexer, error) {
	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	l := lexString(string(text))
	return l, nil
}

func lexString(text string) *lexer {
	r := strings.NewReader(text)
	return lexReader(r)
}

func lexReader(r io.Reader) *lexer {
	rr := bufio.NewReader(r)
	return &lexer{tokens: make(chan token), lexBuf: newLexBuf(rr)}
}

func (l *lexer) postToken(t lexType) {
	text := l.consume()
	l.tokens <- token{l.pos - len(text), l.line, text, t}
}

func (l *lexer) handleLexErr(err error) {
	l.err = err
	if err == io.EOF {
		l.tokens <- token{Pos: l.pos, Text: "", Type: tEof}
		return
	}
	l.tokens <- token{Pos: l.pos, Text: err.Error(), Type: tErr}
}

func (l *lexer) isDigit(r rune) bool {
	return digitRegex.MatchString(string(r))
}

func (l *lexer) lex() {
	defer close(l.tokens)
	for {
		if l.err != nil {
			return
		}
		if err := l.skipWhitespace(); err != nil {
			l.handleLexErr(err)
			return
		}
		c, err := l.read()
		if err != nil {
			l.handleLexErr(err)
			return
		}
		switch {
		case c == '\n':
			l.postToken(tEol)
			l.line++
			l.pos = 1
		case c == '"':
			if err := l.readString(); err != nil {
				l.handleLexErr(err)
				return
			}
		case l.isDigit(c):
			if err := l.readInt(); err != nil {
				l.handleLexErr(err)
				return
			}
		case c == '(':
			l.postToken(tLpar)
		case c == ')':
			l.postToken(tRpar)
		case c == '=':
			l.postToken(tEq)
		case c == ',':
			l.postToken(tComma)
		case c == '.':
			l.postToken(tDot)
		default:
			if err := l.readKeywords(); err != nil {
				l.handleLexErr(err)
				return
			}
		}
	}
}

func (l *lexer) readString() error {
	for {
		c, err := l.read()
		if err != nil {
			return err
		}
		switch {
		case c == '\\':
			_, err := l.read()
			if err != nil {
				return err
			}
		case c == '"':
			l.postToken(tString)
			return nil
		}
	}
}

func (l *lexer) readInt() error {
	for {
		c, err := l.read()
		if err != nil {
			return err
		}
		switch {
		case c == '.':
			return l.readDecimal()
		case l.isDigit(c):
			continue
		default:
			l.unread()
			l.postToken(tInt)
			return nil
		}
	}
}

func (l *lexer) readDecimal() error {
	var hasRead bool
	for {
		c, err := l.read()
		if err != nil {
			return err
		}
		if l.isDigit(c) {
			hasRead = true
			continue
		}
		if !hasRead {
			return ErrNoDigitAfterDecimal
		}
		l.unread()
		l.postToken(tNumber)
		return nil
	}
}

func (l *lexer) readKeywords() error {
	if err := l.readUntilWhitespaceOrBreak(); err != nil {
		return err
	}
	s := l.preview()

	switch s {
	case "as":
		l.postToken(tAs)
	case "and":
		l.postToken(tAnd)
	case "to":
		l.postToken(tTo)
	case "source":
		l.postToken(tSource)
	case "var":
		l.postToken(tVar)
	case "sink":
		l.postToken(tSink)
	case "async":
		l.postToken(tAsync)
	case "merge":
		l.postToken(tMerge)
	case "dupe":
		l.postToken(tDupe)
	case "append":
		l.postToken(tAppend)
	case "cut":
		l.postToken(tCut)
	case "set":
		l.postToken(tSet)
	case "with":
		l.postToken(tWith)
	case "fanout":
		l.postToken(tFanout)
	default:
		if idRegex.MatchString(s) {
			l.postToken(tIdentifier)
			return nil
		}
		return fmt.Errorf("%w: %s", ErrUnknownToken, s)
	}
	return nil
}

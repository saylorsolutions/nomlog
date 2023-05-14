package dsl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

const (
	alpha       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	number      = "0123456789"
	idStart     = alpha
	idRemainder = alpha + number + "_"
)

var (
	ErrNoDigitAfterDecimal = errors.New("missing digit(s) after decimal")
	ErrUnknownToken        = errors.New("unknown token")
	ErrInvalidNumber       = errors.New("invalid number")
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
		l.tokens <- token{Pos: l.pos, Line: l.line, Text: "", Type: tEof}
		return
	}
	text := l.consume()
	text = ": '" + text + "'"
	l.tokens <- token{Pos: l.pos, Line: l.line, Text: err.Error() + text, Type: tErr}
}

func (l *lexer) isDigit(r rune) bool {
	digits := "0123456789"
	for _, d := range []rune(digits) {
		if r == d {
			return true
		}
	}
	return false
}

func (l *lexer) stream() *tokenStream {
	return newTokenStream(l.tokens)
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
		case c == '-' || l.isDigit(c):
			l.unread()
			if err := l.readNumber(); err != nil {
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
		var lineAdd int
		c, err := l.read()
		if err != nil {
			l.line += lineAdd
			if err == io.EOF {
				return fmt.Errorf("%w: unterminated string", err)
			}
			return err
		}
		switch {
		case c == '\\':
			_, err := l.read()
			if err != nil {
				l.line += lineAdd
				return err
			}
		case c == '\n':
			lineAdd++
		case c == '"':
			l.postToken(tString)
			l.line += lineAdd
			return nil
		}
	}
}

func (l *lexer) readNumber() error {
	l.acceptOne("-")
	if !l.accept(number) {
		return fmt.Errorf("%w: no digit found", ErrInvalidNumber)
	}
	if !l.acceptOne(".") {
		l.postToken(tInt)
		return nil
	}
	if !l.accept(number) {
		return ErrNoDigitAfterDecimal
	}
	l.postToken(tNumber)
	return nil
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
		l.reset()
		if !l.readIdentifier() {
			return fmt.Errorf("%w: %s", ErrUnknownToken, s)
		}
	}
	return nil
}

func (l *lexer) readIdentifier() bool {
	if !l.acceptOne(idStart) {
		return false
	}
	l.accept(idRemainder)
	l.postToken(tIdentifier)
	return true
}

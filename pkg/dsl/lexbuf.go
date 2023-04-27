package dsl

import (
	"io"
	"unicode"
)

var (
	lexBufferSize = 256
)

type lexBuf struct {
	startPtr int
	readPtr  int
	writePtr int
	buf      []rune
	r        io.RuneReader
	pos      int
}

func newLexBuf(reader io.RuneReader) *lexBuf {
	return newLexBufSize(reader, lexBufferSize)
}

func newLexBufSize(reader io.RuneReader, size int) *lexBuf {
	return &lexBuf{
		r:   reader,
		buf: make([]rune, size),
	}
}

func (b *lexBuf) decrement(i int) int {
	return (i - 1 + len(b.buf)) % len(b.buf)
}

func (b *lexBuf) increment(i int) int {
	return (i + 1) % len(b.buf)
}

func (b *lexBuf) readerRead() error {
	r, _, err := b.r.ReadRune()
	if err != nil {
		return err
	}
	b.buf[b.writePtr] = r
	b.writePtr = b.increment(b.writePtr)

	if b.startPtr == b.writePtr {
		b.startPtr = b.increment(b.startPtr)
	}
	return nil
}

func (b *lexBuf) read() (rune, error) {
	c, err := b.peek()
	if err != nil {
		return 0, err
	}
	b.readPtr = b.increment(b.readPtr)
	b.pos++
	return c, nil
}

func (b *lexBuf) peek() (rune, error) {
	if b.readPtr == b.writePtr {
		if err := b.readerRead(); err != nil && err != io.EOF {
			return 0, err
		}
	}
	if b.readPtr == b.writePtr {
		return 0, io.EOF
	}
	return b.buf[b.readPtr], nil
}

func (b *lexBuf) unread() {
	if b.readPtr == b.startPtr {
		return
	}
	b.readPtr = b.decrement(b.readPtr)
	b.pos--
}

func (b *lexBuf) reset() {
	b.readPtr = b.startPtr
}

func (b *lexBuf) discard() {
	b.startPtr = b.readPtr
}

func (b *lexBuf) consume() string {
	s := b.preview()
	b.discard()
	return s
}

func (b *lexBuf) preview() string {
	if b.startPtr == b.readPtr {
		return ""
	}

	if b.startPtr > b.readPtr {
		return string(append(b.buf[b.startPtr:], b.buf[0:b.readPtr]...))
	}
	return string(b.buf[b.startPtr:b.readPtr])
}

func (b *lexBuf) skipWhitespace() error {
	defer b.discard()
	for {
		c, err := b.read()
		if err != nil {
			return err
		}
		if c == '\n' || !unicode.IsSpace(c) {
			b.unread()
			return nil
		}
	}
}

func (b *lexBuf) readUntilWhitespaceOrBreak() error {
	for {
		c, err := b.read()
		if err != nil {
			return err
		}
		switch {
		case unicode.IsSpace(c):
			fallthrough
		case c == '(':
			fallthrough
		case c == '.':
			fallthrough
		case c == ')':
			b.unread()
			return nil
		}
	}
}

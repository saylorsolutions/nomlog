package dsl

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLexBuf_Read(t *testing.T) {
	slc := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	rr := bufio.NewReader(strings.NewReader(string(slc)))
	buf := newLexBufSize(rr, 5)

	for i := 0; i < 4; i++ {
		r, err := buf.read()
		assert.NoError(t, err)
		assert.Equal(t, slc[i], r)
	}
	assert.Equal(t, 4, buf.readPtr)
	assert.Equal(t, 4, buf.writePtr)
	assert.Equal(t, 0, buf.startPtr)

	r, err := buf.read()
	assert.NoError(t, err)
	assert.Equal(t, slc[4], r)

	assert.Equal(t, 0, buf.readPtr)
	assert.Equal(t, 0, buf.writePtr)
	assert.Equal(t, 1, buf.startPtr)

	assert.Equal(t, "bcde", buf.consume())
	assert.Equal(t, 0, buf.startPtr)

	r, err = buf.read()
	assert.NoError(t, err)
	assert.Equal(t, slc[5], r)

	assert.Equal(t, 1, buf.readPtr)
	assert.Equal(t, 1, buf.writePtr)
	assert.Equal(t, 0, buf.startPtr)

	assert.Equal(t, "f", buf.consume())
	assert.Equal(t, 1, buf.startPtr)
}

func TestLexBuf_Unread(t *testing.T) {
	slc := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	rr := bufio.NewReader(strings.NewReader(string(slc)))
	buf := newLexBufSize(rr, 5)

	assert.Equal(t, "", buf.consume())

	for i := 0; i < 5; i++ {
		r, err := buf.read()
		assert.NoError(t, err)
		assert.Equal(t, slc[i], r)
	}
	assert.Equal(t, 0, buf.readPtr)
	assert.Equal(t, 0, buf.writePtr)
	assert.Equal(t, 1, buf.startPtr)

	buf.unread()
	assert.Equal(t, 4, buf.readPtr)

	for i := 0; i < 3; i++ {
		buf.unread()
	}
	assert.Equal(t, 1, buf.readPtr)

	buf.unread()
	assert.Equal(t, 1, buf.readPtr)
}

func TestLexBuf_Accept(t *testing.T) {
	slc := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	rr := bufio.NewReader(strings.NewReader(string(slc)))
	buf := newLexBuf(rr)

	accepted := buf.accept("abd")
	assert.True(t, accepted, "Accept should have found a match")
	assert.Equal(t, "ab", buf.consume())
}

func TestLexBuf_AcceptOne(t *testing.T) {
	slc := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	rr := bufio.NewReader(strings.NewReader(string(slc)))
	buf := newLexBuf(rr)

	accepted := buf.acceptOne("cab")
	assert.True(t, accepted, "Accept should have found a match")
	assert.Equal(t, "a", buf.consume())
}

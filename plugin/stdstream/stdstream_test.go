package stdstream

import (
	"context"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func ExampleSinkOut() {
	iter := iterator.FromSlice([]entries.LogEntry{
		{"a": "a"},
		{"b": "b"},
		{"c": "c"},
	})
	if err := SinkOut(context.Background(), iter); err != nil {
		panic(err)
	}
	// Output:
	// {"a":"a"}
	// {"b":"b"}
	// {"c":"c"}
}

func TestSinkErr(t *testing.T) {
	iter := iterator.FromSlice([]entries.LogEntry{
		{"a": "a"},
		{"b": "b"},
		{"c": "c"},
	})
	str, err := redirectErr(func() error {
		return SinkErr(context.Background(), iter)
	})
	assert.NoError(t, err)
	expected := `{"a":"a"}
{"b":"b"}
{"c":"c"}
`
	assert.Equal(t, expected, str)
}

func TestSourceIn(t *testing.T) {
	expected := `{"a":"a"}
{"b":"b"}
{"c":"c"}
`
	var iter iterator.Iterator
	err, cleanup := redirectIn(expected, func() error {
		_iter, err := SourceIn(context.Background())
		iter = _iter
		return err
	})
	defer cleanup()
	assert.NoError(t, err)
	str, err := redirectErr(func() error {
		return SinkErr(context.Background(), iter)
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, str)
}

func redirectErr(fn func() error) (string, error) {
	var (
		oldErr = os.Stderr
		output strings.Builder
	)
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stderr = w
	defer func() {
		os.Stderr = oldErr
	}()
	err = fn()
	_ = w.Close()
	_, cperr := io.Copy(&output, r)
	if cperr != nil {
		err = cperr
	}
	return output.String(), err
}

func redirectIn(data string, fn func() error) (error, func()) {
	var (
		oldIn   = os.Stdin
		cleanup = func() {}
	)
	r, w, err := os.Pipe()
	if err != nil {
		return err, cleanup
	}
	os.Stdin = r
	cleanup = func() {
		os.Stdin = oldIn
	}
	_, err = io.Copy(w, strings.NewReader(data))
	_ = w.Close()
	if err != nil {
		return err, cleanup
	}
	if err := fn(); err != nil {
		return err, cleanup
	}
	return nil, cleanup
}

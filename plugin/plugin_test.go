package plugin

import (
	"context"
	"fmt"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistration_AllDocs(t *testing.T) {
	reg := NewRegistration()
	newTestPlugin(t).Register(reg)

	expectedDocs := `Sources:
  test.Empty

  test.Source

  Returns test data.

Sinks:
  test.Sink

  Prints the elements of the given iterator.

`
	docs := reg.AllDocs()
	fmt.Println(docs)
	assert.Equal(t, expectedDocs, docs)
}

var _ Plugin = (*testPlugin)(nil)

type testPlugin struct {
	t *testing.T
}

func newTestPlugin(t *testing.T) Plugin {
	return &testPlugin{t: t}
}

func (t *testPlugin) ID() string {
	return "test"
}

func (t *testPlugin) Register(reg *Registration) {
	reg.RegisterSource("test", "Empty", func(ctx context.Context, args ...*dsl.Arg) (iterator.Iterator, error) {
		return iterator.Empty(), nil
	})
	reg.RegisterSource("test", "Source", func(ctx context.Context, args ...*dsl.Arg) (iterator.Iterator, error) {
		return iterator.FromSlice([]entries.LogEntry{
			{
				"value": "a",
			},
			{
				"value": "b",
			},
			{
				"value": "c",
			},
		}), nil
	})
	reg.DocumentSource("test", "Source", `test.Source

Returns test data.`)
	reg.RegisterSink("test", "Sink", func(ctx context.Context, src iterator.Iterator, args ...*dsl.Arg) error {
		return src.Iterate(func(entry entries.LogEntry, i int) error {
			t.t.Log("Entry", i)
			t.t.Log(entry)
			return nil
		})
	})
	reg.DocumentSink("test", "Sink", `test.Sink

Prints the elements of the given iterator.`)
}

func (t *testPlugin) Stopping() error {
	return nil
}

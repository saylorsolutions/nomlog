package plugin

import (
	"context"
	"errors"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
)

var (
	ErrArgs = errors.New("argument error")
)

// Plugin represents the operations expected of a source/sink plugin.
type Plugin interface {
	// Register is called to allow registration of source and sink functions.
	Register(*Registration)
	// Closing is called after all source and sink operations, when the nomlog session is shutting down.
	Closing() error
}

// SourceFunc is a function that takes 0 or more dsl.Arg to produce an iterator.Iterator.
type SourceFunc = func(ctx context.Context, args ...dsl.Arg) (iterator.Iterator, error)

// SinkFunc is a function that consumes an iterator.Iterator and 0 or more dsl.Arg.
type SinkFunc = func(ctx context.Context, src iterator.Iterator, args ...dsl.Arg) error

// Registration is a collection of SourceFunc and SinkFunc to be used by other components.
type Registration struct {
	sources map[string]map[string]SourceFunc
	sinks   map[string]map[string]SinkFunc
}

func (r *Registration) RegisterSource(qualifier, class string, src SourceFunc) {
	if src == nil {
		panic("source is nil")
	}
	sourceMap, ok := r.sources[qualifier]
	if !ok {
		sourceMap = map[string]SourceFunc{}
		r.sources[qualifier] = sourceMap
	}
	sourceMap[class] = src
}

func (r *Registration) Source(qualifier, class string) (SourceFunc, bool) {
	sources, ok := r.sources[qualifier]
	if !ok {
		return nil, false
	}
	source, ok := sources[class]
	return source, ok
}

func (r *Registration) RegisterSink(qualifier, class string, sink SinkFunc) {
	if sink == nil {
		panic("sink is nil")
	}
	sinkMap, ok := r.sinks[qualifier]
	if !ok {
		sinkMap = map[string]SinkFunc{}
		r.sinks[qualifier] = sinkMap
	}
	sinkMap[class] = sink
}

func (r *Registration) Sink(qualifier, class string) (SinkFunc, bool) {
	sinks, ok := r.sinks[qualifier]
	if !ok {
		return nil, false
	}
	sink, ok := sinks[class]
	return sink, ok
}

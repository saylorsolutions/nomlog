package plugin

import (
	"context"
	"errors"
	"fmt"
	"github.com/saylorsolutions/nomlog/pkg/dsl"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"sort"
	"strings"
)

var (
	ErrArgs = errors.New("argument error")
)

// Plugin represents the operations expected of a source/sink plugin.
type Plugin interface {
	// ID should return a unique identifier for this plugin.
	ID() string
	// Register is called to allow registration of source and sink functions.
	Register(*Registration)
	// Stopping is called after all source and sink operations, when the nomlog session is shutting down.
	Stopping() error
}

// SourceFunc is a function that takes 0 or more dsl.Arg to produce an iterator.Iterator.
type SourceFunc = func(ctx context.Context, args ...*dsl.Arg) (iterator.Iterator, error)

// SinkFunc is a function that consumes an iterator.Iterator and 0 or more dsl.Arg.
type SinkFunc = func(ctx context.Context, src iterator.Iterator, args ...*dsl.Arg) error

// Registration is a collection of SourceFunc and SinkFunc to be used by other components.
type Registration struct {
	sources    map[string]map[string]SourceFunc
	sourcesDoc map[string]map[string]string
	sinks      map[string]map[string]SinkFunc
	sinksDoc   map[string]map[string]string
}

func NewRegistration() *Registration {
	return &Registration{
		sources:    map[string]map[string]SourceFunc{},
		sourcesDoc: map[string]map[string]string{},
		sinks:      map[string]map[string]SinkFunc{},
		sinksDoc:   map[string]map[string]string{},
	}
}

// RegisterSource is called by Plugin.Register to provide a source for use in DSL scripts.
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

// DocumentSource is used to document a provided plugin source. It's recommended to provide usage information in this documentation.
func (r *Registration) DocumentSource(qualifier, class, doc string) {
	sourceMap, ok := r.sourcesDoc[qualifier]
	if !ok {
		sourceMap = map[string]string{}
		r.sourcesDoc[qualifier] = sourceMap
	}
	sourceMap[class] = doc
}

// Source retrieves a source known to this Registration.
// It returns the SourceFunc if it exists, documentation, and a bool indicating whether the qualifier and class pair matches a known source.
func (r *Registration) Source(qualifier, class string) (SourceFunc, string, bool) {
	sources, ok := r.sources[qualifier]
	if !ok {
		return nil, "", false
	}
	source, ok := sources[class]
	if !ok {
		return nil, "", false
	}
	defaultDoc := fmt.Sprintf("%s.%s", qualifier, class)
	sourceDoc, ok := r.sourcesDoc[qualifier]
	if !ok {
		return source, defaultDoc, true
	}
	doc, ok := sourceDoc[class]
	if !ok {
		return source, defaultDoc, true
	}
	return source, doc, true
}

// RegisterSink is called by Plugin.Register to provide a sink for use in DSL scripts.
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

// DocumentSink is used to document a provided plugin sink. It's recommended to provide usage information in this documentation.
func (r *Registration) DocumentSink(qualifier, class, doc string) {
	sinkMap, ok := r.sinksDoc[qualifier]
	if !ok {
		sinkMap = map[string]string{}
		r.sinksDoc[qualifier] = sinkMap
	}
	sinkMap[class] = doc
}

// Sink retrieves a sink known to this Registration.
// It returns the SinkFunc if it exists, documentation, and a bool indicating whether the qualifier and class pair matches a known sink.
func (r *Registration) Sink(qualifier, class string) (SinkFunc, string, bool) {
	sinks, ok := r.sinks[qualifier]
	if !ok {
		return nil, "", false
	}
	sink, ok := sinks[class]
	if !ok {
		return nil, "", false
	}
	defaultDoc := fmt.Sprintf("%s.%s", qualifier, class)
	sinksDoc, ok := r.sinksDoc[qualifier]
	if !ok {
		return sink, defaultDoc, true
	}
	doc, ok := sinksDoc[class]
	if !ok {
		return sink, defaultDoc, true
	}
	return sink, doc, true
}

// AllDocs will return a string containing all the documentation for all loaded plugins.
// The listing will include sources, then sinks, in alphabetical order by qualifier and class.
func (r *Registration) AllDocs() string {
	var buf strings.Builder
	buf.WriteString("Sources:\n")
	populateDocs(&buf, r.sources, r.sourcesDoc)
	buf.WriteString("Sinks:\n")
	populateDocs(&buf, r.sinks, r.sinksDoc)
	return buf.String()
}

func getDocs(docs map[string]map[string]string, qualifier, class string) string {
	defaultDoc := fmt.Sprintf("%s.%s", qualifier, class)
	qualDocs, ok := docs[qualifier]
	if !ok {
		return defaultDoc
	}
	doc, ok := qualDocs[class]
	if !ok {
		return defaultDoc
	}
	return doc
}

const (
	indent = "  "
)

func indentString(s string) string {
	s = strings.TrimSuffix(strings.ReplaceAll(indent+s, "\n", "\n"+indent), indent)
	return strings.ReplaceAll(s, "\n"+indent+"\n", "\n\n")
}

func populateDocs[T any](buf *strings.Builder, model map[string]map[string]T, docs map[string]map[string]string) {
	var (
		_buf       strings.Builder
		qualifiers []string
		qualMap    = map[string][]string{}
	)
	for qual, classMap := range model {
		qualifiers = append(qualifiers, qual)
		var classes []string
		for class := range classMap {
			classes = append(classes, class)
		}
		sort.Strings(classes)
		qualMap[qual] = classes
	}
	if len(qualifiers) == 0 {
		_buf.WriteString("None\n")
	} else {
		sort.Strings(qualifiers)
		for _, qual := range qualifiers {
			for _, class := range qualMap[qual] {
				doc := getDocs(docs, qual, class)
				if !strings.HasSuffix(doc, "\n") {
					doc += "\n"
				}
				_buf.WriteString(doc)
				_buf.WriteString("\n")
			}
		}
	}
	buf.WriteString(indentString(_buf.String()))
}

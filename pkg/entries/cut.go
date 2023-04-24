package entries

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrNotACutString = errors.New("field is not a cuttable string")
)

type cutOpts struct {
	field        string
	delimiter    rune
	collector    func(entry LogEntry, fields []string) (collected LogEntry, remaining string)
	removeSource bool
}

// CutOpt represents a functional option for Cut.
type CutOpt func(opts *cutOpts)

// CutField specifies the field to use as the basis for Cut.
func CutField(field string) CutOpt {
	return func(opts *cutOpts) {
		opts.field = field
	}
}

// CutDelim specifies the delimiter to use to split the field in Cut.
func CutDelim(delim rune) CutOpt {
	return func(opts *cutOpts) {
		opts.delimiter = delim
	}
}

// CutCollector defines a function that will be used to stitch Cut fields into the message.
func CutCollector(fn func(entry LogEntry, fields []string) (LogEntry, string)) CutOpt {
	return func(opts *cutOpts) {
		opts.collector = fn
	}
}

// RemoveSource specifies that the source field transformed by Cut should be removed after successful processing.
func RemoveSource() CutOpt {
	return func(opts *cutOpts) {
		opts.removeSource = true
	}
}

// CutCollectSpec will specify the destination for a field as the output of Cut.
// Any unmapped fields will be ignored.
type CutCollectSpec map[int]func(entry LogEntry, value string)

func NewCutCollectSpec() CutCollectSpec {
	return CutCollectSpec{}
}

// Map will copy the Cut field at idx to field.
// Map calls can override each other by specifying the same field and/or idx multiple times.
// Map can accept negative indexes to refer to fields at the end of a line of text, starting with -1.
// Any value set by a mapping may be overridden by another mapping. Ensure that all calls to Map reference any given field only once.
func (c CutCollectSpec) Map(field string, idx int) CutCollectSpec {
	c[idx] = func(entry LogEntry, value string) {
		entry[field] = value
	}
	return c
}

func (c CutCollectSpec) Collector() func(entry LogEntry, fields []string) (LogEntry, string) {
	return func(entry LogEntry, fields []string) (LogEntry, string) {
		var (
			firstWrite = true
			buf        strings.Builder
		)
		for i, f := range fields {
			fn, ok := c[i]
			inverseIdx := i - len(fields)
			ifn, iok := c[inverseIdx]
			switch {
			case ok, iok:
				if ok {
					fn(entry, f)
				}
				if iok {
					ifn(entry, f)
				}
			case !firstWrite:
				buf.WriteString(" ")
				fallthrough
			default:
				firstWrite = false
				buf.WriteString(f)
			}
		}
		return entry, buf.String()
	}
}

func defaultCutCollector(entry LogEntry, fields []string) (LogEntry, string) {
	for i, f := range fields {
		entry[strconv.Itoa(i)] = f
	}
	return entry, ""
}

// Cut allows programmatically parsing out a log message into more atomic parts by splitting on one or more instances of delimiter, much like the unix cut command.
// Cut assumes it should be parsing the StandardMessageField with a space character unless overridden.
// The source field must be a string for Cut to operate as intended.
func Cut(entry LogEntry, opt ...CutOpt) (LogEntry, error) {
	opts := &cutOpts{
		field:        StandardMessageField,
		delimiter:    ' ',
		removeSource: false,
	}
	for _, o := range opt {
		o(opts)
	}

	if opts.collector == nil {
		opts.collector = defaultCutCollector
	}
	if entry.HasField(opts.field) {
		str, ok := entry.AsString(opts.field)
		if !ok {
			return entry, ErrNotACutString
		}
		fields := strings.Split(str, string([]rune{opts.delimiter}))
		entry, remaining := opts.collector(entry, fields)
		if opts.removeSource {
			delete(entry, opts.field)
		} else {
			entry[opts.field] = remaining
		}
	}
	return entry, nil
}

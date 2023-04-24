package iterator

import (
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"regexp"
)

// Joiner will traverse an Iterator, returning messages that may be joined based on a set of startPatterns.
// A start pattern defines what a @message value must look like to be interpreted as the start of a log message.
// Subsequent messages that do not match this pattern will have their @message field appended to the last start line.
// If the Iterator starts with a LogEntry with a @message field that doesn't match the startPatterns, it will be added as a start anyway.
func Joiner(iter Iterator, startPatterns ...string) Iterator {
	opts := new(joinOpts).WithStartRegex(startPatterns...)
	j := &joinerState{
		iter: iter,
		opts: opts,
	}
	return Func(j.nextFunc)
}

type joinOpts struct {
	startMatchRegex []*regexp.Regexp
}

func (s *joinOpts) WithStartRegex(patterns ...string) *joinOpts {
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err == nil {
			s.startMatchRegex = append(s.startMatchRegex, r)
		}
	}
	return s
}

type joinerState struct {
	start entries.LogEntry
	opts  *joinOpts
	msg   string
	idx   int
	iter  Iterator
}

func (j *joinerState) isStart(entry entries.LogEntry) bool {
	msg, ok := entry.AsString(entries.StandardMessageField)
	if !ok {
		return false
	}

	for _, r := range j.opts.startMatchRegex {
		if r.MatchString(msg) {
			return true
		}
	}
	return false
}

func (j *joinerState) setStart(entry entries.LogEntry, idx int) {
	j.start, j.idx = entry, idx
	msg, ok := entry.AsString(entries.StandardMessageField)
	if ok {
		j.msg = msg
	}
}

func (j *joinerState) appendMessage(entry entries.LogEntry) {
	_msg, ok := entry.AsString(entries.StandardMessageField)
	if ok {
		j.msg += "\n" + _msg
	}
}

func (j *joinerState) startDefined() bool {
	return j.start != nil
}

func (j *joinerState) finalizeEntry() (entries.LogEntry, int) {
	_start, _idx := j.start, j.idx
	_start[entries.StandardMessageField] = j.msg
	j.start, j.idx = nil, -1
	j.msg = ""
	return _start, _idx
}

func (j *joinerState) nextFunc() (entries.LogEntry, int, error) {
	for {
		entry, i, err := j.iter.Next()
		switch {
		case err != nil:
			if j.startDefined() {
				final, i := j.finalizeEntry()
				return final, i, nil
			}
			return nil, -1, err
		case j.startDefined():
			if j.isStart(entry) {
				final, _i := j.finalizeEntry()
				j.setStart(entry, i)
				return final, _i, nil
			}
			j.appendMessage(entry)
			continue
		default:
			j.setStart(entry, i)
		}
	}
}

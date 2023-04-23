package iterator

import (
	"github.com/saylorsolutions/nomlog/pkg/entries"
)

// Cutter injects entries.Cut for each entry in the iterator.
func Cutter(iter Iterator, opt ...entries.CutOpt) Iterator {
	return Func(func() (entries.LogEntry, int, error) {
		entry, i, err := iter.Next()
		if err != nil {
			return nil, -1, err
		}
		entry, err = entries.Cut(entry, opt...)
		if err != nil {
			return nil, -1, err
		}
		return entry, i, nil
	})
}

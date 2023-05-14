package iterator

import (
	"github.com/saylorsolutions/nomlog/pkg/entries"
)

// Tag will set the standard tag field to the value specified in tag.
// If the standard tag field already exists, then the parameter will be appended with a period separator.
// A Tag is intended to classify the log information in some way to make it easier to filter for later.
func Tag(iter Iterator, tag string) Iterator {
	return Func(func() (entries.LogEntry, int, error) {
		entry, i, err := iter.Next()
		if err != nil {
			return Err(err)
		}
		entry.Tag(tag)
		return entry, i, nil
	})
}

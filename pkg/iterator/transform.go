package iterator

import "github.com/saylorsolutions/nomlog/pkg/entries"

// Transformer adds entries.Transform logic to this Iterator.
func Transformer(iter Iterator, spec entries.TransformSpec) Iterator {
	return Func(func() (entries.LogEntry, int, error) {
		entry, i, err := iter.Next()
		if err != nil {
			return nil, -1, err
		}
		entry = entries.Transform(entry, spec)
		return entry, i, nil
	})
}

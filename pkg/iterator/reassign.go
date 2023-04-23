package iterator

import "github.com/saylorsolutions/nomlog/pkg/entries"

// Reassignerator runs entries.Reassign on each entry that passes through the Iterator.
func Reassignerator(iter Iterator, spec entries.ReassignSpec) Iterator {
	return Func(func() (entries.LogEntry, int, error) {
		entry, i, err := iter.Next()
		if err != nil {
			return nil, -1, err
		}
		entry = entries.Reassign(entry, spec)
		return entry, i, nil
	})
}

package entries

// LogEntry is a single entry in a log, with potentially many fields.
type LogEntry map[string]any

func (e LogEntry) HasField(name string) bool {
	_, ok := e[name]
	return ok
}

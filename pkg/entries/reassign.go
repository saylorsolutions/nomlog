package entries

type SourceField string
type TargetField string

// ReassignSpec contains mapping details for renaming a field from SourceField to TargetField.
type ReassignSpec map[SourceField]TargetField

func NewReassignSpec() ReassignSpec {
	return ReassignSpec{}
}

// Move sets a value in the ReassignSpec that moves a field's value.
func (s ReassignSpec) Move(source SourceField, target TargetField) ReassignSpec {
	s[source] = target
	return s
}

func (e LogEntry) getSourceFieldVal(field SourceField) (any, bool) {
	s := string(field)
	if e.HasField(s) {
		return e[s], true
	}
	return nil, false
}

func (e LogEntry) setTargetFieldVal(field TargetField, val any) {
	e[string(field)] = val
}

// Reassign runs a ReassignSpec against a LogEntry to move one or more fields.
func Reassign(entry LogEntry, spec ReassignSpec) LogEntry {
	for s, t := range spec {
		val, ok := entry.getSourceFieldVal(s)
		if ok {
			entry.setTargetFieldVal(t, val)
			delete(entry, string(s))
		}
	}
	return entry
}

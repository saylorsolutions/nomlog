package entries

type ReassignSpec map[string]string

func NewReassignSpec() ReassignSpec {
	return ReassignSpec{}
}

func (s ReassignSpec) Move(source, target string) ReassignSpec {
	s[source] = target
	return s
}

func Reassign(entry LogEntry, spec ReassignSpec) LogEntry {
	for s, t := range spec {
		val, ok := entry[s]
		if ok {
			entry[t] = val
			delete(entry, s)
		}
	}
	return entry
}

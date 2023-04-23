package entries

type SubjectField string

type transFunc func(val any) any

func (fn transFunc) then(after transFunc) transFunc {
	if fn == nil {
		return after
	}
	return func(val any) any {
		val = fn(val)
		return after(val)
	}
}

type TransformSpec map[SubjectField]transFunc

func NewTransformSpec() TransformSpec {
	return TransformSpec{}
}

func (e LogEntry) getSubjectFieldVal(field SubjectField) any {
	return e[string(field)]
}

func (e LogEntry) setSubjectField(field SubjectField, val any) {
	s := string(field)
	if e.HasField(s) {
		e[s] = val
	}
}

// Transform will add a field transform.
// Adding a transform for a field where one is already assigned will append the given transform function to the existing one.
func (s TransformSpec) Transform(field SubjectField, trans transFunc) TransformSpec {
	s[field] = s[field].then(trans)
	return s
}

// TransformString will append a Transform for the named field if it contains a string.
func (s TransformSpec) TransformString(field SubjectField, trans func(val string) string) TransformSpec {
	s[field] = s[field].then(func(val any) any {
		if s, ok := val.(string); ok {
			return trans(s)
		}
		return val
	})
	return s
}

// TODO: Add more support for typed transforms.

func Transform(entry LogEntry, spec TransformSpec) LogEntry {
	for field, trans := range spec {
		if trans == nil {
			continue
		}
		val := entry.getSubjectFieldVal(field)
		if val != nil {
			entry.setSubjectField(field, trans(val))
		}
	}
	return entry
}

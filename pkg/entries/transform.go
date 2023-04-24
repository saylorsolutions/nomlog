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

func TransformType[T any](trans func(val T) T) func(any) any {
	return func(a any) any {
		if a == nil {
			return a
		}
		if v, ok := a.(T); ok {
			return trans(v)
		}
		return a
	}
}

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

package entries

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const (
	StandardMessageField   = "@message"
	StandardTimestampField = "@timestamp"
	StandardLevelField     = "@level"
	StandardModuleField    = "@module"
	StandardCallerField    = "@caller"
)

// LogEntry is a single entry in a log, with potentially many fields.
type LogEntry map[string]any

func (e LogEntry) HasField(name string) bool {
	_, ok := e[name]
	return ok
}

func (e LogEntry) AsFloat(name string) (float64, bool) {
	if !e.HasField(name) {
		return 0, false
	}
	if f, ok := e[name].(float64); ok {
		return f, true
	}
	if s, ok := e[name].(string); ok {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}
	v := reflect.ValueOf(e[name])
	if v.CanFloat() {
		switch v.Kind() {
		case reflect.Float64:
			return e[name].(float64), true
		case reflect.Float32:
			return float64(e[name].(float32)), true
		}
	}
	return 0, false
}

func (e LogEntry) AsInt(name string) (int64, bool) {
	if !e.HasField(name) {
		return 0, false
	}
	if i, ok := e[name].(int64); ok {
		return i, true
	}
	if s, ok := e[name].(string); ok {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	}
	v := reflect.ValueOf(e[name])
	if v.CanInt() {
		switch v.Kind() {
		case reflect.Int64:
			return e[name].(int64), true
		case reflect.Int32:
			return int64(e[name].(int32)), true
		case reflect.Int:
			return int64(e[name].(int)), true
		}
	}
	return 0, false
}

func (e LogEntry) AsUint(name string) (uint64, bool) {
	if !e.HasField(name) {
		return 0, false
	}
	if i, ok := e[name].(uint64); ok {
		return i, true
	}
	if s, ok := e[name].(string); ok {
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	}
	v := reflect.ValueOf(e[name])
	if v.CanUint() {
		switch v.Kind() {
		case reflect.Uint64:
			return e[name].(uint64), true
		case reflect.Uint32:
			return uint64(e[name].(uint32)), true
		case reflect.Uint:
			return uint64(e[name].(uint)), true
		}
	}
	return 0, false
}

func (e LogEntry) AsString(name string) (string, bool) {
	if !e.HasField(name) {
		return "", false
	}
	if s, ok := e[name].(string); ok {
		return s, true
	}
	if s, ok := e[name].(interface{ String() string }); ok {
		return s.String(), true
	}
	if err, ok := e[name].(error); ok {
		return err.Error(), true
	}
	return fmt.Sprintf("%v", e[name]), true
}

func (e LogEntry) AsTime(name string, format ...string) (time.Time, bool) {
	var none time.Time
	if !e.HasField(name) {
		return none, false
	}
	if t, ok := e[name].(time.Time); ok {
		return t.UTC(), true
	}
	if s, ok := e.AsString(name); ok {
		if len(format) > 0 {
			for _, f := range format {
				t, err := time.Parse(f, s)
				if err == nil {
					return t.UTC(), true
				}
			}
		} else {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t.UTC(), true
			}
		}
	}
	return none, false
}

func (e LogEntry) Format(format string, fields ...string) string {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = e[f]
	}
	return fmt.Sprintf(format, args...)
}

func FromString(msg string) LogEntry {
	entry := LogEntry{}
	if err := json.Unmarshal([]byte(msg), &entry); err != nil {
		entry[StandardMessageField] = msg
	}
	return entry
}

package observability

import "time"

// Field represents a key-value log attribute with explicit types.
type Field struct {
	Key      string
	StrValue string
	IntValue int64
	IsInt    bool
}

// Span represents a single trace segment capturing details and duration.
type Span struct {
	TraceID   string
	Name      string
	StartTime time.Time
}

// NewStringField constructs a Field representing a string value.
func NewStringField(key string, value string) Field {
	field := Field{
		Key:      key,
		StrValue: value,
		IsInt:    false,
	}
	return field
}

// NewIntField constructs a Field representing an integer value.
func NewIntField(key string, value int64) Field {
	field := Field{
		Key:      key,
		IntValue: value,
		IsInt:    true,
	}
	return field
}

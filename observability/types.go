package observability

import "time"

// Field represents a key-value log attribute with explicit types.
//
// Usage example:
//
//	f := observability.Field{
//		Key:      "role",
//		StrValue: "admin",
//		IsInt:    false,
//	}
type Field struct {
	Key      string
	StrValue string
	IntValue int64
	IsInt    bool
}

// Span represents a single trace segment capturing details and duration.
//
// Usage example:
//
//	s := observability.Span{
//		TraceID:   "db3bda",
//		Name:      "MySQLQuery",
//		StartTime: time.Now(),
//	}
type Span struct {
	TraceID   string
	Name      string
	StartTime time.Time
}

// NewStringField constructs a Field representing a string value.
//
// Usage example:
//
//	f := observability.NewStringField("user_id", "123")
func NewStringField(key string, value string) Field {
	field := Field{
		Key:      key,
		StrValue: value,
		IsInt:    false,
	}
	return field
}

// NewIntField constructs a Field representing an integer value.
//
// Usage example:
//
//	f := observability.NewIntField("bytes_sent", 2048)
func NewIntField(key string, value int64) Field {
	field := Field{
		Key:      key,
		IntValue: value,
		IsInt:    true,
	}
	return field
}

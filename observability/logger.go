package observability

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	LevelDebug int32 = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelMap = map[string]int32{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// Logger defines the interface for structured logging.
//
// Usage example:
//
//	var l observability.Logger = observability.NewJSONLogger(os.Stdout, "auth-service")
type Logger interface {
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	SetLevel(levelStr string)
}

// JSONLogger is a structured logger that writes to an io.Writer.
//
// Usage example:
//
//	logger := observability.NewJSONLogger(os.Stdout, "auth-service")
type JSONLogger struct {
	writer        io.Writer
	service       string
	traceProvider func() string
	level         int32
}

// NewJSONLogger creates a new JSONLogger instance.
//
// Usage example:
//
//	logger := observability.NewJSONLogger(os.Stdout, "my-service")
func NewJSONLogger(w io.Writer, service string) *JSONLogger {
	logger := &JSONLogger{
		writer:  w,
		service: service,
		level:   LevelInfo,
	}
	return logger
}

// SetTraceProvider configures a dynamic trace ID provider for log correlation.
//
// Usage example:
//
//	logger.SetTraceProvider(observability.GetActiveTraceID)
func (l *JSONLogger) SetTraceProvider(provider func() string) {
	l.traceProvider = provider
	return
}

// SetLevel changes the active log level dynamically in a thread-safe manner.
//
// Usage example:
//
//	logger.SetLevel("debug")
func (l *JSONLogger) SetLevel(levelStr string) {
	if val, exists := levelMap[strings.ToLower(levelStr)]; exists {
		atomic.StoreInt32(&l.level, val)
	}
}

// Info logs an informational message.
//
// Usage example:
//
//	logger.Info("processing request", observability.NewStringField("user", "admin"))
func (l *JSONLogger) Info(msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelInfo {
		return
	}
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("INFO", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Debug logs a debug-level message.
//
// Usage example:
//
//	logger.Debug("cache miss", observability.NewStringField("key", "user_1"))
func (l *JSONLogger) Debug(msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelDebug {
		return
	}
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("DEBUG", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Warn logs a warning message.
//
// Usage example:
//
//	logger.Warn("slow query performance")
func (l *JSONLogger) Warn(msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelWarn {
		return
	}
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("WARN", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Error logs an error message.
//
// Usage example:
//
//	logger.Error("connection failure", err)
func (l *JSONLogger) Error(msg string, err error, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelError {
		return
	}
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	errFields := append(fields, NewStringField("error", err.Error()))
	jsonStr := formatJSON("ERROR", l.service, traceID, msg, errFields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// formatJSON constructs a raw JSON formatted string for the log.
func formatJSON(level string, service string, traceID string, msg string, fields []Field) string {
	var buf bytes.Buffer
	buf.WriteString(`{"level":"` + level + `",`)
	buf.WriteString(`"service":"` + service + `",`)
	if traceID != "" {
		buf.WriteString(`"trace_id":"` + traceID + `",`)
	}
	buf.WriteString(`"msg":"` + msg + `"`)
	writeFields(&buf, fields)
	buf.WriteString("}\n")
	return buf.String()
}

// writeFields appends typed log fields cleanly to the JSON buffer.
func writeFields(buf *bytes.Buffer, fields []Field) {
	for _, f := range fields {
		buf.WriteString(",")
		buf.WriteString(`"` + f.Key + `":`)
		if f.IsInt {
			buf.WriteString(strconv.FormatInt(f.IntValue, 10))
		} else {
			buf.WriteString(`"` + f.StrValue + `"`)
		}
	}
}

package observability

import (
	"bytes"
	"io"
	"strconv"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
}

// JSONLogger is a structured logger that writes to an io.Writer.
type JSONLogger struct {
	writer        io.Writer
	service       string
	traceProvider func() string
}

// NewJSONLogger creates a new JSONLogger instance.
func NewJSONLogger(w io.Writer, service string) *JSONLogger {
	logger := &JSONLogger{
		writer:  w,
		service: service,
	}
	return logger
}

// SetTraceProvider configures a dynamic trace ID provider for log correlation.
func (l *JSONLogger) SetTraceProvider(provider func() string) {
	l.traceProvider = provider
	return
}

// Info logs an informational message.
func (l *JSONLogger) Info(msg string, fields ...Field) {
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("INFO", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Debug logs a debug-level message.
func (l *JSONLogger) Debug(msg string, fields ...Field) {
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("DEBUG", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Warn logs a warning message.
func (l *JSONLogger) Warn(msg string, fields ...Field) {
	traceID := ""
	if l.traceProvider != nil {
		traceID = l.traceProvider()
	}
	jsonStr := formatJSON("WARN", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Error logs an error message.
func (l *JSONLogger) Error(msg string, err error, fields ...Field) {
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

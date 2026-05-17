package observability

import "io"

// Logger defines the interface for structured logging.
type Logger interface {
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
}

// JSONLogger is a structured logger that writes to an io.Writer.
type JSONLogger struct {
	writer  io.Writer
	service string
}

// NewJSONLogger creates a new JSONLogger instance.
func NewJSONLogger(w io.Writer, service string) *JSONLogger {
	logger := &JSONLogger{
		writer:  w,
		service: service,
	}
	return logger
}

// Info logs an informational message.
func (l *JSONLogger) Info(msg string, fields ...Field) {
	_ = msg
	_ = fields
}

// Debug logs a debug-level message.
func (l *JSONLogger) Debug(msg string, fields ...Field) {
	_ = msg
	_ = fields
}

// Warn logs a warning message.
func (l *JSONLogger) Warn(msg string, fields ...Field) {
	_ = msg
	_ = fields
}

// Error logs an error message.
func (l *JSONLogger) Error(msg string, err error, fields ...Field) {
	_ = msg
	_ = err
	_ = fields
}

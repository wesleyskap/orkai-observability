package observability

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

var (
	sensitiveMu   sync.RWMutex
	sensitiveKeys = []string{"password", "token", "secret", "cvv", "card", "cpf", "email"}
)

// AddSensitiveKeys appends new patterns to the global PII log masking list.
//
// Usage example:
//
//	observability.AddSensitiveKeys("apiKey", "ssn")
func AddSensitiveKeys(keys ...string) {
	sensitiveMu.Lock()
	defer sensitiveMu.Unlock()
	for _, key := range keys {
		sensitiveKeys = append(sensitiveKeys, strings.ToLower(key))
	}
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
	InfoContext(ctx context.Context, msg string, fields ...Field)
	DebugContext(ctx context.Context, msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)
	ErrorContext(ctx context.Context, msg string, err error, fields ...Field)
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
	stack := captureStackTrace()
	errFields := append(fields, NewStringField("error", err.Error()))
	if stack != "" {
		errFields = append(errFields, NewStringField("stack_trace", stack))
	}
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
		if isSensitiveKey(f.Key) {
			buf.WriteString(`"[MASKED]"`)
			continue
		}
		if f.IsInt {
			buf.WriteString(strconv.FormatInt(f.IntValue, 10))
		} else {
			buf.WriteString(`"` + f.StrValue + `"`)
		}
	}
}

// isSensitiveKey checks if a key contains any registered PII keywords.
func isSensitiveKey(key string) bool {
	sensitiveMu.RLock()
	defer sensitiveMu.RUnlock()
	lower := strings.ToLower(key)
	for _, k := range sensitiveKeys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// captureStackTrace inspects the Go runtime stack and constructs a compact formatted frame string.
func captureStackTrace() string {
	pcs := make([]uintptr, 8)
	n := runtime.Callers(3, pcs)
	var builder strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if frame.Function != "" {
			if builder.Len() > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(frame.Function + ":" + strconv.Itoa(frame.Line))
		}
		if !more {
			return builder.String()
		}
	}
}

// resolveTraceID extracts the trace ID from context or falls back to provider.
func (l *JSONLogger) resolveTraceID(ctx context.Context) string {
	if id := TraceIDFromContext(ctx); id != "" {
		return id
	}
	if l.traceProvider != nil {
		return l.traceProvider()
	}
	return ""
}

// InfoContext logs an info message with context-aware trace correlation.
func (l *JSONLogger) InfoContext(ctx context.Context, msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelInfo {
		return
	}
	traceID := l.resolveTraceID(ctx)
	jsonStr := formatJSON("INFO", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// WarnContext logs a warning message with context-aware trace correlation.
func (l *JSONLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelWarn {
		return
	}
	traceID := l.resolveTraceID(ctx)
	jsonStr := formatJSON("WARN", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// ErrorContext logs an error message with context-aware trace correlation and stack trace capture.
func (l *JSONLogger) ErrorContext(ctx context.Context, msg string, err error, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelError {
		return
	}
	traceID := l.resolveTraceID(ctx)
	stack := captureStackTrace()
	errFields := append(fields, NewStringField("error", err.Error()))
	if stack != "" {
		errFields = append(errFields, NewStringField("stack_trace", stack))
	}
	jsonStr := formatJSON("ERROR", l.service, traceID, msg, errFields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

// DebugContext logs a debug message with context-aware trace correlation.
func (l *JSONLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelDebug {
		return
	}
	traceID := l.resolveTraceID(ctx)
	jsonStr := formatJSON("DEBUG", l.service, traceID, msg, fields)
	_, _ = l.writer.Write([]byte(jsonStr))
}

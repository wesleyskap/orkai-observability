package observability

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	sensitiveMu      sync.RWMutex
	sensitiveKeys    = []string{"password", "token", "secret", "cvv", "card", "cpf", "email"}
	sensitiveRegexes = map[string]*regexp.Regexp{
		"cpf":    regexp.MustCompile(`\d{3}\.\d{3}\.\d{3}-\d{2}`),
		"jwt":    regexp.MustCompile(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`),
		"credit": regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
	}
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

// RegisterPIIPattern registers a new regex pattern for PII sanitization.
//
// Usage example:
//
//	observability.RegisterPIIPattern("cnpj", regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}`))
func RegisterPIIPattern(name string, pattern *regexp.Regexp) {
	sensitiveMu.Lock()
	defer sensitiveMu.Unlock()
	sensitiveRegexes[name] = pattern
}

func maskPII(val string) string {
	sensitiveMu.RLock()
	defer sensitiveMu.RUnlock()
	res := val
	for _, re := range sensitiveRegexes {
		res = re.ReplaceAllString(res, "[MASKED_PATTERN]")
	}
	return res
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
	environment   string
	traceProvider func() string
	rateLimiter   *LogRateLimiter
	level         int32
	asyncEnabled  bool
	asyncChan     chan string
	asyncStop     chan struct{}
	asyncWg       sync.WaitGroup
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

// SetEnvironment configures the application environment name.
//
// Usage example:
//
//	logger.SetEnvironment("dev")
func (l *JSONLogger) SetEnvironment(env string) {
	l.environment = env
}

// SetRateLimiter configures rate limiting on the JSONLogger.
//
// Usage example:
//
//	logger.SetRateLimiter(limiter)
func (l *JSONLogger) SetRateLimiter(limiter *LogRateLimiter) {
	l.rateLimiter = limiter
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
	l.writeEntry("INFO", traceID, msg, fields)
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
	l.writeEntry("DEBUG", traceID, msg, fields)
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
	l.writeEntry("WARN", traceID, msg, fields)
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
	l.writeEntry("ERROR", traceID, msg, errFields)
}

// writeEntry handles rate limiting, sampling flags, serialization, and writer delivery.
func (l *JSONLogger) writeEntry(level string, traceID string, msg string, fields []Field) {
	if l.rateLimiter != nil {
		if allow, throttled := l.rateLimiter.Allow(); !allow {
			Counter("observability_dropped_logs_total")
			return
		} else if throttled {
			fields = append(fields, NewStringField("log_burst_throttled", "true"))
		}
	}
	var logStr string
	if l.environment == "dev" {
		logStr = formatDevConsole(level, traceID, msg, fields)
	} else {
		logStr = formatJSON(level, l.service, traceID, msg, fields)
	}
	l.deliverLog(logStr)
}

// ConfigureAsync initializes the background worker and channel for asynchronous logging.
//
// Usage example:
//
//	logger.ConfigureAsync(true, 4096)
func (l *JSONLogger) ConfigureAsync(enabled bool, size int) {
	if !enabled {
		return
	}
	if size <= 0 {
		size = 4096
	}
	l.asyncEnabled = true
	l.asyncChan = make(chan string, size)
	l.asyncStop = make(chan struct{})
	l.asyncWg.Add(1)
	go l.asyncWorker()
}

func (l *JSONLogger) asyncWorker() {
	defer l.asyncWg.Done()
	for {
		select {
		case logStr, ok := <-l.asyncChan:
			if !ok {
				return
			}
			_, _ = l.writer.Write([]byte(logStr))
		case <-l.asyncStop:
			l.flushChan()
			return
		}
	}
}

func (l *JSONLogger) flushChan() {
	for {
		select {
		case logStr, ok := <-l.asyncChan:
			if ok {
				_, _ = l.writer.Write([]byte(logStr))
				continue
			}
		default:
		}
		break
	}
}

func (l *JSONLogger) deliverLog(jsonStr string) {
	if l.asyncEnabled {
		ratio := float64(len(l.asyncChan)) / float64(cap(l.asyncChan))
		Gauge("observability_async_buffer_saturation_ratio", ratio)
		select {
		case l.asyncChan <- jsonStr:
		default:
			Counter("observability_dropped_logs_total")
			_, _ = l.writer.Write([]byte(jsonStr))
		}
		return
	}
	_, _ = l.writer.Write([]byte(jsonStr))
}

// Close gracefully flushes all pending logs and terminates background worker goroutines.
//
// Usage example:
//
//	err := logger.Close()
func (l *JSONLogger) Close() error {
	if l.asyncEnabled {
		close(l.asyncStop)
		l.asyncWg.Wait()
		close(l.asyncChan)
		l.asyncEnabled = false
	}
	if closer, ok := l.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
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
			masked := maskPII(f.StrValue)
			buf.WriteString(`"` + masked + `"`)
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
	fields = appendBaggageFields(ctx, fields)
	l.writeEntry("INFO", traceID, msg, fields)
}

// WarnContext logs a warning message with context-aware trace correlation.
func (l *JSONLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelWarn {
		return
	}
	traceID := l.resolveTraceID(ctx)
	fields = appendBaggageFields(ctx, fields)
	l.writeEntry("WARN", traceID, msg, fields)
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
	errFields = appendBaggageFields(ctx, errFields)
	l.writeEntry("ERROR", traceID, msg, errFields)
}

// DebugContext logs a debug message with context-aware trace correlation.
func (l *JSONLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
	if atomic.LoadInt32(&l.level) > LevelDebug {
		return
	}
	traceID := l.resolveTraceID(ctx)
	fields = appendBaggageFields(ctx, fields)
	l.writeEntry("DEBUG", traceID, msg, fields)
}

func appendBaggageFields(ctx context.Context, fields []Field) []Field {
	bag := BaggageFromContext(ctx)
	for k, v := range bag {
		fields = append(fields, NewStringField("baggage."+k, v))
	}
	return fields
}

func getColorForLevel(level string) (string, string) {
	reset := "\033[0m"
	switch level {
	case "DEBUG":
		return "\033[90m", reset
	case "INFO":
		return "\033[32m", reset
	case "WARN":
		return "\033[33m", reset
	case "ERROR":
		return "\033[31m", reset
	}
	return reset, reset
}

func formatTraceID(traceID string) string {
	if traceID == "" {
		return ""
	}
	shortID := traceID
	if len(traceID) > 7 {
		shortID = traceID[:7]
	}
	return " \033[36m[trace_id: " + shortID + "]\033[0m"
}

func formatDevFields(fields []Field) (string, string) {
	var extra []string
	var stack string
	for _, f := range fields {
		if f.Key == "stack_trace" {
			stack = f.StrValue
			continue
		}
		val := getFieldValueStr(f)
		extra = append(extra, f.Key+"="+val)
	}
	fieldsPart := ""
	if len(extra) > 0 {
		fieldsPart = " | " + strings.Join(extra, " ")
	}
	return fieldsPart, stack
}

func getFieldValueStr(f Field) string {
	if f.IsInt {
		return strconv.FormatInt(f.IntValue, 10)
	}
	return maskPII(f.StrValue)
}

func formatDevConsole(level string, traceID string, msg string, fields []Field) string {
	ts := time.Now().Format("15:04:05")
	color, reset := getColorForLevel(level)
	tracePart := formatTraceID(traceID)
	fieldsPart, stack := formatDevFields(fields)
	line := ts + " " + color + "[" + level + "]" + reset + tracePart + " " + msg + fieldsPart
	if stack != "" {
		line += "\n\t\033[31mStack Trace: " + stack + "\033[0m"
	}
	return line + "\n"
}

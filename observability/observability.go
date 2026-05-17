package observability

import (
	"context"
	"os"
	"time"
)

// GlobalFacade controls the global logger, metrics, and tracer.
//
// Usage example:
//	facade := &observability.GlobalFacade{
//		Logger:  logger,
//		Metrics: metrics,
//		Tracer:  tracer,
//	}
type GlobalFacade struct {
	Logger  Logger
	Metrics Metrics
	Tracer  Tracer
}

// globalInstance is the internal singleton facade.
var globalInstance *GlobalFacade

// Init initializes the global observability facade instance.
//
// Usage example:
//	err := observability.Init(cfg)
func Init(cfg Config) error {
	logger := NewJSONLogger(os.Stdout, cfg.ServiceName)
	logger.SetLevel(cfg.LogLevel)
	logger.SetTraceProvider(GetActiveTraceID)
	globalInstance = &GlobalFacade{
		Logger:  logger,
		Metrics: NewInMemoryMetrics(cfg.ServiceName),
		Tracer:  NewLocalTracer(cfg.ServiceName),
	}
	return nil
}

// SetLogLevel delegates a log level change to the global logger.
//
// Usage example:
//	observability.SetLogLevel("debug")
func SetLogLevel(levelStr string) {
	if globalInstance != nil {
		globalInstance.Logger.SetLevel(levelStr)
	}
}

// Info delegates an informational message to the global logger.
//
// Usage example:
//	observability.Info("processing order", observability.NewStringField("id", "321"))
func Info(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Info(msg, fields...)
	}
}

// Debug delegates a debug message to the global logger.
//
// Usage example:
//	observability.Debug("cache lookup result", observability.NewStringField("status", "hit"))
func Debug(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Debug(msg, fields...)
	}
}

// Warn delegates a warning message to the global logger.
//
// Usage example:
//	observability.Warn("deprecated API route called")
func Warn(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Warn(msg, fields...)
	}
}

// Error delegates an error message to the global logger.
//
// Usage example:
//	observability.Error("failed to process credit card payment", err)
func Error(msg string, err error, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Error(msg, err, fields...)
	}
}

// Counter increments the global counter metric.
//
// Usage example:
//	observability.Counter("http_requests_total")
func Counter(name string) {
	if globalInstance != nil {
		globalInstance.Metrics.IncCounter(name)
	}
}

// Latency records a latency metric globally.
//
// Usage example:
//	observability.Latency("db_query_duration", duration)
func Latency(name string, d time.Duration) {
	if globalInstance != nil {
		globalInstance.Metrics.RecordLatency(name, d)
	}
}

// Gauge sets a gauge metric globally.
//
// Usage example:
//	observability.Gauge("thread_count", 42)
func Gauge(name string, value float64) {
	if globalInstance != nil {
		globalInstance.Metrics.SetGauge(name, value)
	}
}

// StartSpan starts a global span trace.
//
// Usage example:
//	ctx, span := observability.StartSpan(context.Background(), "DBQuery")
func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if globalInstance != nil {
		return globalInstance.Tracer.StartTrace(ctx, name)
	}
	return ctx, Span{}
}

// EndSpan ends a global span trace.
//
// Usage example:
//	observability.EndSpan(span)
func EndSpan(span Span) {
	if globalInstance != nil {
		globalInstance.Tracer.EndTrace(span)
	}
}

// Dump prints all metrics collected globally.
//
// Usage example:
//	observability.Dump()
func Dump() {
	if globalInstance != nil {
		globalInstance.Metrics.Print()
	}
}

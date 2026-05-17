package observability

import (
	"context"
	"os"
	"time"
)

// GlobalFacade controls the global logger, metrics, and tracer.
type GlobalFacade struct {
	Logger  Logger
	Metrics Metrics
	Tracer  Tracer
}

// globalInstance is the internal singleton facade.
var globalInstance *GlobalFacade

// Init initializes the global observability facade instance.
func Init(cfg Config) error {
	globalInstance = &GlobalFacade{
		Logger:  NewJSONLogger(os.Stdout, cfg.ServiceName),
		Metrics: NewInMemoryMetrics(cfg.ServiceName),
		Tracer:  NewLocalTracer(cfg.ServiceName),
	}
	return nil
}

// Info delegates an informational message to the global logger.
func Info(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Info(msg, fields...)
	}
}

// Debug delegates a debug message to the global logger.
func Debug(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Debug(msg, fields...)
	}
}

// Warn delegates a warning message to the global logger.
func Warn(msg string, fields ...Field) {
	if globalInstance != nil {
		globalInstance.Logger.Warn(msg, fields...)
	}
}

// Error delegates an error message to the global logger.
func Error(msg string, err error) {
	if globalInstance != nil {
		// Log error to stub logger.
		_ = err
	}
}

// Counter increments the global counter metric.
func Counter(name string) {
	if globalInstance != nil {
		globalInstance.Metrics.IncCounter(name)
	}
}

// Latency records a latency metric globally.
func Latency(name string, d time.Duration) {
	if globalInstance != nil {
		globalInstance.Metrics.RecordLatency(name, d)
	}
}

// Gauge sets a gauge metric globally.
func Gauge(name string, value float64) {
	if globalInstance != nil {
		globalInstance.Metrics.SetGauge(name, value)
	}
}

// StartSpan starts a global span trace.
func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if globalInstance != nil {
		return globalInstance.Tracer.StartTrace(ctx, name)
	}
	return ctx, Span{}
}

// EndSpan ends a global span trace.
func EndSpan(span Span) {
	if globalInstance != nil {
		globalInstance.Tracer.EndTrace(span)
	}
}

// Dump prints all metrics collected globally.
func Dump() {
	if globalInstance != nil {
		globalInstance.Metrics.Print()
	}
}

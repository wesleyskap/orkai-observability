package observability

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"time"
)

// GlobalFacade controls the global logger, metrics, and tracer.
//
// Usage example:
//
//	facade := &observability.GlobalFacade{
//		Logger:  logger,
//		Metrics: metrics,
//		Tracer:  tracer,
//	}
type GlobalFacade struct {
	Logger  Logger
	Metrics Metrics
	Tracer  Tracer
	Config  Config
}

// globalInstance is the internal singleton facade.
var (
	globalInstance        atomic.Pointer[GlobalFacade]
	cancelSystemTelemetry context.CancelFunc
)

// Init initializes the global observability facade instance.
//
// Usage example:
//
//	err := observability.Init(cfg)
func Init(cfg Config) error {
	logger, err := initLogger(cfg)
	if err != nil {
		return err
	}
	m, t := initMetricsAndTracer(cfg)
	inst := &GlobalFacade{Logger: logger, Metrics: m, Tracer: t, Config: cfg}
	globalInstance.Store(inst)
	initSystemTelemetry(cfg)
	return nil
}

func initLogger(cfg Config) (Logger, error) {
	var writer io.Writer = os.Stdout
	if cfg.LogFilePath != "" {
		w, err := NewRotatingFileWriter(cfg.LogFilePath, cfg.LogFileMaxSize, cfg.LogFileMaxBackups)
		if err != nil {
			return nil, err
		}
		writer = w
	}
	logger := NewJSONLogger(writer, cfg.ServiceName)
	logger.SetEnvironment(cfg.Environment)
	logger.SetLevel(cfg.LogLevel)
	logger.SetTraceProvider(GetActiveTraceID)
	if cfg.EnableRateLimit {
		logger.SetRateLimiter(NewLogRateLimiter(cfg.RateLimitBurst, cfg.RateLimitRate, true))
	}
	if cfg.EnableAsyncLog {
		logger.ConfigureAsync(true, cfg.AsyncLogChannelSize)
	}
	return logger, nil
}

func initMetricsAndTracer(cfg Config) (Metrics, Tracer) {
	var m Metrics = NewInMemoryMetrics(cfg.ServiceName)
	var t Tracer = NewLocalTracer(cfg.ServiceName)
	if cfg.EnableOTel {
		m = NewOTelMetrics(cfg.OTelMeterProvider, cfg.ServiceName)
		t = NewOTelTracer(cfg.OTelTracerProvider, cfg.ServiceName)
	}
	return m, t
}

func initSystemTelemetry(cfg Config) {
	if cfg.EnableSystemTelemetry {
		ctx, cancel := context.WithCancel(context.Background())
		cancelSystemTelemetry = cancel
		StartSystemTelemetry(ctx, cfg.SystemTelemetryInterval)
	}
}

// SetLogLevel delegates a log level change to the global logger.
//
// Usage example:
//
//	observability.SetLogLevel("debug")
func SetLogLevel(levelStr string) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.SetLevel(levelStr)
	}
}

// Info delegates an informational message to the global logger.
//
// Usage example:
//
//	observability.Info("processing order", observability.NewStringField("id", "321"))
func Info(msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.Info(msg, fields...)
	}
}

// Debug delegates a debug message to the global logger.
//
// Usage example:
//
//	observability.Debug("cache lookup result", observability.NewStringField("status", "hit"))
func Debug(msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.Debug(msg, fields...)
	}
}

// Warn delegates a warning message to the global logger.
//
// Usage example:
//
//	observability.Warn("deprecated API route called")
func Warn(msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.Warn(msg, fields...)
	}
}

// Error delegates an error message to the global logger.
//
// Usage example:
//
//	observability.Error("failed to process credit card payment", err)
func Error(msg string, err error, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.Error(msg, err, fields...)
	}
}

// InfoContext delegates a context-aware informational message to the global logger.
//
// Usage example:
//
//	observability.InfoContext(ctx, "processing order", observability.NewStringField("id", "321"))
func InfoContext(ctx context.Context, msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.InfoContext(ctx, msg, fields...)
	}
}

// DebugContext delegates a context-aware debug message to the global logger.
//
// Usage example:
//
//	observability.DebugContext(ctx, "cache lookup result", observability.NewStringField("status", "hit"))
func DebugContext(ctx context.Context, msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.DebugContext(ctx, msg, fields...)
	}
}

// WarnContext delegates a context-aware warning message to the global logger.
//
// Usage example:
//
//	observability.WarnContext(ctx, "deprecated API route called")
func WarnContext(ctx context.Context, msg string, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.WarnContext(ctx, msg, fields...)
	}
}

// ErrorContext delegates a context-aware error message to the global logger.
//
// Usage example:
//
//	observability.ErrorContext(ctx, "payment fail", err)
func ErrorContext(ctx context.Context, msg string, err error, fields ...Field) {
	if inst := getGlobal(); inst != nil {
		inst.Logger.ErrorContext(ctx, msg, err, fields...)
	}
}

// Counter increments the global counter metric.
//
// Usage example:
//
//	observability.Counter("http_requests_total")
func Counter(name string) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.IncCounter(name)
	}
}

// CounterWithLabels increments the global counter metric with labels.
//
// Usage example:
//
//	observability.CounterWithLabels("http_requests_total", map[string]string{"method": "POST"})
func CounterWithLabels(name string, labels map[string]string) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.IncCounterWithLabels(name, labels)
	}
}

// Latency records a latency metric globally.
//
// Usage example:
//
//	observability.Latency("db_query_duration", duration)
func Latency(name string, d time.Duration) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.RecordLatency(name, d)
	}
}

// LatencyWithLabels records a latency metric with labels globally.
//
// Usage example:
//
//	observability.LatencyWithLabels("db_query", 10*time.Millisecond, map[string]string{"op": "select"})
func LatencyWithLabels(name string, d time.Duration, labels map[string]string) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.RecordLatencyWithLabels(name, d, labels)
	}
}

// Gauge sets a gauge metric globally.
//
// Usage example:
//
//	observability.Gauge("thread_count", 42)
func Gauge(name string, value float64) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.SetGauge(name, value)
	}
}

// GaugeWithLabels sets a gauge metric with labels globally.
//
// Usage example:
//
//	observability.GaugeWithLabels("cpu_usage", 85.5, map[string]string{"core": "0"})
func GaugeWithLabels(name string, value float64, labels map[string]string) {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.SetGaugeWithLabels(name, value, labels)
	}
}

// StartSpan starts a global span trace.
//
// Usage example:
//
//	ctx, span := observability.StartSpan(context.Background(), "DBQuery")
func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if inst := getGlobal(); inst != nil {
		return inst.Tracer.StartTrace(ctx, name)
	}
	return ctx, Span{}
}

// EndSpan ends a global span trace.
//
// Usage example:
//
//	observability.EndSpan(span)
func EndSpan(span Span) {
	if inst := getGlobal(); inst != nil {
		inst.Tracer.EndTrace(span)
	}
}

// Dump prints all metrics collected globally.
//
// Usage example:
//
//	observability.Dump()
func Dump() {
	if inst := getGlobal(); inst != nil {
		inst.Metrics.Print()
	}
}

// GetSummary retrieves the current global metrics snapshot.
//
// Usage example:
//
//	summary := observability.GetSummary()
func GetSummary() MetricsSummary {
	if inst := getGlobal(); inst != nil {
		return inst.Metrics.GetSummary()
	}
	return MetricsSummary{}
}

// SetLogger overrides the active global logger.
//
// Usage example:
//
//	observability.SetLogger(customLogger)
func SetLogger(l Logger) {
	if inst := getGlobal(); inst != nil {
		inst.Logger = l
	}
}

// Close gracefully flushes all pending logs and terminates any async resources.
//
// Usage example:
//
//	defer observability.Close()
func Close() {
	if cancelSystemTelemetry != nil {
		cancelSystemTelemetry()
	}
	if inst := getGlobal(); inst != nil {
		if closer, ok := inst.Logger.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
}

func getGlobal() *GlobalFacade {
	return globalInstance.Load()
}

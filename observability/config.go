// Package observability provides a unified, thread-safe observability facade
// for structured JSON logging, metrics aggregation, and nested context tracing.
package observability

import (
	"errors"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Config defines the configuration settings for the observability package.
//
// Usage example:
//
//	cfg := observability.Config{
//		ServiceName:         "user-service",
//		Environment:         "production",
//		LogLevel:            "info",
//		EnableRateLimit:     true,
//		RateLimitBurst:      100,
//		RateLimitRate:       50,
//		EnableAsyncLog:      true,
//		AsyncLogChannelSize: 4096,
//		EnableOTel:          true,
//	}
type Config struct {
	ServiceName             string
	Environment             string
	LogLevel                string
	LogFilePath             string
	OTLPEndpoint            string
	PprofOutputDir          string
	OTelTracerProvider      trace.TracerProvider
	OTelMeterProvider       metric.MeterProvider
	OTLPHeaders             map[string]string
	SystemTelemetryInterval time.Duration
	SlowQueryThreshold      time.Duration
	ExportInterval          time.Duration
	PprofProfileDuration    time.Duration
	LogFileMaxSize          int64
	PprofHeapThresholdBytes int64
	PayloadLoggingSample    float64
	RateLimitBurst          int
	RateLimitRate           int
	AsyncLogChannelSize     int
	LogFileMaxBackups       int
	MaxPayloadLogSizeBytes  int
	PprofGoroutinesLimit    int
	EnableRateLimit         bool
	EnableAsyncLog          bool
	EnableOTel              bool
	EnableSystemTelemetry   bool
	EnableSlowQueryAlert    bool
	EnablePayloadLogging    bool
	EnableAutoPprof         bool
}

// ValidateConfig verifies that the configuration fields are not empty.
//
// Usage example:
//
//	err := observability.ValidateConfig(cfg)
func ValidateConfig(cfg Config) error {
	if cfg.ServiceName == "" {
		return errors.New("ServiceName cannot be empty")
	}
	if cfg.Environment == "" {
		return errors.New("Environment cannot be empty")
	}
	return nil
}

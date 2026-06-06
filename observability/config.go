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
	EnableRateLimit         bool
	RateLimitBurst          int
	RateLimitRate           int
	EnableAsyncLog          bool
	AsyncLogChannelSize     int
	EnableOTel              bool
	OTelTracerProvider      trace.TracerProvider
	OTelMeterProvider       metric.MeterProvider
	EnableSystemTelemetry   bool
	SystemTelemetryInterval time.Duration
	LogFilePath             string
	LogFileMaxSize          int64
	LogFileMaxBackups       int
	EnableSlowQueryAlert    bool
	SlowQueryThreshold      time.Duration
	PayloadLoggingSample    float64
	MaxPayloadLogSizeBytes  int
	EnablePayloadLogging    bool
	OTLPEndpoint            string
	OTLPHeaders             map[string]string
	ExportInterval          time.Duration
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

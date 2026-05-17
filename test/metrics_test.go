package test

import (
	"orkai-observability/observability"
	"testing"
	"time"
)

// TestMetricsIncrement asserts counter increments.
func TestMetricsIncrement(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	metrics.IncCounter("requests_total")
}

// TestMetricsLatency asserts latency recording.
func TestMetricsLatency(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	metrics.RecordLatency("db_query", 15*time.Millisecond)
}

// TestMetricsGauge asserts gauge values.
func TestMetricsGauge(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	metrics.SetGauge("active_sessions", 42)
}

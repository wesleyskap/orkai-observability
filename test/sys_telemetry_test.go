package test

import (
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// TestSystemTelemetryCollector asserts that runtime metrics are collected periodically.
func TestSystemTelemetryCollector(t *testing.T) {
	cfg := observability.Config{
		ServiceName:             "test-telemetry",
		Environment:             "dev",
		EnableSystemTelemetry:   true,
		SystemTelemetryInterval: 10 * time.Millisecond,
	}
	err := observability.Init(cfg)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer observability.Close()
	time.Sleep(50 * time.Millisecond)
	verifySystemMetricsCollected(t)
}

func verifySystemMetricsCollected(t *testing.T) {
	summary := observability.GetSummary()
	val, exists := summary.Gauges["go_goroutines"]
	if !exists {
		t.Fatal("expected 'go_goroutines' gauge to exist")
	}
	if val <= 0 {
		t.Fatalf("expected positive goroutine count, got %g", val)
	}
	if _, exists := summary.Gauges["go_mem_heap_alloc_bytes"]; !exists {
		t.Fatal("expected 'go_mem_heap_alloc_bytes' gauge to exist")
	}
}

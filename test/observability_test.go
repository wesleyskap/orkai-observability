package test

import (
	"context"
	"orkai-observability/observability"
	"testing"
	"time"
)

// TestGlobalFacadeInit verifies global facade initializes correctly.
func TestGlobalFacadeInit(t *testing.T) {
	cfg := observability.Config{
		ServiceName: "test-service",
		Environment: "dev",
		LogLevel:    "debug",
	}
	err := observability.Init(cfg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// TestGlobalFacadeDelegation verifies facade functions execute without panic.
func TestGlobalFacadeDelegation(t *testing.T) {
	_ = observability.Init(observability.Config{ServiceName: "test-service", Environment: "dev"})
	observability.Info("delegated log", observability.NewStringField("key", "val"))
	observability.Counter("test_count")
	observability.Latency("test_latency", 10*time.Millisecond)
	observability.Gauge("test_gauge", 10.5)
	_, span := observability.StartSpan(context.Background(), "test-span")
	observability.EndSpan(span)
	observability.Dump()
}

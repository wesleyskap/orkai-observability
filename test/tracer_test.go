package test

import (
	"context"
	"orkai-observability/observability"
	"testing"
)

// TestTracerStart verifies trace context and span generation.
func TestTracerStart(t *testing.T) {
	tracer := observability.NewLocalTracer("test-service")
	ctx, span := tracer.StartTrace(context.Background(), "test-span")
	if span.Name != "test-span" {
		t.Fatalf("expected span name 'test-span', got %s", span.Name)
	}
	if ctx == nil {
		t.Fatal("expected returned context to be non-nil")
	}
}

// TestTracerEnd verifies trace completion executes without panic.
func TestTracerEnd(t *testing.T) {
	tracer := observability.NewLocalTracer("test-service")
	_, span := tracer.StartTrace(context.Background(), "test-span")
	tracer.EndTrace(span)
}

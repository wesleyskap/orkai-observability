package test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// TestTracerStart verifies trace context, start print, and active trace ID.
func TestTracerStart(t *testing.T) {
	tracer := observability.NewLocalTracer("test-service")
	buf := &bytes.Buffer{}
	tracer.SetWriter(buf)
	ctx, span := tracer.StartTrace(context.Background(), "test-span")
	defer tracer.EndTrace(span)
	if span.Name != "test-span" {
		t.Fatalf("expected span name 'test-span', got %s", span.Name)
	}
	output := buf.String()
	if !strings.Contains(output, "[TRACE] Start test-span trace_id=") {
		t.Fatalf("expected start trace log, got %s", output)
	}
	activeID := observability.GetActiveTraceID()
	if activeID != span.TraceID || ctx == nil {
		t.Fatalf("expected active ID %s and valid ctx", span.TraceID)
	}
}

// TestTracerEnd verifies trace completion and clears active trace ID.
func TestTracerEnd(t *testing.T) {
	tracer := observability.NewLocalTracer("test-service")
	buf := &bytes.Buffer{}
	tracer.SetWriter(buf)
	_, span := tracer.StartTrace(context.Background(), "test-span")
	tracer.EndTrace(span)
	output := buf.String()
	if !strings.Contains(output, "[TRACE] End test-span duration=") {
		t.Fatalf("expected end trace log, got %s", output)
	}
	activeID := observability.GetActiveTraceID()
	if activeID != "" {
		t.Fatalf("expected active ID to be empty, got %s", activeID)
	}
}

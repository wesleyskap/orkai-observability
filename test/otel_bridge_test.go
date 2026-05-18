package test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/observability"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestOTelBridgeTracing asserts that tracing correctly bridges to native OTel SDK trace.Span.
func TestOTelBridgeTracing(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	cfg := observability.Config{
		ServiceName:        "otel-tracing-service",
		Environment:        "test",
		LogLevel:           "info",
		EnableOTel:         true,
		OTelTracerProvider: tp,
	}
	_ = observability.Init(cfg)
	_, span := observability.StartSpan(context.Background(), "otel-test-span")
	observability.EndSpan(span)
	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 ended span, got %d", len(spans))
	}
	if spans[0].Name() != "otel-test-span" {
		t.Fatalf("expected span name 'otel-test-span', got %s", spans[0].Name())
	}
}

// TestOTelBridgeMetrics asserts that metrics are correctly forwarded to OTel SDK instruments.
func TestOTelBridgeMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	cfg := observability.Config{
		ServiceName:       "otel-metrics-service",
		Environment:       "test",
		LogLevel:          "info",
		EnableOTel:        true,
		OTelMeterProvider: mp,
	}
	_ = observability.Init(cfg)
	observability.Counter("otel_requests_total")
	observability.Latency("otel_db_duration", 100*time.Millisecond)
	observability.Gauge("otel_system_memory", 1048576)
	var rm metricdata.ResourceMetrics
	err := reader.Collect(context.Background(), &rm)
	if err != nil {
		t.Fatalf("failed to collect metrics from OTel reader: %v", err)
	}
	if len(rm.ScopeMetrics) == 0 {
		t.Fatalf("expected scope metrics, got zero")
	}
	foundCounter := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "otel_requests_total" {
				foundCounter = true
			}
		}
	}
	if !foundCounter {
		t.Fatalf("expected to find metric 'otel_requests_total' in OTel scope")
	}
}

// TestOTelBridgeLogCorrelation asserts context trace correlation under active OpenTelemetry context.
func TestOTelBridgeLogCorrelation(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	cfg := observability.Config{
		ServiceName:        "otel-log-service",
		Environment:        "test",
		LogLevel:           "info",
		EnableOTel:         true,
		OTelTracerProvider: tp,
	}
	_ = observability.Init(cfg)
	buf := &bytes.Buffer{}
	facade := observability.NewJSONLogger(buf, "otel-log-service")
	facade.SetTraceProvider(observability.GetActiveTraceID)
	ctx, span := observability.StartSpan(context.Background(), "otel-span")
	facade.InfoContext(ctx, "correlated span message")
	observability.EndSpan(span)
	output := buf.String()
	if !strings.Contains(output, `"trace_id":"`+span.TraceID+`"`) {
		t.Fatalf("expected trace correlation for OTel span in log, got: %s", output)
	}
}

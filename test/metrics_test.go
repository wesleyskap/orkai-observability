package test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/observability"
)

// TestMetricsIncrement asserts counter increments and prints correctly.
func TestMetricsIncrement(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	buf := &bytes.Buffer{}
	metrics.SetWriter(buf)
	metrics.IncCounter("requests_total")
	metrics.Print()
	output := buf.String()
	if !strings.Contains(output, "requests_total: 1") {
		t.Fatalf("expected requests_total: 1, got %s", output)
	}
	summary := metrics.GetSummary()
	if val := summary.Counters["requests_total"]; val != 1 {
		t.Errorf("expected summary counter to be 1, got %d", val)
	}
}

// TestMetricsLatency asserts average latency calculation and printing.
func TestMetricsLatency(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	buf := &bytes.Buffer{}
	metrics.SetWriter(buf)
	metrics.RecordLatency("db_query", 10*time.Millisecond)
	metrics.RecordLatency("db_query", 20*time.Millisecond)
	metrics.Print()
	output := buf.String()
	if !strings.Contains(output, "db_query_latency_avg: 15ms") {
		t.Fatalf("expected db_query_latency_avg: 15ms, got %s", output)
	}
	summary := metrics.GetSummary()
	if val := summary.Latencies["db_query"]; val != 15.0 {
		t.Errorf("expected summary latency average to be 15.0, got %f", val)
	}
}

// TestMetricsGauge asserts gauge setting and printing.
func TestMetricsGauge(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	buf := &bytes.Buffer{}
	metrics.SetWriter(buf)
	metrics.SetGauge("active_sessions", 42.5)
	metrics.Print()
	output := buf.String()
	if !strings.Contains(output, "active_sessions: 42.5") {
		t.Fatalf("expected active_sessions: 42.5, got %s", output)
	}
	summary := metrics.GetSummary()
	if val := summary.Gauges["active_sessions"]; val != 42.5 {
		t.Errorf("expected summary gauge to be 42.5, got %f", val)
	}
}

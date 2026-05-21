package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

func TestMetricsPercentilesCalculation(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	for i := 1; i <= 10; i++ {
		metrics.RecordLatency("test_latency", time.Duration(i*10)*time.Millisecond)
	}
	s := metrics.GetSummary()
	pct := s.Percentiles["test_latency"]
	if pct.P50 != 60 || pct.P90 != 100 || pct.P99 != 100 {
		t.Fatalf("unexpected percentiles: P50=%f, P90=%f, P99=%f", pct.P50, pct.P90, pct.P99)
	}
}

func TestMetricsHistogramBuckets(t *testing.T) {
	metrics := observability.NewInMemoryMetrics("test-service")
	metrics.RecordLatency("test_latency", 2*time.Millisecond)
	metrics.RecordLatency("test_latency", 8*time.Millisecond)
	metrics.RecordLatency("test_latency", 12*time.Millisecond)
	metrics.RecordLatency("test_latency", 120*time.Millisecond)
	hist := metrics.GetSummary().Histograms["test_latency"]
	if hist["5"] != 1 || hist["10"] != 2 || hist["25"] != 3 || hist["250"] != 4 || hist["+Inf"] != 4 {
		t.Fatalf("unexpected histogram bucket counts: %v", hist)
	}
}

func TestPrometheusExporterHistogram(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	observability.Latency("http_duration", 10*time.Millisecond)
	observability.Latency("http_duration", 50*time.Millisecond)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics?format=prometheus", nil)
	observability.MetricsHTTPHandler().ServeHTTP(rec, req)
	body := rec.Body.String()
	assertPrometheusHistogramLines(t, body)
}

func assertPrometheusHistogramLines(t *testing.T, body string) {
	if !strings.Contains(body, "# TYPE http_duration histogram") {
		t.Errorf("expected histogram type line, got: %s", body)
	}
	if !strings.Contains(body, "http_duration_bucket{le=\"10\"} 1") {
		t.Errorf("expected bucket le=10 count 1, got: %s", body)
	}
	if !strings.Contains(body, "http_duration_sum 60") {
		t.Errorf("expected sum 60, got: %s", body)
	}
	if !strings.Contains(body, "http_duration_count 2") {
		t.Errorf("expected count 2, got: %s", body)
	}
}

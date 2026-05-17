package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/observability"
)

// TestMetricsHTTPHandlerSuccess asserts that JSON is returned by default.
func TestMetricsHTTPHandlerSuccess(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	observability.Counter("test_counter")
	handler := observability.MetricsHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var summary observability.MetricsSummary
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if val := summary.Counters["test_counter"]; val != 1 {
		t.Errorf("expected test_counter to be 1, got %d", val)
	}
}

// TestMetricsHTTPHandlerPrometheus asserts that standard Prometheus text format is served under format query param.
func TestMetricsHTTPHandlerPrometheus(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	observability.Counter("test_counter")
	observability.Gauge("test_gauge", 12.3)
	observability.Latency("test_latency", 10*time.Millisecond)

	handler := observability.MetricsHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics?format=prometheus", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# TYPE test_counter counter") {
		t.Errorf("expected test_counter counter type in body: %s", body)
	}
	if !strings.Contains(body, "test_gauge 12.3") {
		t.Errorf("expected test_gauge value in body: %s", body)
	}
}

// TestMetricsHTTPHandlerPrometheusLabels asserts Prometheus format with labels is formatted correctly.
func TestMetricsHTTPHandlerPrometheusLabels(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	observability.CounterWithLabels("api_requests", map[string]string{"path": "/users", "method": "GET"})
	handler := observability.MetricsHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics?format=prometheus", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# TYPE api_requests counter") {
		t.Errorf("expected base counter type, got: %s", body)
	}
	expectedLine := `api_requests{method="GET",path="/users"} 1`
	if !strings.Contains(body, expectedLine) {
		t.Errorf("expected formatted label line '%s', got: %s", expectedLine, body)
	}
}

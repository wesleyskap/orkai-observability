package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesleyskap/orkai-observability/observability"
)

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

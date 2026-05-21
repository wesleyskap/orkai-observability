package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

func panickingHandler(w http.ResponseWriter, req *http.Request) {
	// Triggering panic for recovery middleware test
	msg := "something went terribly wrong"
	panic(msg)
}

// TestPanicRecoveryMiddleware verifies that panics are recovered, a 500 error is returned,
// and the http_panics_total metric is incremented.
func TestPanicRecoveryMiddleware(t *testing.T) {
	cfg := observability.Config{ServiceName: "test-panic", Environment: "dev"}
	_ = observability.Init(cfg)
	mw := observability.PanicRecoveryMiddleware(http.HandlerFunc(panickingHandler))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	assertPanicResponse(t, rec)
	assertPanicMetric(t)
}

func assertPanicResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 status code, got %d", rec.Code)
	}
	expected := `{"error":"Internal Server Error"}`
	body := rec.Body.String()
	if body != expected {
		t.Errorf("expected body %q, got %q", expected, body)
	}
}

func assertPanicMetric(t *testing.T) {
	summary := observability.GetSummary()
	val, exists := summary.Counters["http_panics_total"]
	if !exists {
		t.Fatal("expected 'http_panics_total' metric to exist")
	}
	if val != 1 {
		t.Fatalf("expected panic count 1, got %d", val)
	}
}

package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesleyskap/orkai-observability/observability"
)

func TestHTTPMiddlewareNewTrace(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})
	mw := observability.HTTPMiddleware(nextHandler)
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHTTPMiddlewareResumedTrace(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := observability.GetActiveTraceID()
		if id != "db3bda" {
			t.Errorf("expected trace id db3bda, got %s", id)
		}
	})
	mw := observability.HTTPMiddleware(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("X-Trace-ID", "db3bda")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
}

package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesleyskap/orkai-observability/v2/observability"
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

func TestHTTPMiddlewareW3CTrace(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := observability.GetActiveTraceID()
		if id != "4bf92f3577b34da6a3ce929d0e0e4736" {
			t.Errorf("expected trace id 4bf92f3577b34da6a3ce929d0e0e4736, got %s", id)
		}
	})
	mw := observability.HTTPMiddleware(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
}

func TestHTTPMiddlewareB3Trace(t *testing.T) {
	cfg := observability.Config{ServiceName: "test", Environment: "dev"}
	_ = observability.Init(cfg)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := observability.GetActiveTraceID()
		if id != "80f198ee56343ba8" {
			t.Errorf("expected trace id 80f198ee56343ba8, got %s", id)
		}
	})
	mw := observability.HTTPMiddleware(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("b3", "80f198ee56343ba8-00f067aa0ba902b7-1")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
}

package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesleyskap/orkai-observability/observability"
)

// TestTracingRoundTripperNoActiveTrace asserts that no headers are injected when the LIFO trace stack is empty.
func TestTracingRoundTripperNoActiveTrace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceHeader := r.Header.Get("X-Trace-ID")
		if traceHeader != "" {
			t.Errorf("expected no X-Trace-ID header, got %s", traceHeader)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := observability.NewTracingClient()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

// TestTracingRoundTripperActiveTrace asserts that the active trace ID on the LIFO stack is automatically injected.
func TestTracingRoundTripperActiveTrace(t *testing.T) {
	expectedTrace := "test-outbound-trace-id"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceHeader := r.Header.Get("X-Trace-ID")
		if traceHeader != expectedTrace {
			t.Errorf("expected X-Trace-ID header %s, got %s", expectedTrace, traceHeader)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observability.PushActiveTraceID(expectedTrace)
	defer observability.PopActiveTraceID()

	client := observability.NewTracingClient()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

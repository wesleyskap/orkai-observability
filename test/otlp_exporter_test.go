package test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

type mockOTLPServer struct {
	server *httptest.Server
	traces []map[string]interface{}
	logs   []map[string]interface{}
}

func newMockOTLPServer() *mockOTLPServer {
	m := &mockOTLPServer{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if strings.HasSuffix(r.URL.Path, "/v1/traces") {
			m.traces = append(m.traces, payload)
		} else if strings.HasSuffix(r.URL.Path, "/v1/logs") {
			m.logs = append(m.logs, payload)
		}
		w.WriteHeader(http.StatusOK)
	}))
	return m
}

func (m *mockOTLPServer) Close() {
	m.server.Close()
}

func TestOTLPExporterTracesAndLogs(t *testing.T) {
	mockSrv := newMockOTLPServer()
	defer mockSrv.Close()

	cfg := observability.Config{
		ServiceName:    "otlp-test",
		Environment:    "production",
		OTLPEndpoint:   mockSrv.server.URL,
		ExportInterval: 50 * time.Millisecond,
	}

	err := observability.Init(cfg)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer observability.Close()

	// Trigger a span export
	_, span := observability.StartSpan(context.Background(), "my-otlp-span")
	time.Sleep(10 * time.Millisecond)
	observability.EndSpan(span)

	// Trigger a log export
	observability.Info("my otlp log message")

	// Wait for export interval to trigger
	time.Sleep(150 * time.Millisecond)

	assertTracesExported(t, mockSrv)
	assertLogsExported(t, mockSrv)
}

func assertTracesExported(t *testing.T, m *mockOTLPServer) {
	if len(m.traces) == 0 {
		t.Fatalf("expected traces to be exported, got 0 requests")
	}
	tPayload := m.traces[0]
	rs, ok := tPayload["resourceSpans"].([]interface{})
	if !ok || len(rs) == 0 {
		t.Fatalf("invalid resourceSpans format in traces payload: %v", tPayload)
	}
}

func assertLogsExported(t *testing.T, m *mockOTLPServer) {
	if len(m.logs) == 0 {
		t.Fatalf("expected logs to be exported, got 0 requests")
	}
	lPayload := m.logs[0]
	rl, ok := lPayload["resourceLogs"].([]interface{})
	if !ok || len(rl) == 0 {
		t.Fatalf("invalid resourceLogs format in logs payload: %v", lPayload)
	}
}

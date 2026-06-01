package test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// TestSlowQueryAlert checks if a slow query warning is logged when duration exceeds threshold.
func TestSlowQueryAlert(t *testing.T) {
	fw := initSlowQueryTest()
	ctx, end := observability.TraceSQL(context.Background(), "SELECT", "users")
	time.Sleep(10 * time.Millisecond)
	end()
	verifySlowQueryLog(t, fw.Buf.String())
	_ = ctx
}

func initSlowQueryTest() *FakeWriter {
	fw := &FakeWriter{}
	cfg := observability.Config{
		ServiceName:          "test-slow-query",
		Environment:          "dev",
		EnableSlowQueryAlert: true,
		SlowQueryThreshold:   5 * time.Millisecond,
	}
	_ = observability.Init(cfg)
	observability.SetLogger(observability.NewJSONLogger(fw, "test-slow-query"))
	return fw
}

func verifySlowQueryLog(t *testing.T, output string) {
	if !strings.Contains(output, `"level":"WARN"`) {
		t.Fatalf("expected WARN log for slow query, got: %s", output)
	}
	if !strings.Contains(output, `"query_type":"SELECT"`) || !strings.Contains(output, `"table":"users"`) {
		t.Errorf("expected query details in log, got: %s", output)
	}
}

// TestW3CBaggageContext checks context setting and extraction for baggage.
func TestW3CBaggageContext(t *testing.T) {
	ctx := context.Background()
	baggage := map[string]string{"user_id": "alice", "tenant_id": "t1"}
	ctx = observability.ContextWithBaggage(ctx, baggage)
	extracted := observability.BaggageFromContext(ctx)
	if extracted["user_id"] != "alice" || extracted["tenant_id"] != "t1" {
		t.Fatalf("expected baggage values to be retrieved, got: %v", extracted)
	}
}

// TestW3CBaggagePropagation checks header insertion and parsing for W3C baggage.
func TestW3CBaggagePropagation(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	baggage := map[string]string{"session_id": "123", "role": "admin"}
	observability.InjectBaggage(req, baggage)

	extracted := observability.ExtractBaggage(req)
	if extracted["session_id"] != "123" || extracted["role"] != "admin" {
		t.Fatalf("expected session_id and role in extracted baggage, got: %v", extracted)
	}
}

// TestW3CBaggageLoggerAndMiddleware checks if baggage is logged correctly.
func TestW3CBaggageLoggerAndMiddleware(t *testing.T) {
	fw := &FakeWriter{}
	cfg := observability.Config{ServiceName: "test-baggage", Environment: "dev"}
	_ = observability.Init(cfg)
	observability.SetLogger(observability.NewJSONLogger(fw, "test-baggage"))

	ctx := context.Background()
	baggage := map[string]string{"foo": "bar"}
	ctx = observability.ContextWithBaggage(ctx, baggage)

	observability.InfoContext(ctx, "hello baggage")
	output := fw.Buf.String()
	if !strings.Contains(output, `"baggage.foo":"bar"`) {
		t.Fatalf("expected baggage.foo:bar in output log, got: %s", output)
	}
}

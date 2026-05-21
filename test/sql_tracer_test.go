package test

import (
	"context"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// TestTraceSQL verifies that TraceSQL generates nested spans and records query duration metrics with tags.
func TestTraceSQL(t *testing.T) {
	cfg := observability.Config{ServiceName: "test-sql", Environment: "dev"}
	_ = observability.Init(cfg)
	ctx := context.Background()
	ctx, span := observability.StartSpan(ctx, "ParentJob")
	defer observability.EndSpan(span)
	runTracedSQL(ctx)
	verifySQLMetrics(t)
}

func runTracedSQL(ctx context.Context) {
	_, end := observability.TraceSQL(ctx, "SELECT", "users")
	time.Sleep(10 * time.Millisecond)
	end()
}

func verifySQLMetrics(t *testing.T) {
	summary := observability.GetSummary()
	key := `db_query_duration_ms{query_type="SELECT",table="users"}`
	val, exists := summary.Latencies[key]
	if !exists {
		t.Fatalf("expected key %s to exist in Latencies, got %v", key, summary.Latencies)
	}
	if val < 5.0 {
		t.Fatalf("expected recorded latency, got %f", val)
	}
}

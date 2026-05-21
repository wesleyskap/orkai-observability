package observability

import (
	"context"
	"time"
)

// TraceSQL starts a trace span for a SQL query and returns a callback to record metrics and end the span.
func TraceSQL(ctx context.Context, op string, table string) (context.Context, func()) {
	spanName := "SQL:" + op + ":" + table
	newCtx, span := StartSpan(ctx, spanName)
	start := time.Now()
	endFunc := func() {
		dur := time.Since(start)
		labels := map[string]string{
			"query_type": op,
			"table":      table,
		}
		LatencyWithLabels("db_query_duration_ms", dur, labels)
		EndSpan(span)
	}
	return newCtx, endFunc
}

package observability

import (
	"context"
	"time"
)

// TraceSQL starts a trace span for a SQL query and returns a callback to record metrics and end the span.
//
// Usage example:
//
//	ctx, end := observability.TraceSQL(ctx, "SELECT", "users")
//	defer end()
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
		checkSlowQuery(newCtx, op, table, dur)
		EndSpan(span)
	}
	return newCtx, endFunc
}

func checkSlowQuery(ctx context.Context, op string, table string, dur time.Duration) {
	if globalInstance == nil || !globalInstance.Config.EnableSlowQueryAlert {
		return
	}
	if dur < globalInstance.Config.SlowQueryThreshold {
		return
	}
	logSlowQuery(ctx, op, table, dur)
}

func logSlowQuery(ctx context.Context, op string, table string, dur time.Duration) {
	WarnContext(ctx, "slow SQL query detected",
		NewStringField("query_type", op),
		NewStringField("table", table),
		NewIntField("duration_ms", dur.Milliseconds()),
	)
}

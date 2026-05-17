package observability

import "context"

// ContextWithTraceID returns a new context carrying the specified trace ID.
//
// Usage example:
//
//	ctx := observability.ContextWithTraceID(ctx, "my-trace-id")
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from the given context if present.
//
// Usage example:
//
//	id := observability.TraceIDFromContext(ctx)
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(traceIDKey)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

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
	if span := trace.SpanFromContext(ctx); span != nil && span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	val := ctx.Value(traceIDKey)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

const baggageKey contextKey = "baggage"

// ContextWithBaggage returns a new context containing the provided baggage.
//
// Usage example:
//
//	ctx := observability.ContextWithBaggage(ctx, map[string]string{"user": "alice"})
func ContextWithBaggage(ctx context.Context, baggage map[string]string) context.Context {
	if len(baggage) == 0 {
		return ctx
	}
	return context.WithValue(ctx, baggageKey, baggage)
}

// BaggageFromContext extracts the baggage map from the given context if present.
//
// Usage example:
//
//	baggage := observability.BaggageFromContext(ctx)
func BaggageFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(baggageKey)
	if baggage, ok := val.(map[string]string); ok {
		return baggage
	}
	return nil
}

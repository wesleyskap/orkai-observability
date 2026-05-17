package observability

import (
	"context"
)

// Tracer defines the interface for tracing.
type Tracer interface {
	StartTrace(ctx context.Context, name string) (context.Context, Span)
	EndTrace(span Span)
}

// LocalTracer implements Tracer locally.
type LocalTracer struct {
	service string
}

// NewLocalTracer creates a new LocalTracer instance.
func NewLocalTracer(service string) *LocalTracer {
	tracer := &LocalTracer{
		service: service,
	}
	return tracer
}

// StartTrace starts a new span and returns the context.
func (t *LocalTracer) StartTrace(ctx context.Context, name string) (context.Context, Span) {
	span := Span{
		TraceID: "stub-id",
		Name:    name,
	}
	return ctx, span
}

// EndTrace ends a span and prints details.
func (t *LocalTracer) EndTrace(span Span) {
	_ = span
}

package observability

import (
	"context"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// HTTPMiddleware wraps an http.Handler with trace context, automated request/response logging, and latency recording.
//
// Usage example:
//	mux := http.NewServeMux()
//	loggedMux := observability.HTTPMiddleware(mux)
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceID := req.Header.Get("X-Trace-ID")
		ctx := req.Context()
		var span Span
		if traceID != "" {
			ctx, span = resumeTrace(ctx, traceID, req.URL.Path)
		} else {
			ctx, span = StartSpan(ctx, req.URL.Path)
		}
		defer EndSpan(span)
		executePipeline(w, req.WithContext(ctx), next, span)
	})
}

func resumeTrace(ctx context.Context, id string, path string) (context.Context, Span) {
	PushActiveTraceID(id)
	span := Span{
		TraceID:   id,
		Name:      path,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, traceIDKey, id), span
}

func executePipeline(w http.ResponseWriter, req *http.Request, next http.Handler, s Span) {
	wrapper := newResponseWriterWrapper(w)
	Info("incoming request started",
		NewStringField("method", req.Method),
		NewStringField("path", req.URL.Path),
	)
	next.ServeHTTP(wrapper, req)
	duration := time.Since(s.StartTime)
	Latency("http_request_duration", duration)
	Counter("http_requests_total")
	Info("outgoing request finished",
		NewStringField("method", req.Method),
		NewStringField("path", req.URL.Path),
		NewIntField("status", int64(wrapper.statusCode)),
		NewIntField("duration_ms", duration.Milliseconds()),
	)
}

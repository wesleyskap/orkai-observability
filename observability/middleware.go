package observability

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
	maxSize    int
}

func newResponseWriterWrapper(w http.ResponseWriter, maxSize int) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		maxSize:        maxSize,
	}
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(p []byte) (int, error) {
	if rw.maxSize > 0 && rw.body.Len() < rw.maxSize {
		rem := rw.maxSize - rw.body.Len()
		if len(p) < rem {
			rw.body.Write(p)
		} else {
			rw.body.Write(p[:rem])
		}
	}
	// codeql[go/reflected-xss] False positive..this is a delegating ResponseWriter wrapper
	return rw.ResponseWriter.Write(p)
}

// HTTPMiddleware wraps an http.Handler with trace context, automated request/response logging, and latency recording.
//
// Usage example:
//
//	mux := http.NewServeMux()
//	loggedMux := observability.HTTPMiddleware(mux)
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceID := ExtractTraceID(req)
		baggage := ExtractBaggage(req)
		ctx := ContextWithBaggage(req.Context(), baggage)
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
	inst := getGlobal()
	var cfg Config
	if inst != nil {
		cfg = inst.Config
	}
	maxSize := getPayloadMaxSize(cfg)
	reqBody := captureReqBodyIfNeeded(req, cfg, maxSize)
	wrapper := newResponseWriterWrapper(w, maxSize)
	Info("incoming request started",
		NewStringField("method", req.Method),
		NewStringField("path", req.URL.Path),
	)
	next.ServeHTTP(wrapper, req)
	duration := time.Since(s.StartTime)
	Latency("http_request_duration", duration)
	Counter("http_requests_total")
	logFinishedRequest(req, wrapper, duration, cfg, reqBody)
}

func getPayloadMaxSize(cfg Config) int {
	if cfg.MaxPayloadLogSizeBytes <= 0 {
		return 4096
	}
	return cfg.MaxPayloadLogSizeBytes
}

func captureReqBodyIfNeeded(req *http.Request, cfg Config, maxSize int) string {
	if cfg.EnablePayloadLogging {
		return captureRequestBody(req, maxSize)
	}
	return ""
}

func captureRequestBody(req *http.Request, maxSize int) string {
	if req.Body == nil {
		return ""
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if len(bodyBytes) > maxSize {
		return string(bodyBytes[:maxSize])
	}
	return string(bodyBytes)
}

func shouldLogPayload(cfg Config, statusCode int) bool {
	if !cfg.EnablePayloadLogging {
		return false
	}
	if statusCode >= 500 {
		return true
	}
	return rand.Float64() < cfg.PayloadLoggingSample
}

func logFinishedRequest(req *http.Request, wrapper *responseWriterWrapper, duration time.Duration, cfg Config, reqBody string) {
	fields := []Field{
		NewStringField("method", req.Method),
		NewStringField("path", req.URL.Path),
		NewIntField("status", int64(wrapper.statusCode)),
		NewIntField("duration_ms", duration.Milliseconds()),
	}
	if shouldLogPayload(cfg, wrapper.statusCode) {
		fields = append(fields, NewStringField("request_payload", reqBody))
		fields = append(fields, NewStringField("response_payload", wrapper.body.String()))
	}
	Info("outgoing request finished", fields...)
}

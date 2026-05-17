package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"sync"
	"time"
)

// Tracer defines the interface for tracing.
type Tracer interface {
	StartTrace(ctx context.Context, name string) (context.Context, Span)
	EndTrace(span Span)
}

// LocalTracer implements Tracer locally.
type LocalTracer struct {
	writer  io.Writer
	service string
}

type contextKey string

const traceIDKey contextKey = "trace_id"

var (
	traceMu       sync.RWMutex
	activeTraceID string
)

// NewLocalTracer creates a new LocalTracer instance.
func NewLocalTracer(service string) *LocalTracer {
	tracer := &LocalTracer{
		writer:  os.Stdout,
		service: service,
	}
	return tracer
}

// SetWriter configures a custom output writer for testing or redirection.
func (t *LocalTracer) SetWriter(w io.Writer) {
	t.writer = w
	return
}

// SetActiveTraceID sets the current active trace ID thread-safely.
func SetActiveTraceID(id string) {
	traceMu.Lock()
	defer traceMu.Unlock()
	activeTraceID = id
}

// GetActiveTraceID retrieves the current active trace ID thread-safely.
func GetActiveTraceID() string {
	traceMu.RLock()
	defer traceMu.RUnlock()
	return activeTraceID
}

// StartTrace starts a new span and returns the context.
func (t *LocalTracer) StartTrace(ctx context.Context, name string) (context.Context, Span) {
	traceID := generateRandomHex(8)
	span := Span{
		TraceID:   traceID,
		Name:      name,
		StartTime: time.Now(),
	}
	SetActiveTraceID(traceID)
	newCtx := context.WithValue(ctx, traceIDKey, traceID)
	t.printStart(span)
	return newCtx, span
}

// EndTrace ends a span and prints details.
func (t *LocalTracer) EndTrace(span Span) {
	duration := time.Since(span.StartTime)
	line := "[TRACE] End " + span.Name + " duration=" + duration.String() + "\n"
	_, _ = io.WriteString(t.writer, line)
	SetActiveTraceID("")
}

// printStart outputs the trace start message to the writer.
func (t *LocalTracer) printStart(span Span) {
	line := "[TRACE] Start " + span.Name + " trace_id=" + span.TraceID + "\n"
	_, _ = io.WriteString(t.writer, line)
}

// generateRandomHex generates a cryptographically secure random hex string.
func generateRandomHex(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

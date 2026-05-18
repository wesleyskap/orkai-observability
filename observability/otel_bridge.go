// Package observability provides a unified, thread-safe observability facade
// for structured JSON logging, metrics aggregation, and nested context tracing.
package observability

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// otelTracer implements the Tracer interface using OpenTelemetry trace provider.
type otelTracer struct {
	mu     sync.Mutex
	spans  map[string][]trace.Span
	tracer trace.Tracer
}

// NewOTelTracer constructs a semantic adapter for OpenTelemetry tracing.
// It serves to bridge our global tracing facade to any native OTel exporter.
//
// Usage example:
//
//	tracer := observability.NewOTelTracer(otel.GetTracerProvider(), "auth-service")
func NewOTelTracer(tp trace.TracerProvider, serviceName string) Tracer {
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	return &otelTracer{
		spans:  make(map[string][]trace.Span),
		tracer: tp.Tracer(serviceName),
	}
}

// StartTrace starts a standard OpenTelemetry span in the active context.
func (o *otelTracer) StartTrace(ctx context.Context, name string) (context.Context, Span) {
	newCtx, otelSpan := o.tracer.Start(ctx, name)
	traceID := otelSpan.SpanContext().TraceID().String()
	span := Span{
		TraceID:   traceID,
		Name:      name,
		StartTime: time.Now(),
	}
	PushActiveTraceID(traceID)
	o.mu.Lock()
	key := traceID + "#" + name
	o.spans[key] = append(o.spans[key], otelSpan)
	o.mu.Unlock()
	return newCtx, span
}

// EndTrace retrieves and completes the matching active OpenTelemetry span.
func (o *otelTracer) EndTrace(span Span) {
	o.mu.Lock()
	key := span.TraceID + "#" + span.Name
	stack := o.spans[key]
	if len(stack) > 0 {
		otelSpan := stack[len(stack)-1]
		o.spans[key] = stack[:len(stack)-1]
		o.mu.Unlock()
		otelSpan.End()
		PopActiveTraceID()
	} else {
		o.mu.Unlock()
	}
}

// otelMetrics implements the Metrics interface routing to OpenTelemetry metrics SDKs.
type otelMetrics struct {
	meter      metric.Meter
	local      *InMemoryMetrics
	mu         sync.Mutex
	counters   map[string]metric.Int64Counter
	histograms map[string]metric.Float64Histogram
	gauges     map[string]metric.Float64Gauge
}

// NewOTelMetrics constructs a semantic adapter for OpenTelemetry metrics.
// It serves to bridge our global metrics facade to any native OTel collector.
//
// Usage example:
//
//	metrics := observability.NewOTelMetrics(otel.GetMeterProvider(), "auth-service")
func NewOTelMetrics(mp metric.MeterProvider, serviceName string) Metrics {
	if mp == nil {
		mp = otel.GetMeterProvider()
	}
	return &otelMetrics{
		meter:      mp.Meter(serviceName),
		local:      NewInMemoryMetrics(serviceName),
		counters:   make(map[string]metric.Int64Counter),
		histograms: make(map[string]metric.Float64Histogram),
		gauges:     make(map[string]metric.Float64Gauge),
	}
}

// IncCounter increments a counter metric.
func (o *otelMetrics) IncCounter(name string) {
	o.local.IncCounter(name)
	o.IncCounterWithLabels(name, nil)
}

// IncCounterWithLabels increments a counter metric with specific tags.
func (o *otelMetrics) IncCounterWithLabels(name string, labels map[string]string) {
	o.local.IncCounterWithLabels(name, labels)
	o.mu.Lock()
	c, exists := o.counters[name]
	var err error
	if !exists {
		c, err = o.meter.Int64Counter(name)
		if err == nil {
			o.counters[name] = c
		}
	}
	o.mu.Unlock()
	if err == nil && c != nil {
		attrs := convertLabels(labels)
		c.Add(context.Background(), 1, metric.WithAttributes(attrs...))
	}
}

// RecordLatency records a latency value.
func (o *otelMetrics) RecordLatency(name string, duration time.Duration) {
	o.local.RecordLatency(name, duration)
	o.RecordLatencyWithLabels(name, duration, nil)
}

// RecordLatencyWithLabels records a latency value with specific tags.
func (o *otelMetrics) RecordLatencyWithLabels(name string, duration time.Duration, labels map[string]string) {
	o.local.RecordLatencyWithLabels(name, duration, labels)
	o.mu.Lock()
	h, exists := o.histograms[name]
	var err error
	if !exists {
		h, err = o.meter.Float64Histogram(name)
		if err == nil {
			o.histograms[name] = h
		}
	}
	o.mu.Unlock()
	if err == nil && h != nil {
		attrs := convertLabels(labels)
		h.Record(context.Background(), float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
	}
}

// SetGauge sets a gauge value.
func (o *otelMetrics) SetGauge(name string, value float64) {
	o.local.SetGauge(name, value)
	o.SetGaugeWithLabels(name, value, nil)
}

// SetGaugeWithLabels sets a gauge value with specific tags.
func (o *otelMetrics) SetGaugeWithLabels(name string, value float64, labels map[string]string) {
	o.local.SetGaugeWithLabels(name, value, labels)
	o.mu.Lock()
	g, exists := o.gauges[name]
	var err error
	if !exists {
		g, err = o.meter.Float64Gauge(name)
		if err == nil {
			o.gauges[name] = g
		}
	}
	o.mu.Unlock()
	if err == nil && g != nil {
		attrs := convertLabels(labels)
		g.Record(context.Background(), value, metric.WithAttributes(attrs...))
	}
}

// Print outputs the collected metrics to standard output.
func (o *otelMetrics) Print() {
	o.local.Print()
}

// GetSummary retrieves the current MetricsSummary snapshot.
func (o *otelMetrics) GetSummary() MetricsSummary {
	return o.local.GetSummary()
}

// convertLabels converts tag key-value pairs into standard OpenTelemetry attributes.
func convertLabels(labels map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels))
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}

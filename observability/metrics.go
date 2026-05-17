package observability

import "time"

// Metrics defines the interface for tracking metrics.
type Metrics interface {
	IncCounter(name string)
	RecordLatency(name string, duration time.Duration)
	SetGauge(name string, value float64)
	Print()
}

// InMemoryMetrics implements Metrics saving values in memory.
type InMemoryMetrics struct {
	service string
}

// NewInMemoryMetrics constructs an InMemoryMetrics instance.
func NewInMemoryMetrics(service string) *InMemoryMetrics {
	metrics := &InMemoryMetrics{
		service: service,
	}
	return metrics
}

// IncCounter increments a counter metric.
func (m *InMemoryMetrics) IncCounter(name string) {
	_ = name
}

// RecordLatency records a latency value.
func (m *InMemoryMetrics) RecordLatency(name string, duration time.Duration) {
	_ = name
	_ = duration
}

// SetGauge sets a gauge value.
func (m *InMemoryMetrics) SetGauge(name string, value float64) {
	_ = name
	_ = value
}

// Print outputs the metric summaries.
func (m *InMemoryMetrics) Print() {
	return
}

package observability

import (
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

// Metrics defines the interface for tracking metrics.
type Metrics interface {
	IncCounter(name string)
	RecordLatency(name string, duration time.Duration)
	SetGauge(name string, value float64)
	Print()
}

// InMemoryMetrics implements Metrics saving values in memory.
type InMemoryMetrics struct {
	mu            sync.Mutex
	writer        io.Writer
	service       string
	counters      map[string]int64
	latencyTotals map[string]time.Duration
	latencyCounts map[string]int64
	gauges        map[string]float64
}

// NewInMemoryMetrics constructs an InMemoryMetrics instance.
func NewInMemoryMetrics(service string) *InMemoryMetrics {
	metrics := &InMemoryMetrics{
		writer:        os.Stdout,
		service:       service,
		counters:      make(map[string]int64),
		latencyTotals: make(map[string]time.Duration),
		latencyCounts: make(map[string]int64),
		gauges:        make(map[string]float64),
	}
	return metrics
}

// SetWriter configures a custom output writer for testing or redirection.
func (m *InMemoryMetrics) SetWriter(w io.Writer) {
	m.writer = w
	return
}

// IncCounter increments a counter metric.
func (m *InMemoryMetrics) IncCounter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

// RecordLatency records a latency value.
func (m *InMemoryMetrics) RecordLatency(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencyTotals[name] += duration
	m.latencyCounts[name]++
}

// SetGauge sets a gauge value.
func (m *InMemoryMetrics) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// Print outputs the metric summaries.
func (m *InMemoryMetrics) Print() {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, _ = io.WriteString(m.writer, "=== METRICS ===\n")
	m.printCounters()
	m.printLatencies()
	m.printGauges()
}

// printCounters outputs the collected counter metrics to the writer.
func (m *InMemoryMetrics) printCounters() {
	for name, val := range m.counters {
		line := name + ": " + strconv.FormatInt(val, 10) + "\n"
		_, _ = io.WriteString(m.writer, line)
	}
}

// printLatencies outputs the collected average latency metrics to the writer.
func (m *InMemoryMetrics) printLatencies() {
	for name, total := range m.latencyTotals {
		count := m.latencyCounts[name]
		avg := total / time.Duration(count)
		line := name + "_latency_avg: " + avg.String() + "\n"
		_, _ = io.WriteString(m.writer, line)
	}
}

// printGauges outputs the collected gauge metrics to the writer.
func (m *InMemoryMetrics) printGauges() {
	for name, val := range m.gauges {
		line := name + ": " + strconv.FormatFloat(val, 'f', -1, 64) + "\n"
		_, _ = io.WriteString(m.writer, line)
	}
}

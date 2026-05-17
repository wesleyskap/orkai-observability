package observability

import (
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricsSummary holds a snapshot copy of in-memory metrics.
type MetricsSummary struct {
	Counters  map[string]int64   `json:"counters"`
	Latencies map[string]float64 `json:"latencies"`
	Gauges    map[string]float64 `json:"gauges"`
}

// Metrics defines the interface for tracking metrics.
//
// Usage example:
//
//	var m observability.Metrics = observability.NewInMemoryMetrics("auth-service")
type Metrics interface {
	IncCounter(name string)
	IncCounterWithLabels(name string, labels map[string]string)
	RecordLatency(name string, duration time.Duration)
	RecordLatencyWithLabels(name string, duration time.Duration, labels map[string]string)
	SetGauge(name string, value float64)
	SetGaugeWithLabels(name string, value float64, labels map[string]string)
	Print()
	GetSummary() MetricsSummary
}

// InMemoryMetrics implements Metrics saving values in memory.
//
// Usage example:
//
//	metrics := observability.NewInMemoryMetrics("user-service")
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
//
// Usage example:
//
//	metrics := observability.NewInMemoryMetrics("auth-service")
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
//
// Usage example:
//
//	metrics.SetWriter(buf)
func (m *InMemoryMetrics) SetWriter(w io.Writer) {
	m.writer = w
	return
}

// IncCounter increments a counter metric.
//
// Usage example:
//
//	metrics.IncCounter("http_requests_total")
func (m *InMemoryMetrics) IncCounter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

// RecordLatency records a latency value.
//
// Usage example:
//
//	metrics.RecordLatency("db_query_duration", 15*time.Millisecond)
func (m *InMemoryMetrics) RecordLatency(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencyTotals[name] += duration
	m.latencyCounts[name]++
}

// SetGauge sets a gauge value.
//
// Usage example:
//
//	metrics.SetGauge("memory_usage_bytes", 52428800)
func (m *InMemoryMetrics) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// Print outputs the metric summaries.
//
// Usage example:
//
//	metrics.Print()
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

// GetSummary returns a snapshot copy of all collected metrics.
//
// Usage example:
//
//	summary := metrics.GetSummary()
func (m *InMemoryMetrics) GetSummary() MetricsSummary {
	m.mu.Lock()
	defer m.mu.Unlock()
	summary := MetricsSummary{
		Counters:  make(map[string]int64),
		Latencies: make(map[string]float64),
		Gauges:    make(map[string]float64),
	}
	for k, v := range m.counters {
		summary.Counters[k] = v
	}
	for k, v := range m.gauges {
		summary.Gauges[k] = v
	}
	for k, v := range m.latencyTotals {
		count := m.latencyCounts[k]
		if count > 0 {
			summary.Latencies[k] = float64(v.Milliseconds()) / float64(count)
		}
	}
	return summary
}

// formatMetricKey generates a deterministic, sorted string representation of name and labels.
func formatMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + `="` + labels[k] + `"`
	}
	return name + "{" + strings.Join(parts, ",") + "}"
}

// IncCounterWithLabels increments a counter with labels.
func (m *InMemoryMetrics) IncCounterWithLabels(name string, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.IncCounter(key)
}

// RecordLatencyWithLabels records latency with labels.
func (m *InMemoryMetrics) RecordLatencyWithLabels(name string, duration time.Duration, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.RecordLatency(key, duration)
}

// SetGaugeWithLabels sets gauge with labels.
func (m *InMemoryMetrics) SetGaugeWithLabels(name string, value float64, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.SetGauge(key, value)
}

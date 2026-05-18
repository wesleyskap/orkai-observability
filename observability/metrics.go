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

// LatencyPercentiles holds computed p50, p90, and p99 percentile values in milliseconds.
// It serves to diagnose long-tail latency outliers in production environments.
//
// Usage example:
//
//	pct := summary.Percentiles["http_requests"]
//	medianMs := pct.P50
type LatencyPercentiles struct {
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	P99 float64 `json:"p99"`
}

// MetricsSummary holds a snapshot copy of in-memory metrics, carrying counters, latencies,
// percentiles, cumulative histograms, and gauges. It serves to convey formatted JSON outputs.
//
// Usage example:
//
//	summary := metrics.GetSummary()
//	totalHits := summary.Counters["http_requests_total"]
type MetricsSummary struct {
	Counters    map[string]int64              `json:"counters"`
	Latencies   map[string]float64            `json:"latencies"`
	Percentiles map[string]LatencyPercentiles `json:"percentiles"`
	Histograms  map[string]map[string]int64   `json:"histograms"`
	Gauges      map[string]float64            `json:"gauges"`
}

// latencyReservoir is a thread-safe sliding-window list of latency observations.
type latencyReservoir struct {
	mu      sync.Mutex
	samples []float64
	maxSize int
}

// newLatencyReservoir initializes a bounded latency reservoir.
func newLatencyReservoir(maxSize int) *latencyReservoir {
	return &latencyReservoir{
		samples: make([]float64, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new latency observation to the reservoir.
func (r *latencyReservoir) Add(val float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.samples) >= r.maxSize {
		r.samples = append(r.samples[1:], val)
	} else {
		r.samples = append(r.samples, val)
	}
}

// extractPercentile computes a single percentile from a sorted sample slice.
func (r *latencyReservoir) extractPercentile(sorted []float64, percentile float64) float64 {
	n := len(sorted)
	idx := int(float64(n) * percentile)
	if idx < n {
		return sorted[idx]
	}
	return 0
}

// Percentiles calculates the p50, p90, and p99 values from the reservoir.
func (r *latencyReservoir) Percentiles() (p50, p90, p99 float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := len(r.samples)
	if n == 0 {
		return 0, 0, 0
	}
	temp := make([]float64, n)
	copy(temp, r.samples)
	sort.Float64s(temp)
	p50 = r.extractPercentile(temp, 0.50)
	p90 = r.extractPercentile(temp, 0.90)
	p99 = r.extractPercentile(temp, 0.99)
	return
}

// incrementBucketCounters increments cumulative counts for matching buckets.
func (r *latencyReservoir) incrementBucketCounters(buckets map[string]int64, thresholds []float64, val float64) {
	for _, limit := range thresholds {
		if val <= limit {
			limitStr := strconv.FormatFloat(limit, 'f', -1, 64)
			buckets[limitStr]++
		}
	}
	buckets["+Inf"]++
}

// Buckets generates cumulative bucket counts matching Prometheus expectations.
func (r *latencyReservoir) Buckets() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	buckets := make(map[string]int64)
	thresholds := []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}
	for _, limit := range thresholds {
		limitStr := strconv.FormatFloat(limit, 'f', -1, 64)
		buckets[limitStr] = 0
	}
	buckets["+Inf"] = 0
	for _, val := range r.samples {
		r.incrementBucketCounters(buckets, thresholds, val)
	}
	return buckets
}

// Metrics defines the interface for tracking metrics.
//
// Usage example:
//
//	var m observability.Metrics = observability.NewInMemoryMetrics("auth-service")
type Metrics interface {
	// IncCounter increments a counter metric.
	//
	// Usage example:
	//
	//	metrics.IncCounter("http_requests_total")
	IncCounter(name string)

	// IncCounterWithLabels increments a counter metric with specific tags.
	//
	// Usage example:
	//
	//	metrics.IncCounterWithLabels("http_requests_total", map[string]string{"method": "POST"})
	IncCounterWithLabels(name string, labels map[string]string)

	// RecordLatency records a latency value.
	//
	// Usage example:
	//
	//	metrics.RecordLatency("db_query_duration", 15*time.Millisecond)
	RecordLatency(name string, duration time.Duration)

	// RecordLatencyWithLabels records a latency value with specific tags.
	//
	// Usage example:
	//
	//	metrics.RecordLatencyWithLabels("db_query", 10*time.Millisecond, map[string]string{"op": "select"})
	RecordLatencyWithLabels(name string, duration time.Duration, labels map[string]string)

	// SetGauge sets a gauge value.
	//
	// Usage example:
	//
	//	metrics.SetGauge("memory_usage_bytes", 52428800)
	SetGauge(name string, value float64)

	// SetGaugeWithLabels sets a gauge value with specific tags.
	//
	// Usage example:
	//
	//	metrics.SetGaugeWithLabels("cpu_usage", 85.5, map[string]string{"core": "0"})
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
	mu                sync.Mutex
	writer            io.Writer
	service           string
	counters          map[string]int64
	latencyTotals     map[string]time.Duration
	latencyCounts     map[string]int64
	latencyReservoirs map[string]*latencyReservoir
	gauges            map[string]float64
}

// NewInMemoryMetrics constructs an InMemoryMetrics instance.
//
// Usage example:
//
//	metrics := observability.NewInMemoryMetrics("auth-service")
func NewInMemoryMetrics(service string) *InMemoryMetrics {
	metrics := &InMemoryMetrics{
		writer:            os.Stdout,
		service:           service,
		counters:          make(map[string]int64),
		latencyTotals:     make(map[string]time.Duration),
		latencyCounts:     make(map[string]int64),
		latencyReservoirs: make(map[string]*latencyReservoir),
		gauges:            make(map[string]float64),
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
	m.addSampleToReservoir(name, duration)
}

// addSampleToReservoir initializes and populates the latency reservoir safely.
func (m *InMemoryMetrics) addSampleToReservoir(name string, duration time.Duration) {
	r, exists := m.latencyReservoirs[name]
	if !exists {
		r = newLatencyReservoir(2000)
		m.latencyReservoirs[name] = r
	}
	r.Add(float64(duration.Milliseconds()))
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
	summary := m.initEmptySummary()
	m.copyCountersToSummary(summary)
	m.copyGaugesToSummary(summary)
	m.copyLatenciesToSummary(summary)
	m.copyReservoirsToSummary(summary)
	return summary
}

// initEmptySummary initializes the standard empty MetricsSummary with allocated maps.
func (m *InMemoryMetrics) initEmptySummary() MetricsSummary {
	return MetricsSummary{
		Counters:    make(map[string]int64),
		Latencies:   make(map[string]float64),
		Percentiles: make(map[string]LatencyPercentiles),
		Histograms:  make(map[string]map[string]int64),
		Gauges:      make(map[string]float64),
	}
}

// copyCountersToSummary copies in-memory counters into the summary.
func (m *InMemoryMetrics) copyCountersToSummary(s MetricsSummary) {
	for k, v := range m.counters {
		s.Counters[k] = v
	}
}

// copyGaugesToSummary copies in-memory gauges into the summary.
func (m *InMemoryMetrics) copyGaugesToSummary(s MetricsSummary) {
	for k, v := range m.gauges {
		s.Gauges[k] = v
	}
}

// copyLatenciesToSummary copies in-memory average latency estimates into the summary.
func (m *InMemoryMetrics) copyLatenciesToSummary(s MetricsSummary) {
	for k, v := range m.latencyTotals {
		count := m.latencyCounts[k]
		if count > 0 {
			s.Latencies[k] = float64(v.Milliseconds()) / float64(count)
		}
	}
}

// copyReservoirsToSummary computes percentiles and buckets to populate the summary.
func (m *InMemoryMetrics) copyReservoirsToSummary(s MetricsSummary) {
	for k, r := range m.latencyReservoirs {
		p50, p90, p99 := r.Percentiles()
		s.Percentiles[k] = LatencyPercentiles{
			P50: p50,
			P90: p90,
			P99: p99,
		}
		s.Histograms[k] = r.Buckets()
	}
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
//
// Usage example:
//
//	metrics.IncCounterWithLabels("http_requests_total", map[string]string{"method": "POST"})
func (m *InMemoryMetrics) IncCounterWithLabels(name string, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.IncCounter(key)
}

// RecordLatencyWithLabels records latency with labels.
//
// Usage example:
//
//	metrics.RecordLatencyWithLabels("db_query", 10*time.Millisecond, map[string]string{"op": "select"})
func (m *InMemoryMetrics) RecordLatencyWithLabels(name string, duration time.Duration, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.RecordLatency(key, duration)
}

// SetGaugeWithLabels sets gauge with labels.
//
// Usage example:
//
//	metrics.SetGaugeWithLabels("cpu_usage", 85.5, map[string]string{"core": "0"})
func (m *InMemoryMetrics) SetGaugeWithLabels(name string, value float64, labels map[string]string) {
	key := formatMetricKey(name, labels)
	m.SetGauge(key, value)
}

package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MetricsHTTPHandler returns an http.HandlerFunc that exports metrics in either Prometheus or JSON format.
//
// Usage example:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/metrics", observability.MetricsHTTPHandler())
func MetricsHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if globalInstance == nil {
			http.Error(w, "observability package is not initialized", http.StatusServiceUnavailable)
			return
		}
		summary := globalInstance.Metrics.GetSummary()
		if requestsPrometheus(req) {
			writePrometheusMetrics(w, summary)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(summary)
	}
}

// requestsPrometheus checks if the client requests the standard Prometheus text format.
func requestsPrometheus(req *http.Request) bool {
	accept := req.Header.Get("Accept")
	format := req.URL.Query().Get("format")
	return format == "prometheus" || strings.Contains(accept, "text/plain")
}

// writePrometheusMetrics writes the metrics in the official Prometheus text format.
func writePrometheusMetrics(w http.ResponseWriter, summary MetricsSummary) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	writeCounters(w, summary.Counters)
	writeGauges(w, summary.Gauges)
	writeLatencies(w, summary.Latencies)
}

// writeCounters formats and writes all cumulative counters.
func writeCounters(w io.Writer, counters map[string]int64) {
	for name, val := range counters {
		base, labels := parseMetricKey(name)
		_, _ = fmt.Fprintf(w, "# HELP %s Cumulative counter of %s\n", base, base)
		_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", base)
		_, _ = fmt.Fprintf(w, "%s%s %d\n", base, labels, val)
	}
}

// writeGauges formats and writes all decimal gauges.
func writeGauges(w io.Writer, gauges map[string]float64) {
	for name, val := range gauges {
		base, labels := parseMetricKey(name)
		_, _ = fmt.Fprintf(w, "# HELP %s Current %s\n", base, base)
		_, _ = fmt.Fprintf(w, "# TYPE %s gauge\n", base)
		_, _ = fmt.Fprintf(w, "%s%s %g\n", base, labels, val)
	}
}

// writeLatencies formats and writes all average latency metrics in milliseconds.
func writeLatencies(w io.Writer, latencies map[string]float64) {
	for name, val := range latencies {
		base, labels := parseMetricKey(name)
		fullName := base + "_latency_avg"
		_, _ = fmt.Fprintf(w, "# HELP %s Average latency of %s in milliseconds\n", fullName, base)
		_, _ = fmt.Fprintf(w, "# TYPE %s gauge\n", fullName)
		_, _ = fmt.Fprintf(w, "%s%s %g\n", fullName, labels, val)
	}
}

// parseMetricKey splits a formatted metric name into the base name and labels block.
func parseMetricKey(key string) (string, string) {
	if idx := strings.Index(key, "{"); idx != -1 {
		return key[:idx], key[idx:]
	}
	return key, ""
}

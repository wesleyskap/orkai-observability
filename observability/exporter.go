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
		_, _ = fmt.Fprintf(w, "# HELP %s Cumulative counter of %s\n", name, name)
		_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", name)
		_, _ = fmt.Fprintf(w, "%s %d\n", name, val)
	}
}

// writeGauges formats and writes all decimal gauges.
func writeGauges(w io.Writer, gauges map[string]float64) {
	for name, val := range gauges {
		_, _ = fmt.Fprintf(w, "# HELP %s Current %s\n", name, name)
		_, _ = fmt.Fprintf(w, "# TYPE %s gauge\n", name)
		_, _ = fmt.Fprintf(w, "%s %g\n", name, val)
	}
}

// writeLatencies formats and writes all average latency metrics in milliseconds.
func writeLatencies(w io.Writer, latencies map[string]float64) {
	for name, val := range latencies {
		fullName := name + "_latency_avg"
		_, _ = fmt.Fprintf(w, "# HELP %s Average latency of %s in milliseconds\n", fullName, name)
		_, _ = fmt.Fprintf(w, "# TYPE %s gauge\n", fullName)
		_, _ = fmt.Fprintf(w, "%s %g\n", fullName, val)
	}
}

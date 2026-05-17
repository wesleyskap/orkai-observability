package observability

import (
	"encoding/json"
	"net/http"
)

// MetricsHTTPHandler returns an http.HandlerFunc that exports the active metrics snapshot in JSON format.
//
// Usage example:
//	mux := http.NewServeMux()
//	mux.HandleFunc("/metrics", observability.MetricsHTTPHandler())
func MetricsHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if globalInstance == nil {
			http.Error(w, "observability package is not initialized", http.StatusServiceUnavailable)
			return
		}
		summary := globalInstance.Metrics.GetSummary()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(summary)
	}
}

package observability

import (
	"fmt"
	"net/http"
)

// PanicRecoveryMiddleware recovers from handler panics, logging stack trace and recording metrics.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer runRecovery(w, req)
		next.ServeHTTP(w, req)
	})
}

func runRecovery(w http.ResponseWriter, req *http.Request) {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic: %v", r)
		ErrorContext(req.Context(), "panic in http handler", err)
		Counter("http_panics_total")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
	}
}

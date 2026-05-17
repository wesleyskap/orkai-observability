package observability

import (
	"net/http"
	"strings"
)

// ExtractTraceID resolves trace identifier from HTTP headers supporting W3C, B3, and X-Trace-ID formats.
//
// Usage example:
//
//	traceID := observability.ExtractTraceID(req)
func ExtractTraceID(req *http.Request) string {
	if tp := req.Header.Get("traceparent"); tp != "" {
		if parts := strings.Split(tp, "-"); len(parts) >= 3 && len(parts[1]) == 32 {
			return parts[1]
		}
	}
	if b3 := req.Header.Get("b3"); b3 != "" {
		if idx := strings.Index(b3, "-"); idx != -1 {
			return b3[:idx]
		}
		return b3
	}
	if b3Trace := req.Header.Get("X-B3-TraceId"); b3Trace != "" {
		return b3Trace
	}
	return req.Header.Get("X-Trace-ID")
}

// InjectTraceID injects trace parent identifiers into HTTP headers for outbound service propagation.
//
// Usage example:
//
//	observability.InjectTraceID(req, "4bf92f3577b34da6a3ce929d0e0e4736")
func InjectTraceID(req *http.Request, traceID string) {
	if traceID == "" {
		return
	}
	spanID := "0000000000000000"
	if len(traceID) > 16 {
		spanID = traceID[:16]
	}
	req.Header.Set("traceparent", "00-"+traceID+"-"+spanID+"-01")
	req.Header.Set("b3", traceID+"-"+spanID+"-1")
	req.Header.Set("X-Trace-ID", traceID)
}

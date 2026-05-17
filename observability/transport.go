package observability

import (
	"net/http"
)

// TracingRoundTripper intercepts outbound HTTP requests injecting active trace headers.
type TracingRoundTripper struct {
	next http.RoundTripper
}

// NewTracingRoundTripper creates an HTTP transport wrapper injecting parent span trace IDs.
//
// Usage example:
//
//	transport := observability.NewTracingRoundTripper(http.DefaultTransport)
func NewTracingRoundTripper(next http.RoundTripper) *TracingRoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &TracingRoundTripper{next: next}
}

// RoundTrip executes standard request operations injecting active trace headers dynamically.
//
// Usage example:
//
//	resp, err := client.Transport.RoundTrip(req)
func (rt *TracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	activeID := GetActiveTraceID()
	if activeID != "" {
		reqCopy := req.Clone(req.Context())
		reqCopy.Header.Set("X-Trace-ID", activeID)
		return rt.next.RoundTrip(reqCopy)
	}
	return rt.next.RoundTrip(req)
}

// NewTracingClient constructs a custom standard HTTP client carrying tracing wrappers.
//
// Usage example:
//
//	client := observability.NewTracingClient()
func NewTracingClient() *http.Client {
	return &http.Client{
		Transport: NewTracingRoundTripper(nil),
	}
}

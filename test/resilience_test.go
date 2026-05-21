package test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// FakeHTTPTransport mocks http.RoundTripper for resilience testing.
type FakeHTTPTransport struct {
	Attempts int
	Fail     bool
	FailErr  error
	Status   int
}

// RoundTrip simulates HTTP request execution capturing attempts.
func (f *FakeHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.Attempts++
	if f.FailErr != nil {
		return nil, f.FailErr
	}
	if f.Fail {
		return nil, errors.New("network error")
	}
	resp := &http.Response{StatusCode: f.Status}
	return resp, nil
}

// TestCircuitBreakerTransitionsToOpen verifies that a circuit trips to OPEN state under errors.
func TestCircuitBreakerTransitionsToOpen(t *testing.T) {
	cb := observability.NewCircuitBreaker(0.5, 3, 50*time.Millisecond)
	if cb.State() != observability.StateClosed {
		t.Fatalf("expected initial state CLOSED, got %v", cb.State())
	}
	for i := 0; i < 3; i++ {
		cb.RecordResult(errors.New("failure"))
	}
	if cb.State() != observability.StateOpen {
		t.Fatalf("expected tripped state OPEN, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatalf("expected Allow to return false when open")
	}
}

// TestCircuitBreakerTransitionsToClosed verifies recovery from OPEN to HALF-OPEN to CLOSED.
func TestCircuitBreakerTransitionsToClosed(t *testing.T) {
	cb := observability.NewCircuitBreaker(0.5, 1, 10*time.Millisecond)
	cb.RecordResult(errors.New("failure"))
	time.Sleep(15 * time.Millisecond)
	if !cb.Allow() {
		t.Fatalf("expected Allow to return true after cooldown")
	}
	if cb.State() != observability.StateHalfOpen {
		t.Fatalf("expected state HALF-OPEN, got %v", cb.State())
	}
	cb.RecordResult(nil)
	if cb.State() != observability.StateClosed {
		t.Fatalf("expected state CLOSED after success, got %v", cb.State())
	}
}

// TestResilientRoundTripperCircuitTrip asserts that transport failures trip the circuit.
func TestResilientRoundTripperCircuitTrip(t *testing.T) {
	fake := &FakeHTTPTransport{Fail: true}
	cb := observability.NewCircuitBreaker(0.5, 2, 50*time.Millisecond)
	rt := observability.NewResilientRoundTripper(fake, cb, 0, 0)
	req, _ := http.NewRequest("GET", "http://local", nil)
	_, _ = rt.RoundTrip(req)
	_, _ = rt.RoundTrip(req)
	if cb.State() != observability.StateOpen {
		t.Fatalf("expected circuit to be tripped to OPEN, got %v", cb.State())
	}
	_, err := rt.RoundTrip(req)
	if err != observability.ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

// TestResilientRoundTripperExponentialRetry asserts that transient failures trigger retries.
func TestResilientRoundTripperExponentialRetry(t *testing.T) {
	fake := &FakeHTTPTransport{Status: http.StatusServiceUnavailable}
	cb := observability.NewCircuitBreaker(0.5, 10, 50*time.Millisecond)
	rt := observability.NewResilientRoundTripper(fake, cb, 2, 1*time.Millisecond)
	req, _ := http.NewRequest("GET", "http://local", nil)
	_, _ = rt.RoundTrip(req)
	if fake.Attempts != 3 {
		t.Fatalf("expected exactly 3 attempts (1 initial + 2 retries), got %d", fake.Attempts)
	}
}

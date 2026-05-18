package observability

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// CircuitState represents the current state of a CircuitBreaker.
type CircuitState int32

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// ErrCircuitOpen is returned when the circuit breaker is in OPEN state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker acts as a protective state machine for outbound operations.
//
// Usage example:
//
//	cb := observability.NewCircuitBreaker(0.5, 5, 10*time.Second)
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failures         int
	consecutive      int
	total            int
	lastChange       time.Time
	thresholdRatio   float64
	consecutiveLimit int
	cooldown         time.Duration
}

// NewCircuitBreaker creates a configured thread-safe CircuitBreaker instance.
//
// Usage example:
//
//	cb := observability.NewCircuitBreaker(0.5, 5, 10*time.Second)
func NewCircuitBreaker(ratio float64, consecutive int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		thresholdRatio:   ratio,
		consecutiveLimit: consecutive,
		cooldown:         cooldown,
		lastChange:       time.Now(),
	}
}

// Allow checks if a request is permitted by the circuit breaker state.
//
// Usage example:
//
//	if !cb.Allow() { return ErrCircuitOpen }
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateOpen {
		if time.Since(cb.lastChange) > cb.cooldown {
			cb.state = StateHalfOpen
			cb.lastChange = time.Now()
			return true
		}
		return false
	}
	return true
}

// RecordResult updates the circuit breaker state based on the outcome of the request.
//
// Usage example:
//
//	cb.RecordResult(err)
func (cb *CircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.total++
	if err != nil {
		cb.failures++
		cb.consecutive++
		cb.evaluateFailure()
	} else {
		cb.evaluateSuccess()
	}
}

func (cb *CircuitBreaker) evaluateFailure() {
	if cb.state == StateHalfOpen {
		cb.trip()
		return
	}
	ratio := float64(cb.failures) / float64(cb.total)
	if cb.consecutive >= cb.consecutiveLimit || (cb.total >= 10 && ratio >= cb.thresholdRatio) {
		cb.trip()
	}
}

func (cb *CircuitBreaker) evaluateSuccess() {
	cb.consecutive = 0
	if cb.state == StateHalfOpen {
		cb.reset()
	}
}

func (cb *CircuitBreaker) trip() {
	cb.state = StateOpen
	cb.lastChange = time.Now()
}

func (cb *CircuitBreaker) reset() {
	cb.state = StateClosed
	cb.failures = 0
	cb.consecutive = 0
	cb.total = 0
	cb.lastChange = time.Now()
}

// State returns the current state of the circuit breaker.
//
// Usage example:
//
//	state := cb.State()
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ResilientRoundTripper wraps an http.RoundTripper adding CircuitBreaker and Retry logic.
//
// Usage example:
//
//	rt := observability.NewResilientRoundTripper(nil, cb, 3, 100*time.Millisecond)
type ResilientRoundTripper struct {
	next    http.RoundTripper
	breaker *CircuitBreaker
	retries int
	backoff time.Duration
}

// NewResilientRoundTripper constructs a resilient HTTP transport wrapping a CircuitBreaker.
//
// Usage example:
//
//	rt := observability.NewResilientRoundTripper(nil, cb, 3, 100*time.Millisecond)
func NewResilientRoundTripper(next http.RoundTripper, cb *CircuitBreaker, retries int, backoff time.Duration) *ResilientRoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &ResilientRoundTripper{
		next:    next,
		breaker: cb,
		retries: retries,
		backoff: backoff,
	}
}

// RoundTrip executes requests through a CircuitBreaker protecting outbound network calls.
//
// Usage example:
//
//	resp, err := rt.RoundTrip(req)
func (rt *ResilientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !rt.breaker.Allow() {
		return nil, ErrCircuitOpen
	}
	resp, err := rt.executeWithRetry(req)
	rt.breaker.RecordResult(rt.checkOutcome(resp, err))
	return resp, err
}

func (rt *ResilientRoundTripper) checkOutcome(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp != nil && (resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout) {
		return ErrCircuitOpen
	}
	return nil
}

func (rt *ResilientRoundTripper) executeWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; attempt <= rt.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(rt.calculateBackoff(attempt))
		}
		resp, err = rt.next.RoundTrip(req)
		if !rt.isTransientError(resp, err) {
			break
		}
	}
	return resp, err
}

func (rt *ResilientRoundTripper) calculateBackoff(attempt int) time.Duration {
	factor := int64(1 << uint(attempt-1))
	return rt.backoff * time.Duration(factor)
}

func (rt *ResilientRoundTripper) isTransientError(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp != nil && (resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout) {
		return true
	}
	return false
}

package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

// LogRateLimiter implements a thread-safe token-bucket rate limiter for log sampling.
//
// Usage example:
//
//	limiter := observability.NewLogRateLimiter(100, 50, true)
type LogRateLimiter struct {
	mu           sync.Mutex
	burst        int
	rate         float64
	tokens       float64
	lastRefill   time.Time
	skippedCount uint64
	enabled      bool
}

// NewLogRateLimiter constructs a configured token-bucket logger rate limiter.
//
// Usage example:
//
//	limiter := observability.NewLogRateLimiter(100, 50, true)
func NewLogRateLimiter(burst int, rate int, enabled bool) *LogRateLimiter {
	return &LogRateLimiter{
		burst:      burst,
		rate:       float64(rate),
		tokens:     float64(burst),
		lastRefill: time.Now(),
		enabled:    enabled,
	}
}

// Allow determines if a log entry is permitted under burst limits or should be sampled.
//
// Usage example:
//
//	allow, throttled := limiter.Allow()
func (l *LogRateLimiter) Allow() (bool, bool) {
	if !l.enabled {
		return true, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true, false
	}
	skipped := atomic.AddUint64(&l.skippedCount, 1)
	if skipped%10 == 0 {
		return true, true
	}
	return false, false
}

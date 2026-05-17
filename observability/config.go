// Package observability provides a unified, thread-safe observability facade
// for structured JSON logging, metrics aggregation, and nested context tracing.
package observability

import "errors"

// Config defines the configuration settings for the observability package.
//
// Usage example:
//
//	cfg := observability.Config{
//		ServiceName:     "user-service",
//		Environment:     "production",
//		LogLevel:        "info",
//		EnableRateLimit: true,
//		RateLimitBurst:  100,
//		RateLimitRate:   50,
//	}
type Config struct {
	ServiceName     string
	Environment     string
	LogLevel        string
	EnableRateLimit bool
	RateLimitBurst  int
	RateLimitRate   int
}

// ValidateConfig verifies that the configuration fields are not empty.
//
// Usage example:
//
//	err := observability.ValidateConfig(cfg)
func ValidateConfig(cfg Config) error {
	if cfg.ServiceName == "" {
		return errors.New("ServiceName cannot be empty")
	}
	if cfg.Environment == "" {
		return errors.New("Environment cannot be empty")
	}
	return nil
}

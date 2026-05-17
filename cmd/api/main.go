package main

import (
	"context"
	"orkai-observability/observability"
	"time"
)

// main is the entrypoint simulating server requests and showing logs/metrics.
func main() {
	cfg := observability.Config{
		ServiceName: "auth-service",
		Environment: "dev",
		LogLevel:    "info",
	}
	_ = observability.Init(cfg)
	simulateRequest()
	observability.Dump()
}

// simulateRequest simulates an HTTP request handler flow with logs and metrics.
func simulateRequest() {
	start := time.Now()
	ctx, span := observability.StartSpan(context.Background(), "LoginHandler")
	defer observability.EndSpan(span)
	observability.Info("login request received")
	mockDatabaseCall(ctx)
	observability.Info("user authenticated successfully", observability.NewStringField("role", "admin"))
	observability.Counter("login_requests_total")
	observability.Latency("login_duration", time.Since(start))
}

// mockDatabaseCall simulates a nested database query within the current trace.
func mockDatabaseCall(ctx context.Context) {
	_, span := observability.StartSpan(ctx, "DatabaseQuery")
	defer observability.EndSpan(span)
	observability.Info("executing select user query", observability.NewStringField("table", "users"))
	time.Sleep(15 * time.Millisecond)
	observability.Counter("db_queries_total")
}

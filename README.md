[![Go Reference](https://pkg.go.dev/badge/github.com/wesleyskap/orkai-observability.svg)](https://pkg.go.dev/github.com/wesleyskap/orkai-observability)     [![Go Report Card](https://goreportcard.com/badge/github.com/wesleyskap/orkai-observability)](https://goreportcard.com/report/github.com/wesleyskap/orkai-observability)


# Orkai Observability

A modern, high-performance, lightweight observability package for Go backend services. It provides correlated structured JSON logging, thread-safe metrics collection, and LIFO nested parent-child trace spans under a single, unified facade interface.

---

## Package Integration Sequence

Step-by-step lifecycle flow of an incoming HTTP request executing nested operations (like database calls) inside a Go application using our package, highlighting how logs dynamically resolve trace IDs from the LIFO tracking stack.

```mermaid
sequenceDiagram
    autonumber
    actor Client as Client Request
    participant Middleware as HTTP Middleware
    participant App as Your Go Application (Handler)
    box rgba(255, 255, 255, 1) observability Facade
        participant Facade as global facade
        participant Logger as JSON Logger
        participant Tracer as Local Tracer (LIFO Stack)
        participant Metrics as InMemory Metrics
    end
    participant Console as Stdout / Console Output

    Client->>Middleware: Initiates HTTP request
    Note over Middleware, Tracer: Automated request intercepting and span bootstrap
    Middleware->>Facade: StartSpan("/users")
    Facade->>Tracer: Push parent Trace ID ("db3bda")
    Tracer->>Console: Prints: [TRACE] Start /users trace_id=db3bda
    
    Middleware->>App: Executes Handler (next.ServeHTTP)
    App->>Facade: Info("processing signup")
    Facade->>Logger: Write log (requests active trace ID)
    Logger->>Tracer: GetActiveTraceID()
    Tracer-->>Logger: Returns "db3bda"
    Logger->>Console: Output JSON: {"trace_id":"db3bda","msg":"processing signup"}

    Note over App, Tracer: Nested Span Execution (e.g. Database Call)
    App->>Facade: StartSpan("DatabaseQuery")
    Facade->>Tracer: Push nested Trace ID ("1b1ff7")
    Tracer->>Console: Prints: [TRACE] Start DatabaseQuery trace_id=1b1ff7
    App->>Facade: Info("inserting user account")
    Facade->>Logger: Write log (requests active trace ID)
    Logger->>Tracer: GetActiveTraceID()
    Tracer-->>Logger: Returns "1b1ff7"
    Logger->>Console: Output JSON: {"trace_id":"1b1ff7","msg":"inserting user account"}
    App->>Facade: EndSpan(DBQuerySpan)
    Facade->>Tracer: Pop nested Trace ID ("1b1ff7")
    Tracer->>Console: Prints: [TRACE] End DatabaseQuery duration=15ms

    Note over App, Tracer: Restored Parent Context
    App-->>Middleware: Returns Response (status 201)

    Note over Middleware, Metrics: Automated request completion logging & metrics
    Middleware->>Facade: Counter("request_count")
    Facade->>Metrics: Record counter increment
    Middleware->>Facade: EndSpan(HandlerSpan)
    Facade->>Tracer: Pop parent Trace ID ("db3bda")
    Tracer->>Console: Prints: [TRACE] End /users duration=15ms
```

### How It Works

1. **Automatic Trace Correlation:** When an incoming request starts, a unique Trace ID (e.g., `db3bda`) is generated and placed on a **LIFO (Last-In, First-Out) stack**. Think of this stack like a pile of plates: you add new IDs to the top, and always read or remove the top plate first.
2. **Context-Aware Logging:** Whenever you write a log (like calling `Info(...)`), the Logger automatically looks at the stack to grab the active ID currently on top. This correlates your logs without requiring you to manually pass Trace IDs as parameters to every single log function.
3. **Seamless Nesting (Sub-Traces):** If your code performs a nested operation (such as querying a database or calling another service), starting a new trace span generates a new child ID (e.g., `1b1ff7`) and pushes it onto the stack. Any logs written during the database call will automatically carry this new child ID.
4. **Self-Restoring Parent Context:** Once the database call finishes and its span ends, the child ID is popped off the stack. This immediately restores the parent's Trace ID (`db3bda`) back to the top of the stack. All subsequent logs written in the main handler will seamlessly carry the parent ID again.
5. **Thread-Safe Metrics:** Application-level events (counters, latencies, gauges) are recorded concurrently in-memory, completely protected against race conditions, and dumped into a formatted terminal report on demand.

---

## Features

* **Ultra-Fast JSON Logger:** A custom, reflection-free, structured logger that outputs directly to standard output or any custom `io.Writer`.
* **LIFO Nesting Traces:** An advanced, thread-safe LIFO trace stack that propagates unique cryptographically secure hex trace IDs. Sub-traces (e.g., DB queries) automatically nested inside parent spans correctly pop to restore the parent's active trace context on completion.
* **Thread-Safe Metrics:** In-memory tracking for cumulative counters, arithmetic average latencies over multi-sample periods, and decimal gauges—all protected under concurrent mutex locks.
* **Unified Facade API:** Clean, package-level package functions (`Info`, `Counter`, `StartSpan`) that automatically coordinate dynamic trace-log correlations seamlessly.
* **HTTP Tracing Middleware:** Reusable request wrapping that manages span traces, captures response statuses, and timing-logs endpoints out-of-the-box.
* **JSON Metrics Exporter:** Concurrent-safe live performance snapshots available under `/metrics` in a structured JSON payload.
* **Dynamic Log Level Rotation:** Atomic, lock-free runtime changes using `SetLogLevel` to adjust verbosity during live incidents without container restarts.

---

## Directory Structure

```txt
orkai-observability/
├── cmd/
│   └── api/
│       └── main.go         # API simulation entrypoint
├── observability/
│   ├── config.go           # Configuration validation
│   ├── exporter.go         # Metrics HTTP Exporter
│   ├── logger.go           # High-performance structured JSON Logger
│   ├── metrics.go          # Concurrent safe in-memory metrics
│   ├── middleware.go       # Reusable HTTP Tracing & Logging Middleware
│   ├── observability.go    # Global Facade & package-level API
│   ├── tracer.go           # Thread-safe LIFO Trace Stack & cryptographics
│   ├── transport.go        # Outbound HTTP Client Tracing Transport
│   └── types.go            # Explicit types (Field, Span)
├── test/
│   ├── config_test.go      # Configuration validation tests
│   ├── exporter_test.go    # Exporter endpoint tests
│   ├── logger_test.go      # JSON Logger & dynamic levels tests
│   ├── metrics_test.go     # InMemory Metrics & snapshots tests
│   ├── middleware_test.go  # HTTP Middleware tests
│   ├── observability_test.go     # Global Facade tests
│   ├── tracer_test.go            # LIFO Trace Stack tests
│   ├── transport_test.go         # Outbound HTTP Client Transport tests
│   └── types_test.go             # Explicit types tests
├── go.mod                  # Go module definition
├── .gitignore              # Standard Go repository rules
└── README.md               # Complete usage documentation
```

---

## Installation

Initialize or import the module in your Go project:

```bash
go get github.com/wesleyskap/orkai-observability/observability
```

---

## Quickstart

Here is how to initialize and use the observability package in a typical service handler workflow:

```go
package main

import (
	"context"
	"github.com/wesleyskap/orkai-observability/observability"
	"time"
)

func main() {
	// 1. Initialize the global facade
	cfg := observability.Config{
		ServiceName: "user-service",
		Environment: "production",
		LogLevel:    "info",
	}
	_ = observability.Init(cfg)

	// 2. Simulate a client handler request
	processUserRequest()

	// 3. Print metrics report to the terminal
	observability.Dump()
}

func processUserRequest() {
	start := time.Now()

	// Start a trace span (context-aware)
	ctx, span := observability.StartSpan(context.Background(), "GetUserHandler")
	defer observability.EndSpan(span)

	// Logs automatically capture the active span's trace ID
	observability.Info("handling get user request")

	// Simulate a nested call (e.g. database query)
	performDatabaseQuery(ctx)

	// Restores the parent trace ID after the child span ends
	observability.Info("user fetched successfully", observability.NewStringField("role", "member"))

	// Track custom metrics
	observability.Counter("user_requests_total")
	observability.Latency("user_request_duration", time.Since(start))
}

func performDatabaseQuery(ctx context.Context) {
	// Nested span inherits correlation context
	_, span := observability.StartSpan(ctx, "DBQueryUser")
	defer observability.EndSpan(span)

	observability.Info("selecting user from MySQL", observability.NewStringField("table", "users"))
	time.Sleep(20 * time.Millisecond) // Mock database latency
}
```

### Expected Output

Running the code above produces beautifully correlated, structured console logs:

```txt
[TRACE] Start GetUserHandler trace_id=b78c92a18f0c3d9a
{"level":"INFO","service":"user-service","trace_id":"b78c92a18f0c3d9a","msg":"handling get user request"}
[TRACE] Start DBQueryUser trace_id=d48fa290e29bca02
{"level":"INFO","service":"user-service","trace_id":"d48fa290e29bca02","msg":"selecting user from MySQL","table":"users"}
[TRACE] End DBQueryUser duration=20.0051ms
{"level":"INFO","service":"user-service","trace_id":"b78c92a18f0c3d9a","msg":"user fetched successfully","role":"member"}
[TRACE] End GetUserHandler duration=20.0051ms
=== METRICS ===
user_requests_total: 1
user_request_duration_latency_avg: 20.0051ms
```

---

## Advanced Usage: HTTP Middleware & JSON Exporter

Provides out-of-the-box integrations for high-performance HTTP web servers to automatically manage request trace lifecycles and expose scrapable telemetry.

### 1. HTTP Middleware

Simply wrap your router or mux handler with `observability.HTTPMiddleware` in a single line. It will automatically handle spans, log request start/end, and track request durations:

```go
package main

import (
	"net/http"
	"github.com/wesleyskap/orkai-observability/observability"
)

func main() {
	cfg := observability.Config{ServiceName: "api-service", Environment: "prod"}
	_ = observability.Init(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})

	// Wrap mux in a single line to enable full HTTP Tracing & Logging!
	http.ListenAndServe(":8080", observability.HTTPMiddleware(mux))
}
```

### 2. Live Metrics Exporter (JSON & Prometheus)

Expose `/metrics` dynamically for scrapers or dashboards using `observability.MetricsHTTPHandler()`. It natively supports both custom JSON payloads and the official Prometheus Text Exposition format:

```go
mux := http.NewServeMux()
// Exposes counters, gauges, and latencies
mux.HandleFunc("/metrics", observability.MetricsHTTPHandler())
```

- **JSON Format (Default):** Served on standard calls.
- **Prometheus Format:** Served when calling `/metrics?format=prometheus` or when sending request headers containing `Accept: text/plain`.

### 3. Dynamic Log Level Rotation

Change the active log level dynamically at runtime (e.g., to troubleshoot a live incident) without restarting your process:

```go
// Changes the global log level to "debug" on the fly!
observability.SetLogLevel("debug")
```

### 4. Outbound Context Propagation

Automatically propagate the active `X-Trace-ID` header across outgoing HTTP calls to other services/APIs using standard clients wrapped in `TracingRoundTripper`:

```go
// Creates an HTTP client carrying active parent span trace IDs in header payloads
client := observability.NewTracingClient()

// Executing calls now automatically forwards the active trace to downstream microservices!
resp, err := client.Get("https://api.external.service/users/me")
```

### 5. PII Log Masking & Sanitization

Protect sensitive PII (Personally Identifiable Information) data against accidental leakage in structure logs. The package automatically filters and obfuscates values when keys contain keywords like `password`, `token`, `secret`, `cvv`, `card`, `cpf`, or `email`:

Ensure strict compliance with data safety regulations (such as LGPD and GDPR) by automatically masking sensitive field values inside the structured JSON logs.

```go
// Logging sensitive data will automatically replace values with "[MASKED]"
observability.Info("user login attempt", 
	observability.NewStringField("email", "john.doe@example.com"),
	observability.NewStringField("password", "super-secret-123"),
)
```

#### Custom Sensitive Keys

Add custom PII keywords to the global log sanitization list at runtime:

```go
// Add custom keywords (case-insensitive)
observability.AddSensitiveKeys("socialSecurityNumber", "apiKey")
```

### 6. Structured Error Stack Trace Capture

Accelerate incident investigations by automatically capturing Go call stack traces when recording errors. The package inspects runtime stack frames and appends a clean, compact string under the `"stack_trace"` field:

```go
// Captures stack frames automatically when logging an error
observability.Error("failed database operation", err, 
	observability.NewIntField("retry_count", 3),
)

// Serializes under the "stack_trace" JSON key:
// {"level":"ERROR","msg":"failed database operation","error":"timeout","stack_trace":"main.queryUser:42; main.handleRequest:20","retry_count":3}
```

---

## Running Tests

Our tests are fully isolated inside the `/test` directory, exercising the public API of the package just like a real client application:

```bash
$ go test -v ./test/...
=== RUN   TestValidateConfigValid
--- PASS: TestValidateConfigValid (0.00s)
=== RUN   TestValidateConfigEmptyService
--- PASS: TestValidateConfigEmptyService (0.00s)
=== RUN   TestValidateConfigEmptyEnv
--- PASS: TestValidateConfigEmptyEnv (0.00s)
=== RUN   TestMetricsHTTPHandlerSuccess
--- PASS: TestMetricsHTTPHandlerSuccess (0.00s)
=== RUN   TestMetricsHTTPHandlerPrometheus
--- PASS: TestMetricsHTTPHandlerPrometheus (0.00s)
=== RUN   TestJSONLoggerInfo
--- PASS: TestJSONLoggerInfo (0.00s)
=== RUN   TestJSONLoggerError
--- PASS: TestJSONLoggerError (0.00s)
=== RUN   TestJSONLoggerDynamicLevel
--- PASS: TestJSONLoggerDynamicLevel (0.00s)
=== RUN   TestJSONLoggerPIIMasking
--- PASS: TestJSONLoggerPIIMasking (0.00s)
=== RUN   TestJSONLoggerErrorStackTrace
--- PASS: TestJSONLoggerErrorStackTrace (0.00s)
=== RUN   TestLGPDCompliance
--- PASS: TestLGPDCompliance (0.00s)
=== RUN   TestMetricsIncrement
--- PASS: TestMetricsIncrement (0.00s)
=== RUN   TestMetricsLatency
--- PASS: TestMetricsLatency (0.00s)
=== RUN   TestMetricsGauge
--- PASS: TestMetricsGauge (0.00s)
=== RUN   TestHTTPMiddlewareNewTrace
[TRACE] Start /users trace_id=1eca46527b66ae60
{"level":"INFO","service":"test","trace_id":"1eca46527b66ae60","msg":"incoming request started","method":"POST","path":"/users"}
{"level":"INFO","service":"test","trace_id":"1eca46527b66ae60","msg":"outgoing request finished","method":"POST","path":"/users","status":201,"duration_ms":0}
[TRACE] End /users duration=0s
--- PASS: TestHTTPMiddlewareNewTrace (0.00s)
=== RUN   TestHTTPMiddlewareResumedTrace
{"level":"INFO","service":"test","trace_id":"db3bda","msg":"incoming request started","method":"GET","path":"/profile"}
{"level":"INFO","service":"test","trace_id":"db3bda","msg":"outgoing request finished","method":"GET","path":"/profile","status":200,"duration_ms":0}
[TRACE] End /profile duration=0s
--- PASS: TestHTTPMiddlewareResumedTrace (0.00s)
=== RUN   TestGlobalFacadeInit
--- PASS: TestGlobalFacadeInit (0.00s)
=== RUN   TestGlobalFacadeDelegation
{"level":"INFO","service":"test-service","msg":"delegated log","key":"val"}
[TRACE] Start test-span trace_id=b58c6e59f1dc6d9f
[TRACE] End test-span duration=0s
=== METRICS ===
test_count: 1
test_latency_latency_avg: 10ms
test_gauge: 10.5
--- PASS: TestGlobalFacadeDelegation (0.00s)
=== RUN   TestTracerStart
--- PASS: TestTracerStart (0.00s)
=== RUN   TestTracerEnd
--- PASS: TestTracerEnd (0.00s)
=== RUN   TestTracingRoundTripperNoActiveTrace
--- PASS: TestTracingRoundTripperNoActiveTrace (0.00s)
=== RUN   TestTracingRoundTripperActiveTrace
--- PASS: TestTracingRoundTripperActiveTrace (0.00s)
=== RUN   TestNewStringField
--- PASS: TestNewStringField (0.00s)
=== RUN   TestNewIntField
--- PASS: TestNewIntField (0.00s)
PASS
ok  	github.com/wesleyskap/orkai-observability/test	0.700s
```

---

## License

This project is licensed under the MIT License - see the LICENSE file for details.

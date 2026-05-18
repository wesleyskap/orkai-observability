# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.8.0] - 2026-05-17

### Added
- **High-Performance Asynchronous Logging:** Implemented non-blocking log delivery using thread-safe ring-buffer Go channel semantics to offload physical I/O writes from active request execution paths.
- **Drop-Free Saturation Fallback:** Implemented a zero-memory-growth non-blocking fallback that automatically handles full buffer channel saturation by dropping back to synchronous writes to protect vital telemetry.
- **Graceful Shutdown API:** Exposed a global `Close()` facade function implementing graceful termination that drains the active queue and terminates background worker goroutines.

---

## [1.7.0] - 2026-05-17

### Added
- **Distributed Trace Context Propagation:** Implemented full distributed tracing boundary support injecting context into outbound requests and extracting context from inbound handlers.
- **Multi-Standard Header Resolution:** Seamless support for modern W3C Trace Context (`traceparent`), standard B3 (single/multi-headers), and custom fallback (`X-Trace-ID`) formats.
- **Automated Transport Round-Tripping:** Upgraded `TracingRoundTripper` to inject trace information silently into downstream clients.
- **Resilient Middleware Extraction:** Upgraded `HTTPMiddleware` to automatically extract incoming trace IDs to restore trace hierarchy across systems.

---

## [1.6.0] - 2026-05-17

### Added
- **Multi-Dimensional Metrics (Labels/Tags):** Implemented multi-dimensional metric support for counters, latencies, and gauges. Developers can now group and segment operational telemetry using key-value tag dimensions (labels), mimicking standard production TSDB architectures.
- **Prometheus Labeled Exposition:** Adapted the plain-text metrics exporter to split base metrics and alphabetize labels seamlessly for PromQL scraper ingestion.
- **Unified Facade Delegators:** Exposed clean `CounterWithLabels`, `LatencyWithLabels`, and `GaugeWithLabels` methods directly on the global facade.
- **Test-Driven Telemetry:** Extended the suite with dedicated unit tests in `test/metrics_test.go` and integration validation inside `test/exporter_test.go` to assert correct label sorting and exposition formatting.

---

## [1.5.0] - 2026-05-17

### Added
- **Dynamic Log Rate Limiting & Sampling:** Integrated a highly robust, thread-safe token-bucket rate limiter for structured JSON logs. When the configured burst threshold is exceeded, spam logs are capped/dropped to protect upstream log storage (Elasticsearch, Loki, etc.) and save container CPU cycles.
- **Diagnostic Log Sampling:** Under rate-limited state, the package automatically applies 10% diagnostic sampling, allowing 1 out of every 10 dropped logs to be emitted, marked with `"log_burst_throttled": "true"` to maintain minimal context.
- **Tuneable Configuration:** Extended `Config` struct with `EnableRateLimit`, `RateLimitBurst`, and `RateLimitRate` fields for seamless production tuning.
- **Robust Test Coverage:** Added unit test coverage under `test/logger_test.go` (`TestLogRateLimitingDrops` and `TestLogRateLimitingSamples`) verifying exact token consumption, drop actions, and sampling assertions.

---

## [1.4.0] - 2026-05-17

### Added
- **Context-Aware Log Correlation:** Implemented native context-aware log correlation APIs. Developers can use `InfoContext`, `DebugContext`, `WarnContext`, and `ErrorContext` to automatically extract the active trace ID from `context.Context` parameters using package-level `ContextWithTraceID` and `TraceIDFromContext` utilities.

  ```go
  // Invocations are clean and contextual:
  ctx := observability.ContextWithTraceID(context.Background(), "my-trace-id")
  observability.InfoContext(ctx, "processing payment")
  ```

- **Context Correlation Test Coverage:** Added unit test coverage under `test/logger_test.go` (`TestJSONLoggerContextTraceCorrelation`) asserting context-to-log trace propagation and dynamic fallback options.

### Changed
- **Documentation:** Updated the README to include documentation and code snippets for context-aware trace logging and correlation.

---

## [1.3.0] - 2026-05-17

### Added
- **Structured Error Stack Trace Capture:** Implemented dynamic Go execution call frame inspection inside the structured JSON logger. When calling `observability.Error`, the logger uses `runtime.Callers` and `runtime.CallersFrames` to capture, format, and inject call frames under the `"stack_trace"` JSON key.

  ```go
  // No signature or registration changes:
  observability.Error("failed db select", err)

  // Automatically captures call frames:
  // {"level":"ERROR","msg":"failed db select","error":"timeout","stack_trace":"main.queryUser:42; main.main:12"}
  ```

- **Stack Trace Test Coverage:** Added unit test coverage under `test/logger_test.go` (`TestJSONLoggerErrorStackTrace`) confirming runtime calling frame capture and format correctness.

### Changed
- **Documentation:** Updated the README to describe the new automatic error stack trace capture capabilities.

---

## [1.2.0] - 2026-05-17

### Added
- **PII Log Masking & Sanitization:** Integrated automated Personally Identifiable Information (PII) masking inside the structured JSON logger. Key names containing keywords like `password`, `token`, `secret`, `cvv`, `card`, `cpf`, or `email` automatically have their values replaced with `"[MASKED]"`.
- **Runtime Sensitive Keys Registration:** Provided thread-safe global `AddSensitiveKeys(keys ...string)` API protected under concurrent read-write locks (`sync.RWMutex`) to register custom PII keywords at runtime.

Ensure strict compliance with data safety regulations (such as LGPD and GDPR) by automatically masking sensitive field values inside the structured JSON logs.

  ```go
  // Invocations are clean and automatic:
  observability.Info("user signup",
      observability.NewStringField("email", "test@test.com"),
      observability.NewStringField("password", "secret-pass"),
  )

  // Serializes automatically as:
  // {"level":"INFO","msg":"user signup","email":"[MASKED]","password":"[MASKED]"}
  ```

- **Masking Test Coverage:** Added assertions in `test/logger_test.go` (`TestJSONLoggerPIIMasking`) verifying default masking and custom PII registrations.

### Changed
- **Documentation:** Updated the README to describe the PII masking capabilities and runtime registration syntax.

---

## [1.0.11] - 2026-05-17

### Added
- **Prometheus Metric Exposition Format:** Added support for the official Prometheus Text Exposition format (counters, gauges, and latency averages) inside the metrics HTTP handler. The registration invocation remains completely identical and backward-compatible, while the internal handler now dynamically negotiates the response format based on standard headers (`Accept: text/plain`) or query string parameters (`format=prometheus`).

  ```go
  // The invocation syntax remains completely unchanged:
  mux.HandleFunc("/metrics", observability.MetricsHTTPHandler())

  // The handler now automatically responds in:
  // - Custom JSON format (by default)
  // - Official Prometheus Text format (when queried via ?format=prometheus or Accept: text/plain)
  ```

- **Prometheus Test Suites:** Added integration tests inside `test/exporter_test.go` asserting standard Prometheus headers, metric types, and values.

### Changed
- **Documentation:** Updated the README to describe the dual-format capabilities of the metrics exporter.

---

## [1.0.10] - 2026-05-17

### Added
- **Outbound Context Propagation:** Added `TracingRoundTripper` implementing `http.RoundTripper` to automatically inject the active LIFO trace ID (`X-Trace-ID` header) into outbound HTTP requests.
- **Client Transport Wrapper:** Provided `NewTracingClient()` constructor to instantiate pre-configured tracing clients out-of-the-box in zero developer boilerplate.

  #### Before (Manual trace header injection inside every API call)
  ```go
  // In your API client functions, you had to manually fetch and pass headers
  req, _ := http.NewRequest("GET", "https://api.service/users", nil)
  activeID := observability.GetActiveTraceID()
  if activeID != "" {
      req.Header.Set("X-Trace-ID", activeID)
  }
  resp, err := http.DefaultClient.Do(req)
  ```

  #### After (Automatic distributed trace context propagation)
  ```go
  // Instantiates a client pre-configured with tracing transport
  client := observability.NewTracingClient()

  // Outbound calls automatically inherit and carry the active LIFO trace context!
  resp, err := client.Get("https://api.service/users")
  ```

- **Transport Test Suites:** Added integration test coverage under `test/transport_test.go` asserting automated header injection and stack interaction.

### Changed
- **Documentation:** Updated the README directory tree map and added usage examples for outbound transport client injection.

---

## [1.0.9] - 2026-05-17

### Added
- **Repository Changelog:** Created the initial structured `CHANGELOG.md` mapped to release tags.

### Changed
- **Repository Rules:** Updated `.gitignore` to keep IDE files and planning boards properly ignored.

---

## [1.0.8] - 2026-05-17

### Changed
- **Documentation:** Updated the README Mermaid sequence diagram to show the new automated middleware intercept flow, enriched the core Feature list, and appended new components to the directory tree map.

---

## [1.0.7] - 2026-05-17

### Changed
- **Codebase Formatting Alignment:** Formatted and simplified all Go source files recursively using the official `gofmt -w -s .` formatter to achieve a perfect 100% score on Go Report Card criteria.

---

## [1.0.6] - 2026-05-17

### Added
- **HTTP Tracing Middleware:** Added reusable `observability.HTTPMiddleware` to automatically intercept requests, correlate incoming/generated `X-Trace-ID` headers, timing-log cycles, and capture status codes out-of-the-box.
  
  #### Before (Manual Span boilerplate inside each Controller)
  ```go
  func (c *UserController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
      // Manual tracing at the top of every endpoint
      ctx, span := observability.StartSpan(req.Context(), "UserController.ServeHTTP")
      defer observability.EndSpan(span)

      // Controller logic...
  }
  ```
  
  #### After (Clean, automated Middleware integration)
  ```go
  // Inside your Router (Mux) configuration:
  mux := http.NewServeMux()
  mux.Handle("/users", userRes)

  // Intercepts HTTP boundaries, generates traces & logs status codes
  handler := observability.HTTPMiddleware(mux)
  
  // Clean Controllers!
  func (c *UserController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
      // No more tracing boilerplate! Ready to use active traces inside downstream calls.
  }
  ```

- **JSON Metrics Exporter:** Added scrapable `/metrics` endpoint using `observability.MetricsHTTPHandler()` to serve concurrent-safe JSON snapshots of counters, average latencies, and gauges.
  
  #### After (JSON scrape endpoint)
  ```go
  mux := http.NewServeMux()
  // Exposes in-memory counters, gauges & averages as scrapable JSON
  mux.HandleFunc("/metrics", observability.MetricsHTTPHandler())
  ```

- **Dynamic Log Level Rotation:** Supported atomic, lock-free global level changes at runtime (`observability.SetLogLevel`) using high-performance `sync/atomic` operations.

  #### Before (Static configuration setup)
  ```go
  cfg := observability.Config{LogLevel: "info"}
  observability.Init(cfg)
  // Log Level is locked at "info" until the container is restarted
  ```

  #### After (Runtime hot level rotation)
  ```go
  // Switch dynamically to DEBUG in production to troubleshoot an active outage!
  observability.SetLogLevel("debug")
  ```

- **Extended Test Coverage:** Added unit test suites verifying request interception (`test/middleware_test.go`), performance snapshots (`test/exporter_test.go`), and dynamic log-level filtering (`test/logger_test.go`).

---

## [1.0.5] - 2026-05-17

### Changed
- **Go Version Backward Compatibility:** Downgraded the Go compiler requirement to `>= 1.22` in `go.mod` to guarantee compilation inside Docker containers and backward compatibility with older Go environments (e.g. Go `1.22.12` running in standard local API servers).

---

## [1.0.4] - 2026-05-17

### Added
- **Open Source Community Guidelines:** Added permissive `CONTRIBUTING.md` and `CODE_OF_CONDUCT.md` assets to meet standard community requirements.

---

## [1.0.3] - 2026-05-17

### Added
- **Documentation Comments:** Added standard Go doc comments and usage examples across all public interfaces and methods.

---

## [1.0.2] - 2026-05-17

### Fixed
- **Testing Module Paths:** Fixed incorrect package module import paths inside `test/tracer_test.go` and `cmd/api/main.go` to match the official `github.com/wesleyskap/orkai-observability` path.

---

## [1.0.1] - 2026-05-17

### Fixed
- **Go Module Path:** Adjusted root module path inside `go.mod` to correctly point to the public GitHub repository.

---

## [1.0.0] - 2026-05-17

### Added
- **Permissive License:** Created the `LICENSE` file under the permissive MIT License.
- **Unified Facade:** Integrated package-level global functions (`Info`, `Counter`, `StartSpan`) to orchestrate telemetry components transparently.
- **LIFO Nested Spans:** Context-aware local trace spans that automatically restore parent correlation context upon nested completion (using in-memory LIFO stack).
- **Custom Structured JSON Logger:** High-performance, reflection-free logger outputting directly to console or custom writers.
- **Thread-Safe Metrics:** In-memory tracking for cumulative counters, arithmetic average latencies over multi-sample periods, and gauges.

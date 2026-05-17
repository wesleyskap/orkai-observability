# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

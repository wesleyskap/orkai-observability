package test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wesleyskap/orkai-observability/observability"
)

// FakeWriter is a named mock for io.Writer to capture output.
type FakeWriter struct {
	Buf bytes.Buffer
}

// Write captures the input bytes to the buffer.
func (f *FakeWriter) Write(p []byte) (int, error) {
	n, err := f.Buf.Write(p)
	return n, err
}

// TestJSONLoggerInfo verifies that Info logs write correct JSON data.
func TestJSONLoggerInfo(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "auth-service")
	logger.Info("user authenticated", observability.NewStringField("role", "admin"))
	output := fw.Buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Fatalf("expected level INFO, got %s", output)
	}
	if !strings.Contains(output, `"role":"admin"`) {
		t.Fatalf("expected role:admin, got %s", output)
	}
}

// TestJSONLoggerError verifies that Error logs write correct JSON and errors.
func TestJSONLoggerError(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "auth-service")
	errVal := errors.New("connection failed")
	logger.Error("db error", errVal, observability.NewIntField("retry", 3))
	output := fw.Buf.String()
	if !strings.Contains(output, `"level":"ERROR"`) {
		t.Fatalf("expected level ERROR, got %s", output)
	}
	if !strings.Contains(output, `"retry":3`) {
		t.Fatalf("expected retry:3, got %s", output)
	}
}

// TestJSONLoggerDynamicLevel asserts that logs are filtered and rotate levels dynamically.
func TestJSONLoggerDynamicLevel(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "test-service")
	logger.SetLevel("warn")
	logger.Info("should not print")
	logger.Warn("should print")
	output := fw.Buf.String()
	if strings.Contains(output, "should not print") {
		t.Errorf("expected info log to be filtered, got %s", output)
	}
	if !strings.Contains(output, "should print") {
		t.Errorf("expected warn log to print, got %s", output)
	}
	fw.Buf.Reset()
	logger.SetLevel("debug")
	logger.Debug("now prints")
	if !strings.Contains(fw.Buf.String(), "now prints") {
		t.Errorf("expected debug log to print after rotation, got %s", fw.Buf.String())
	}
}

// TestJSONLoggerPIIMasking verifies that sensitive PII fields are automatically masked.
func TestJSONLoggerPIIMasking(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "test-service")

	logger.Info("user input", observability.NewStringField("password", "my-secret-password"))
	output1 := fw.Buf.String()
	if strings.Contains(output1, "my-secret-password") || !strings.Contains(output1, `"[MASKED]"`) {
		t.Errorf("expected password to be masked, got %s", output1)
	}

	fw.Buf.Reset()
	observability.AddSensitiveKeys("socialSecurityNumber")
	logger.Info("user document", observability.NewStringField("socialSecurityNumber", "999-99-9999"))
	output2 := fw.Buf.String()
	if strings.Contains(output2, "999-99-9999") || !strings.Contains(output2, `"[MASKED]"`) {
		t.Errorf("expected socialSecurityNumber to be masked, got %s", output2)
	}
}

// TestJSONLoggerErrorStackTrace verifies that Error logs capture and serialize the stack trace.
func TestJSONLoggerErrorStackTrace(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "test-service")
	logger.Error("failed task", errors.New("internal failure"))
	output := fw.Buf.String()
	if !strings.Contains(output, `"stack_trace"`) {
		t.Fatalf("expected stack_trace key in JSON log, got %s", output)
	}
	if !strings.Contains(output, "TestJSONLoggerErrorStackTrace") {
		t.Errorf("expected calling function in stack trace, got %s", output)
	}
}

// TestLGPDCompliance simulates a strict LGPD/GDPR compliance scan asserting sensitive user fields are masked.
func TestLGPDCompliance(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "user-service")
	observability.AddSensitiveKeys("phone", "address")
	logger.Info("user signup",
		observability.NewStringField("cpf", "123-00"),
		observability.NewStringField("phone", "+55119999"),
		observability.NewStringField("user_id", "usr_99"),
	)
	out := fw.Buf.String()
	if !strings.Contains(out, `"cpf":"[MASKED]"`) || !strings.Contains(out, `"phone":"[MASKED]"`) {
		t.Errorf("expected PII to be masked, got: %s", out)
	}
	if !strings.Contains(out, `"user_id":"usr_99"`) {
		t.Errorf("expected user_id to be plain text, got: %s", out)
	}
}

// TestJSONLoggerContextTraceCorrelation asserts log messages propagate trace IDs correctly from context.
func TestJSONLoggerContextTraceCorrelation(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "test-service")
	ctx := observability.ContextWithTraceID(context.Background(), "my-custom-ctx-trace-123")
	logger.InfoContext(ctx, "contextual log info")
	out := fw.Buf.String()
	if !strings.Contains(out, `"trace_id":"my-custom-ctx-trace-123"`) {
		t.Fatalf("expected context trace ID, got: %s", out)
	}
	fw.Buf.Reset()
	logger.SetTraceProvider(func() string { return "fallback-provider-id" })
	logger.InfoContext(context.Background(), "contextual log fallback")
	out2 := fw.Buf.String()
	if !strings.Contains(out2, `"trace_id":"fallback-provider-id"`) {
		t.Fatalf("expected fallback trace ID, got: %s", out2)
	}
}

// TestLogRateLimitingDrops asserts that logs exceeding token burst capacity are dropped.
func TestLogRateLimitingDrops(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "rate-limit-service")
	limiter := observability.NewLogRateLimiter(3, 0, true)
	logger.SetRateLimiter(limiter)
	for i := 0; i < 5; i++ {
		logger.Info("spam message")
	}
	lines := strings.Split(strings.TrimSpace(fw.Buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected exactly 3 logs, got %d. logs: %v", len(lines), lines)
	}
}

// TestLogRateLimitingSamples asserts that rate-limited logs are sampled with a throttled attribute.
func TestLogRateLimitingSamples(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "rate-limit-service")
	limiter := observability.NewLogRateLimiter(0, 0, true)
	logger.SetRateLimiter(limiter)
	for i := 0; i < 10; i++ {
		logger.Info("spam message")
	}
	sampledLines := strings.Split(strings.TrimSpace(fw.Buf.String()), "\n")
	if len(sampledLines) != 1 {
		t.Fatalf("expected exactly 1 sampled log from 10 rate-limited logs, got %d", len(sampledLines))
	}
	if !strings.Contains(sampledLines[0], `"log_burst_throttled":"true"`) {
		t.Errorf("expected sampled log to be marked, got %s", sampledLines[0])
	}
}

// TestAsyncLoggerSuccess asserts that async logging queues and flushes entries gracefully.
func TestAsyncLoggerSuccess(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "async-service")
	logger.ConfigureAsync(true, 10)
	logger.Info("hello async")
	logger.Info("world async")
	_ = logger.Close()
	output := fw.Buf.String()
	if !strings.Contains(output, "hello async") || !strings.Contains(output, "world async") {
		t.Fatalf("expected flushed async logs, got: %s", output)
	}
}

// TestAsyncLoggerSaturation asserts that full async channel conditions trigger safe synchronous fallback.
func TestAsyncLoggerSaturation(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "saturation-service")
	logger.ConfigureAsync(true, 1)
	for i := 0; i < 20; i++ {
		logger.Info("spam log")
	}
	_ = logger.Close()
	output := fw.Buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 20 {
		t.Fatalf("expected all logs to be preserved via fallback, got only %d", len(lines))
	}
}

package test

import (
	"bytes"
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

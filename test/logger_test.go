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

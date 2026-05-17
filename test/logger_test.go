package test

import (
	"bytes"
	"errors"
	"orkai-observability/observability"
	"testing"
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

// TestJSONLoggerInfo verifies that Info logs write data.
func TestJSONLoggerInfo(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "auth-service")
	logger.Info("user authenticated", observability.NewStringField("role", "admin"))
	output := fw.Buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}
}

// TestJSONLoggerError verifies that Error logs write data.
func TestJSONLoggerError(t *testing.T) {
	fw := &FakeWriter{}
	logger := observability.NewJSONLogger(fw, "auth-service")
	errVal := errors.New("connection failed")
	logger.Error("db error", errVal, observability.NewIntField("retry", 3))
	output := fw.Buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}
}

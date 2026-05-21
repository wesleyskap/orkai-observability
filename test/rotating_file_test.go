package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

// TestRotatingFileWriterSimple asserts standard writing and file creation.
func TestRotatingFileWriterSimple(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	writer, err := observability.NewRotatingFileWriter(logPath, 100, 3)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()
	data := []byte("hello world\n")
	n, err := writer.Write(data)
	if err != nil || n != len(data) {
		t.Fatalf("write failed: n=%d err=%v", n, err)
	}
}

// TestRotatingFileWriterRotation asserts size-based rotation and backups.
func TestRotatingFileWriterRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	writer, err := observability.NewRotatingFileWriter(logPath, 10, 2)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()
	writeThreeBlocks(t, writer)
	assertBackupsExist(t, tmpDir)
}

func writeThreeBlocks(t *testing.T, w *observability.RotatingFileWriter) {
	w.Write([]byte("1234567890")) // 10 bytes -> triggers rotation on next write
	w.Write([]byte("1234567890")) // 10 bytes -> rotates test.log to test.log.1
	w.Write([]byte("1234567890")) // 10 bytes -> rotates test.log.1 to test.log.2
}

func assertBackupsExist(t *testing.T, tmpDir string) {
	_, err1 := os.Stat(filepath.Join(tmpDir, "test.log"))
	_, err2 := os.Stat(filepath.Join(tmpDir, "test.log.1"))
	_, err3 := os.Stat(filepath.Join(tmpDir, "test.log.2"))
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("expected backups to exist: %v, %v, %v", err1, err2, err3)
	}
}

// TestRotatingFileWriterLimit asserts max backups eviction logic.
func TestRotatingFileWriterLimit(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	writer, err := observability.NewRotatingFileWriter(logPath, 5, 1)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()
	writeAndTriggerEviction(t, writer)
	assertEvictionState(t, tmpDir)
}

func writeAndTriggerEviction(t *testing.T, w *observability.RotatingFileWriter) {
	w.Write([]byte("abcde")) // 5 bytes
	w.Write([]byte("abcde")) // 5 bytes -> backup test.log.1 created
	w.Write([]byte("abcde")) // 5 bytes -> backup test.log.1 rotated to test.log.2 (but max is 1!)
}

func assertEvictionState(t *testing.T, tmpDir string) {
	_, err1 := os.Stat(filepath.Join(tmpDir, "test.log"))
	_, err2 := os.Stat(filepath.Join(tmpDir, "test.log.1"))
	_, err3 := os.Stat(filepath.Join(tmpDir, "test.log.2"))
	if err1 != nil || err2 != nil || err3 == nil {
		t.Fatalf("expected log and log.1, but not log.2: %v, %v, %v", err1, err2, err3)
	}
}

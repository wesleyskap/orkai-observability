package test

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wesleyskap/orkai-observability/v2/observability"
)

type FakeWriteCloser struct {
	buffer []byte
	closed bool
}

func (fw *FakeWriteCloser) Write(p []byte) (int, error) {
	fw.buffer = append(fw.buffer, p...)
	return len(p), nil
}

func (fw *FakeWriteCloser) Close() error {
	fw.closed = true
	return nil
}

type FakeProfileFileCreator struct {
	mu           sync.Mutex
	createdFiles map[string]*FakeWriteCloser
}

func newFakeProfileFileCreator() *FakeProfileFileCreator {
	return &FakeProfileFileCreator{
		createdFiles: make(map[string]*FakeWriteCloser),
	}
}

func (fc *FakeProfileFileCreator) CreateProfileFile(dir string, name string) (io.WriteCloser, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fw := &FakeWriteCloser{}
	fc.createdFiles[name] = fw
	return fw, nil
}

func TestAutoProfiler(t *testing.T) {
	fc := newFakeProfileFileCreator()
	ap := observability.NewAutoProfiler(fc)
	cfg := observability.Config{
		EnableAutoPprof:         true,
		PprofGoroutinesLimit:    10,
		PprofHeapThresholdBytes: 1024 * 1024,
		PprofProfileDuration:    50 * time.Millisecond,
		PprofOutputDir:          "test_out",
	}

	// 1. Should not trigger under limit
	ap.CheckAndTrigger(cfg, 5, 500)
	if len(fc.createdFiles) != 0 {
		t.Fatalf("expected 0 files, got %d", len(fc.createdFiles))
	}

	// 2. Should trigger above limit
	ap.CheckAndTrigger(cfg, 15, 500)
	time.Sleep(100 * time.Millisecond) // wait for profile to finish

	fc.mu.Lock()
	defer fc.mu.Unlock()
	if len(fc.createdFiles) != 2 {
		t.Fatalf("expected 2 files (cpu and heap), got %d", len(fc.createdFiles))
	}

	var hasCPU, hasHeap bool
	for name, file := range fc.createdFiles {
		if strings.HasPrefix(name, "cpu_") {
			hasCPU = true
		}
		if strings.HasPrefix(name, "heap_") {
			hasHeap = true
		}
		if !file.closed {
			t.Errorf("file %s was not closed", name)
		}
	}

	if !hasCPU || !hasHeap {
		t.Errorf("expected both CPU and Heap profiles to be created")
	}
}

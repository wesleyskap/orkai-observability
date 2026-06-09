package observability

import (
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

// ProfileFileCreator defines the interface for creating profile output files.
type ProfileFileCreator interface {
	CreateProfileFile(dir string, name string) (io.WriteCloser, error)
}

// OSProfileFileCreator implements ProfileFileCreator using the OS filesystem.
type OSProfileFileCreator struct{}

// CreateProfileFile creates a profile file in the specified directory.
func (o OSProfileFileCreator) CreateProfileFile(dir string, name string) (io.WriteCloser, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(dir, name))
}

// AutoProfiler monitors and triggers pprof reports.
type AutoProfiler struct {
	creator     ProfileFileCreator
	lastRun     time.Time
	mu          sync.Mutex
	isProfiling int32
}

// NewAutoProfiler creates a new instance of AutoProfiler.
func NewAutoProfiler(c ProfileFileCreator) *AutoProfiler {
	return &AutoProfiler{
		creator: c,
	}
}

// CheckAndTrigger checks limits and starts profiling if exceeded.
func (ap *AutoProfiler) CheckAndTrigger(cfg Config, goroutines int, heapAlloc int64) {
	if !cfg.EnableAutoPprof {
		return
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if !ap.shouldProfile(cfg, goroutines, heapAlloc) {
		return
	}
	ap.lastRun = time.Now()
	atomic.StoreInt32(&ap.isProfiling, 1)
	ts := time.Now().Format("20060102_150405")
	go ap.triggerCPUProfile(cfg, ts)
	go ap.triggerHeapProfile(cfg, ts)
}

func (ap *AutoProfiler) shouldProfile(cfg Config, goroutines int, heapAlloc int64) bool {
	if atomic.LoadInt32(&ap.isProfiling) == 1 {
		return false
	}
	if time.Since(ap.lastRun) < 5*time.Minute {
		return false
	}
	gLimit := cfg.PprofGoroutinesLimit > 0 && goroutines >= cfg.PprofGoroutinesLimit
	hLimit := cfg.PprofHeapThresholdBytes > 0 && heapAlloc >= cfg.PprofHeapThresholdBytes
	return gLimit || hLimit
}

func (ap *AutoProfiler) triggerCPUProfile(cfg Config, timestamp string) {
	w, err := ap.creator.CreateProfileFile(cfg.PprofOutputDir, "cpu_"+timestamp+".pprof")
	if err != nil {
		return
	}
	if err := pprof.StartCPUProfile(w); err != nil {
		w.Close()
		return
	}
	go ap.stopCPUProfileAfter(w, cfg.PprofProfileDuration)
}

func (ap *AutoProfiler) triggerHeapProfile(cfg Config, timestamp string) {
	w, err := ap.creator.CreateProfileFile(cfg.PprofOutputDir, "heap_"+timestamp+".pprof")
	if err != nil {
		return
	}
	defer w.Close()
	_ = pprof.WriteHeapProfile(w)
}

func (ap *AutoProfiler) stopCPUProfileAfter(w io.WriteCloser, d time.Duration) {
	time.Sleep(d)
	pprof.StopCPUProfile()
	w.Close()
	atomic.StoreInt32(&ap.isProfiling, 0)
}

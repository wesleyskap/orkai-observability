package observability

import (
	"context"
	"runtime"
	"time"
)

// StartSystemTelemetry starts a background goroutine collecting Go runtime telemetry.
func StartSystemTelemetry(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	go runTelemetryLoop(ctx, interval)
}

func runTelemetryLoop(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			collectRuntimeStats()
		}
	}
}

func collectRuntimeStats() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	goroutines := runtime.NumGoroutine()
	Gauge("go_goroutines", float64(goroutines))
	Gauge("go_mem_heap_alloc_bytes", float64(mem.HeapAlloc))
	Gauge("go_mem_heap_sys_bytes", float64(mem.HeapSys))
	Gauge("go_gc_completed_count", float64(mem.NumGC))
	checkAutoPprof(goroutines, int64(mem.HeapAlloc))
}

func checkAutoPprof(goroutines int, heapAlloc int64) {
	inst := getGlobal()
	if inst == nil || globalProfiler == nil {
		return
	}
	globalProfiler.CheckAndTrigger(inst.Config, goroutines, heapAlloc)
}

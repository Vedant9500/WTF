package metrics

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// PerformanceMonitor tracks application performance metrics
type PerformanceMonitor struct {
	collector *MetricsCollector
	enabled   bool
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		collector: NewMetricsCollector(),
		enabled:   true,
	}
}

// Enable enables or disables performance monitoring
func (pm *PerformanceMonitor) Enable(enabled bool) {
	pm.enabled = enabled
}

// IsEnabled returns whether performance monitoring is enabled
func (pm *PerformanceMonitor) IsEnabled() bool {
	return pm.enabled
}

// RecordSearchOperation records metrics for a search operation
func (pm *PerformanceMonitor) RecordSearchOperation(duration time.Duration, resultCount int, cacheHit bool, queryLength int) {
	if !pm.enabled {
		return
	}

	// Record search duration
	searchTimer := pm.collector.Timer("search_duration", map[string]string{
		"cache_hit": fmt.Sprintf("%t", cacheHit),
	})
	searchTimer.Histogram().Observe(float64(duration.Nanoseconds()) / 1e6) // Convert to milliseconds

	// Record result count
	resultGauge := pm.collector.Gauge("search_results", nil)
	resultGauge.Set(float64(resultCount))

	// Record query length
	queryLengthHist := pm.collector.Histogram("query_length", nil)
	queryLengthHist.Observe(float64(queryLength))

	// Increment search counter
	searchCounter := pm.collector.Counter("searches_total", map[string]string{
		"cache_hit": fmt.Sprintf("%t", cacheHit),
	})
	searchCounter.Inc()

	// Record cache hit ratio
	if cacheHit {
		cacheHitCounter := pm.collector.Counter("cache_hits_total", nil)
		cacheHitCounter.Inc()
	} else {
		cacheMissCounter := pm.collector.Counter("cache_misses_total", nil)
		cacheMissCounter.Inc()
	}
}

// RecordDatabaseOperation records metrics for database operations
func (pm *PerformanceMonitor) RecordDatabaseOperation(operation string, duration time.Duration, success bool) {
	if !pm.enabled {
		return
	}

	// Record operation duration
	dbTimer := pm.collector.Timer("database_operation_duration", map[string]string{
		"operation": operation,
		"success":   fmt.Sprintf("%t", success),
	})
	dbTimer.Histogram().Observe(float64(duration.Nanoseconds()) / 1e6)

	// Increment operation counter
	dbCounter := pm.collector.Counter("database_operations_total", map[string]string{
		"operation": operation,
		"success":   fmt.Sprintf("%t", success),
	})
	dbCounter.Inc()
}

// RecordMemoryUsage records current memory usage
func (pm *PerformanceMonitor) RecordMemoryUsage() {
	if !pm.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Record various memory metrics
	allocGauge := pm.collector.Gauge("memory_alloc_bytes", nil)
	allocGauge.Set(float64(m.Alloc))

	sysGauge := pm.collector.Gauge("memory_sys_bytes", nil)
	sysGauge.Set(float64(m.Sys))

	gcGauge := pm.collector.Gauge("gc_runs_total", nil)
	gcGauge.Set(float64(m.NumGC))

	goroutineGauge := pm.collector.Gauge("goroutines_active", nil)
	goroutineGauge.Set(float64(runtime.NumGoroutine()))
}

// StartMemoryMonitoring starts periodic memory monitoring
func (pm *PerformanceMonitor) StartMemoryMonitoring(ctx context.Context, interval time.Duration) {
	if !pm.enabled {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.RecordMemoryUsage()
		}
	}
}

// GetPerformanceReport generates a performance report
func (pm *PerformanceMonitor) GetPerformanceReport() PerformanceReport {
	metrics := pm.collector.GetAllMetrics()
	systemMetrics := pm.collector.GetSystemMetrics()

	report := PerformanceReport{
		Timestamp:          time.Now(),
		ApplicationMetrics: metrics,
		SystemMetrics:      systemMetrics,
	}

	// Calculate derived metrics
	report.calculateDerivedMetrics()

	return report
}

// PerformanceReport contains performance metrics and analysis
type PerformanceReport struct {
	Timestamp          time.Time `json:"timestamp"`
	ApplicationMetrics []Metric  `json:"application_metrics"`
	SystemMetrics      []Metric  `json:"system_metrics"`

	// Derived metrics
	AverageSearchTime float64 `json:"average_search_time_ms"`
	CacheHitRatio     float64 `json:"cache_hit_ratio"`
	SearchesPerSecond float64 `json:"searches_per_second"`
	MemoryUsageMB     float64 `json:"memory_usage_mb"`
	GoroutineCount    int     `json:"goroutine_count"`
}

// calculateDerivedMetrics calculates derived performance metrics
func (pr *PerformanceReport) calculateDerivedMetrics() {
	metricMap := make(map[string]float64)

	// Build metric map for easy lookup
	for _, metric := range pr.ApplicationMetrics {
		metricMap[metric.Name] = metric.Value
	}
	for _, metric := range pr.SystemMetrics {
		metricMap[metric.Name] = metric.Value
	}

	// Calculate average search time
	if searchCount := metricMap["search_duration_count"]; searchCount > 0 {
		if searchSum := metricMap["search_duration_sum"]; searchSum > 0 {
			pr.AverageSearchTime = searchSum / searchCount
		}
	}

	// Calculate cache hit ratio
	cacheHits := metricMap["cache_hits_total"]
	cacheMisses := metricMap["cache_misses_total"]
	totalRequests := cacheHits + cacheMisses
	if totalRequests > 0 {
		pr.CacheHitRatio = cacheHits / totalRequests
	}

	// Calculate searches per second (approximate)
	if uptime := metricMap["system_uptime"]; uptime > 0 {
		totalSearches := metricMap["searches_total"]
		pr.SearchesPerSecond = totalSearches / uptime
	}

	// Memory usage in MB
	if memAlloc := metricMap["system_memory_alloc"]; memAlloc > 0 {
		pr.MemoryUsageMB = memAlloc / (1024 * 1024)
	}

	// Goroutine count
	pr.GoroutineCount = int(metricMap["system_goroutines"])
}

// String returns a string representation of the performance report
func (pr *PerformanceReport) String() string {
	return fmt.Sprintf(`Performance Report (%s):
  Average Search Time: %.2f ms
  Cache Hit Ratio: %.2f%%
  Searches/Second: %.2f
  Memory Usage: %.2f MB
  Active Goroutines: %d`,
		pr.Timestamp.Format("2006-01-02 15:04:05"),
		pr.AverageSearchTime,
		pr.CacheHitRatio*100,
		pr.SearchesPerSecond,
		pr.MemoryUsageMB,
		pr.GoroutineCount)
}

// BenchmarkResult represents the result of a performance benchmark
type BenchmarkResult struct {
	Name        string        `json:"name"`
	Iterations  int           `json:"iterations"`
	Duration    time.Duration `json:"duration"`
	NsPerOp     int64         `json:"ns_per_op"`
	BytesPerOp  int64         `json:"bytes_per_op"`
	AllocsPerOp int64         `json:"allocs_per_op"`
	MemoryUsed  int64         `json:"memory_used"`
	Timestamp   time.Time     `json:"timestamp"`
}

// String returns a string representation of the benchmark result
func (br *BenchmarkResult) String() string {
	return fmt.Sprintf("%s: %d iterations, %d ns/op, %d B/op, %d allocs/op",
		br.Name, br.Iterations, br.NsPerOp, br.BytesPerOp, br.AllocsPerOp)
}

// Benchmarker provides benchmarking capabilities
type Benchmarker struct {
	monitor *PerformanceMonitor
}

// NewBenchmarker creates a new benchmarker
func NewBenchmarker(monitor *PerformanceMonitor) *Benchmarker {
	return &Benchmarker{
		monitor: monitor,
	}
}

// BenchmarkFunction benchmarks a function execution
func (b *Benchmarker) BenchmarkFunction(name string, fn func(), iterations int) BenchmarkResult {
	// Record initial memory stats
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()

	for i := 0; i < iterations; i++ {
		fn()
	}

	duration := time.Since(start)

	// Record final memory stats
	runtime.ReadMemStats(&m2)

	// Safe integer conversions to prevent overflow
	var bytesPerOp, allocsPerOp, memoryUsed int64
	
	if iterations > 0 {
		// Check for potential overflow before conversion
		totalAllocDiff := m2.TotalAlloc - m1.TotalAlloc
		mallocsDiff := m2.Mallocs - m1.Mallocs
		allocDiff := m2.Alloc - m1.Alloc
		
		// Safe conversion with overflow check
		if totalAllocDiff <= uint64(^uint64(0)>>1) { // Check if fits in int64
			bytesPerOp = int64(totalAllocDiff) / int64(iterations)
		}
		if mallocsDiff <= uint64(^uint64(0)>>1) {
			allocsPerOp = int64(mallocsDiff) / int64(iterations)
		}
		if allocDiff <= uint64(^uint64(0)>>1) {
			memoryUsed = int64(allocDiff)
		}
	}

	return BenchmarkResult{
		Name:        name,
		Iterations:  iterations,
		Duration:    duration,
		NsPerOp:     duration.Nanoseconds() / int64(iterations),
		BytesPerOp:  bytesPerOp,
		AllocsPerOp: allocsPerOp,
		MemoryUsed:  memoryUsed,
		Timestamp:   time.Now(),
	}
}

// ProfileMemory profiles memory usage during function execution
func (b *Benchmarker) ProfileMemory(name string, fn func()) MemoryProfile {
	var m1, m2 runtime.MemStats

	// Force GC and get baseline
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()
	fn()
	duration := time.Since(start)

	runtime.ReadMemStats(&m2)

	return MemoryProfile{
		Name:            name,
		Duration:        duration,
		AllocBefore:     m1.Alloc,
		AllocAfter:      m2.Alloc,
		AllocDelta:      m2.Alloc - m1.Alloc,
		TotalAllocDelta: m2.TotalAlloc - m1.TotalAlloc,
		MallocsDelta:    m2.Mallocs - m1.Mallocs,
		FreeDelta:       m2.Frees - m1.Frees,
		GCRuns:          m2.NumGC - m1.NumGC,
		Timestamp:       time.Now(),
	}
}

// MemoryProfile contains memory profiling information
type MemoryProfile struct {
	Name            string        `json:"name"`
	Duration        time.Duration `json:"duration"`
	AllocBefore     uint64        `json:"alloc_before"`
	AllocAfter      uint64        `json:"alloc_after"`
	AllocDelta      uint64        `json:"alloc_delta"`
	TotalAllocDelta uint64        `json:"total_alloc_delta"`
	MallocsDelta    uint64        `json:"mallocs_delta"`
	FreeDelta       uint64        `json:"frees_delta"`
	GCRuns          uint32        `json:"gc_runs"`
	Timestamp       time.Time     `json:"timestamp"`
}

// String returns a string representation of the memory profile
func (mp *MemoryProfile) String() string {
	return fmt.Sprintf(`Memory Profile for %s:
  Duration: %v
  Memory Delta: %d bytes
  Total Allocated: %d bytes
  Allocations: %d
  Frees: %d
  GC Runs: %d`,
		mp.Name,
		mp.Duration,
		mp.AllocDelta,
		mp.TotalAllocDelta,
		mp.MallocsDelta,
		mp.FreeDelta,
		mp.GCRuns)
}

// Global performance monitor instance
var defaultMonitor = NewPerformanceMonitor()

// Default functions for convenience
func RecordSearchOperation(duration time.Duration, resultCount int, cacheHit bool, queryLength int) {
	defaultMonitor.RecordSearchOperation(duration, resultCount, cacheHit, queryLength)
}

func RecordDatabaseOperation(operation string, duration time.Duration, success bool) {
	defaultMonitor.RecordDatabaseOperation(operation, duration, success)
}

func RecordMemoryUsage() {
	defaultMonitor.RecordMemoryUsage()
}

func GetPerformanceReport() PerformanceReport {
	return defaultMonitor.GetPerformanceReport()
}

func EnablePerformanceMonitoring(enabled bool) {
	defaultMonitor.Enable(enabled)
}

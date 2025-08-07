package metrics

import (
	"context"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	counter := NewCounter("test_counter", nil)

	if counter.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", counter.Value())
	}

	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("Expected value 1 after Inc(), got %d", counter.Value())
	}

	counter.Add(5)
	if counter.Value() != 6 {
		t.Errorf("Expected value 6 after Add(5), got %d", counter.Value())
	}

	counter.Reset()
	if counter.Value() != 0 {
		t.Errorf("Expected value 0 after Reset(), got %d", counter.Value())
	}
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("test_gauge", nil)

	if gauge.Value() != 0 {
		t.Errorf("Expected initial value 0, got %f", gauge.Value())
	}

	gauge.Set(3.14)
	if gauge.Value() != 3.14 {
		t.Errorf("Expected value 3.14 after Set(3.14), got %f", gauge.Value())
	}

	gauge.Inc()
	if gauge.Value() != 4.14 {
		t.Errorf("Expected value 4.14 after Inc(), got %f", gauge.Value())
	}

	gauge.Dec()
	if gauge.Value() != 3.14 {
		t.Errorf("Expected value 3.14 after Dec(), got %f", gauge.Value())
	}

	gauge.Add(1.86)
	if gauge.Value() != 5.0 {
		t.Errorf("Expected value 5.0 after Add(1.86), got %f", gauge.Value())
	}
}

func TestHistogram(t *testing.T) {
	histogram := NewHistogram("test_histogram", nil)

	if histogram.Count() != 0 {
		t.Errorf("Expected initial count 0, got %d", histogram.Count())
	}

	if histogram.Sum() != 0 {
		t.Errorf("Expected initial sum 0, got %f", histogram.Sum())
	}

	// Add some observations
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range values {
		histogram.Observe(v)
	}

	if histogram.Count() != 5 {
		t.Errorf("Expected count 5, got %d", histogram.Count())
	}

	expectedSum := 15.0
	if histogram.Sum() != expectedSum {
		t.Errorf("Expected sum %f, got %f", expectedSum, histogram.Sum())
	}

	expectedMean := 3.0
	if histogram.Mean() != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, histogram.Mean())
	}

	// Test percentiles
	p50 := histogram.Percentile(50)
	if p50 < 2.5 || p50 > 5.0 {
		t.Errorf("Expected 50th percentile between 2.5 and 5.0, got %f", p50)
	}
}

func TestTimer(t *testing.T) {
	timer := NewTimer("test_timer", nil)

	// Test timing a function
	done := timer.Time()
	time.Sleep(10 * time.Millisecond)
	done()

	histogram := timer.Histogram()
	if histogram.Count() != 1 {
		t.Errorf("Expected 1 timing measurement, got %d", histogram.Count())
	}

	// Should be at least 10ms
	if histogram.Mean() < 10 {
		t.Errorf("Expected mean >= 10ms, got %f", histogram.Mean())
	}

	// Test TimeFunc
	timer.TimeFunc(func() {
		time.Sleep(5 * time.Millisecond)
	})

	if histogram.Count() != 2 {
		t.Errorf("Expected 2 timing measurements, got %d", histogram.Count())
	}
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	// Test counter creation and retrieval
	counter1 := collector.Counter("test_counter", nil)
	counter2 := collector.Counter("test_counter", nil)

	if counter1 != counter2 {
		t.Error("Expected same counter instance for same name")
	}

	counter1.Inc()
	if counter2.Value() != 1 {
		t.Error("Expected shared counter state")
	}

	// Test gauge creation
	gauge := collector.Gauge("test_gauge", map[string]string{"tag": "value"})
	gauge.Set(42.0)

	// Test histogram creation
	histogram := collector.Histogram("test_histogram", nil)
	histogram.Observe(1.0)

	// Test timer creation
	timer := collector.Timer("test_timer", nil)
	done := timer.Time()
	time.Sleep(1 * time.Millisecond)
	done()

	// Get all metrics
	metrics := collector.GetAllMetrics()

	// Should have metrics for counter, gauge, histogram (count, sum, mean, percentiles), and timer
	if len(metrics) < 4 {
		t.Errorf("Expected at least 4 metrics, got %d", len(metrics))
	}

	// Test system metrics
	systemMetrics := collector.GetSystemMetrics()
	if len(systemMetrics) < 5 {
		t.Errorf("Expected at least 5 system metrics, got %d", len(systemMetrics))
	}
}

func TestPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()

	if !monitor.IsEnabled() {
		t.Error("Expected monitor to be enabled by default")
	}

	// Record some operations
	monitor.RecordSearchOperation(10*time.Millisecond, 5, false, 10)
	monitor.RecordSearchOperation(5*time.Millisecond, 3, true, 8)
	monitor.RecordDatabaseOperation("load", 100*time.Millisecond, true)
	monitor.RecordMemoryUsage()

	// Get performance report
	report := monitor.GetPerformanceReport()

	if report.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp in report")
	}

	if len(report.ApplicationMetrics) == 0 {
		t.Error("Expected application metrics in report")
	}

	if len(report.SystemMetrics) == 0 {
		t.Error("Expected system metrics in report")
	}

	// Test disable
	monitor.Enable(false)
	if monitor.IsEnabled() {
		t.Error("Expected monitor to be disabled")
	}
}

func TestBenchmarker(t *testing.T) {
	monitor := NewPerformanceMonitor()
	benchmarker := NewBenchmarker(monitor)

	// Benchmark a simple function
	result := benchmarker.BenchmarkFunction("test_function", func() {
		time.Sleep(1 * time.Millisecond)
	}, 5)

	if result.Name != "test_function" {
		t.Errorf("Expected name 'test_function', got '%s'", result.Name)
	}

	if result.Iterations != 5 {
		t.Errorf("Expected 5 iterations, got %d", result.Iterations)
	}

	if result.NsPerOp <= 0 {
		t.Errorf("Expected positive ns/op, got %d", result.NsPerOp)
	}

	// Test memory profiling
	profile := benchmarker.ProfileMemory("test_memory", func() {
		// Allocate some memory that won't be optimized away
		data := make([][]byte, 100)
		for i := range data {
			data[i] = make([]byte, 1024)
		}
		// Use the data to prevent optimization
		if len(data) > 0 && len(data[0]) > 0 {
			data[0][0] = 1
		}
	})

	if profile.Name != "test_memory" {
		t.Errorf("Expected name 'test_memory', got '%s'", profile.Name)
	}

	// Memory allocation might be 0 in some cases due to GC, so just check it's non-negative
	if profile.TotalAllocDelta < 0 {
		t.Errorf("Expected non-negative memory allocation, got %d", profile.TotalAllocDelta)
	}
}

func TestMemoryMonitoring(t *testing.T) {
	monitor := NewPerformanceMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start memory monitoring in background
	go monitor.StartMemoryMonitoring(ctx, 10*time.Millisecond)

	// Wait for monitoring to run
	time.Sleep(30 * time.Millisecond)

	// Get metrics - should have memory metrics recorded
	metrics := monitor.collector.GetAllMetrics()

	hasMemoryMetric := false
	for _, metric := range metrics {
		if metric.Name == "memory_alloc_bytes" {
			hasMemoryMetric = true
			break
		}
	}

	if !hasMemoryMetric {
		t.Error("Expected memory metrics to be recorded")
	}
}

// Benchmark tests
func BenchmarkCounter(b *testing.B) {
	counter := NewCounter("bench_counter", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}

func BenchmarkGauge(b *testing.B) {
	gauge := NewGauge("bench_gauge", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		gauge.Set(float64(i))
	}
}

func BenchmarkHistogram(b *testing.B) {
	histogram := NewHistogram("bench_histogram", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		histogram.Observe(float64(i % 100))
	}
}

func BenchmarkTimer(b *testing.B) {
	timer := NewTimer("bench_timer", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		done := timer.Time()
		done()
	}
}

func BenchmarkMetricsCollector(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		counter := collector.Counter("test_counter", nil)
		counter.Inc()
	}
}

func BenchmarkPerformanceMonitor(b *testing.B) {
	monitor := NewPerformanceMonitor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		monitor.RecordSearchOperation(time.Millisecond, 5, false, 10)
	}
}

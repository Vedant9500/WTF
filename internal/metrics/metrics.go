// Package metrics provides comprehensive performance monitoring and metrics collection.
//
// This package implements a complete metrics system including:
//   - Counters for monotonically increasing values
//   - Gauges for values that can go up and down
//   - Histograms for tracking value distributions
//   - Timers for measuring operation durations
//   - System metrics collection (memory, GC, goroutines)
//   - Thread-safe operations with atomic operations and mutexes
//
// The MetricsCollector provides a centralized way to manage all metrics,
// while individual metric types can be used independently.
//
// Example usage:
//
//	collector := NewMetricsCollector()
//	counter := collector.Counter("requests_total", map[string]string{"method": "GET"})
//	counter.Inc()
//
//	timer := collector.Timer("request_duration", nil)
//	defer timer.Time()()
//	// ... perform operation
package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
)

// Metric represents a single metric with metadata
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Description string            `json:"description"`
	Timestamp   time.Time         `json:"timestamp"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Counter represents a monotonically increasing counter
type Counter struct {
	value int64
	name  string
	tags  map[string]string
}

// NewCounter creates a new counter
func NewCounter(name string, tags map[string]string) *Counter {
	return &Counter{
		name: name,
		tags: tags,
	}
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds the given value to the counter
func (c *Counter) Add(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Value returns the current counter value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Reset resets the counter to 0
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Gauge represents a value that can go up and down
type Gauge struct {
	value int64 // Using int64 for atomic operations, will convert to float64
	name  string
	tags  map[string]string
}

// NewGauge creates a new gauge
func NewGauge(name string, tags map[string]string) *Gauge {
	return &Gauge{
		name: name,
		tags: tags,
	}
}

// Set sets the gauge to the given value
func (g *Gauge) Set(value float64) {
	atomic.StoreInt64(&g.value, int64(value*1000)) // Store as int64 with 3 decimal precision
}

// Value returns the current gauge value
func (g *Gauge) Value() float64 {
	return float64(atomic.LoadInt64(&g.value)) / 1000.0
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1000)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1000)
}

// Add adds the given value to the gauge
func (g *Gauge) Add(value float64) {
	atomic.AddInt64(&g.value, int64(value*1000))
}

// Histogram tracks the distribution of values
type Histogram struct {
	mu      sync.RWMutex
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
	name    string
	tags    map[string]string
}

// NewHistogram creates a new histogram with default buckets
func NewHistogram(name string, tags map[string]string) *Histogram {
	// Default buckets for response times (in milliseconds)
	buckets := []float64{0.1, 0.5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}
	return NewHistogramWithBuckets(name, buckets, tags)
}

// NewHistogramWithBuckets creates a new histogram with custom buckets
func NewHistogramWithBuckets(name string, buckets []float64, tags map[string]string) *Histogram {
	return &Histogram{
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1), // +1 for overflow bucket
		name:    name,
		tags:    tags,
	}
}

// Observe adds an observation to the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Find the appropriate bucket
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			return
		}
	}

	// Value is larger than all buckets, put in overflow bucket
	h.counts[len(h.buckets)]++
}

// Count returns the total number of observations
func (h *Histogram) Count() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Sum returns the sum of all observations
func (h *Histogram) Sum() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
}

// Mean returns the mean of all observations
func (h *Histogram) Mean() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// Percentile returns the value at the given percentile (0-100)
func (h *Histogram) Percentile(p float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return 0
	}

	target := int64(float64(h.count) * p / 100.0)
	cumulative := int64(0)

	for i, count := range h.counts {
		cumulative += count
		if cumulative >= target {
			if i < len(h.buckets) {
				return h.buckets[i]
			}
			// Overflow bucket - return the last bucket value
			return h.buckets[len(h.buckets)-1]
		}
	}

	return 0
}

// Timer measures elapsed time
type Timer struct {
	histogram *Histogram
	name      string
	tags      map[string]string
}

// NewTimer creates a new timer
func NewTimer(name string, tags map[string]string) *Timer {
	return &Timer{
		histogram: NewHistogram(name+"_duration", tags),
		name:      name,
		tags:      tags,
	}
}

// Time returns a function that should be called when the operation completes
func (t *Timer) Time() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		t.histogram.Observe(float64(duration.Nanoseconds()) / 1e6) // Convert to milliseconds
	}
}

// TimeFunc times the execution of a function
func (t *Timer) TimeFunc(fn func()) {
	defer t.Time()()
	fn()
}

// Histogram returns the underlying histogram
func (t *Timer) Histogram() *Histogram {
	return t.histogram
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	timers     map[string]*Timer
	startTime  time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
		startTime:  time.Now(),
	}
}

// Counter gets or creates a counter
func (mc *MetricsCollector) Counter(name string, tags map[string]string) *Counter {
	key := mc.metricKey(name, tags)

	mc.mu.RLock()
	if counter, exists := mc.counters[key]; exists {
		mc.mu.RUnlock()
		return counter
	}
	mc.mu.RUnlock()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Double-check after acquiring write lock
	if counter, exists := mc.counters[key]; exists {
		return counter
	}

	counter := NewCounter(name, tags)
	mc.counters[key] = counter
	return counter
}

// Gauge gets or creates a gauge
func (mc *MetricsCollector) Gauge(name string, tags map[string]string) *Gauge {
	key := mc.metricKey(name, tags)

	mc.mu.RLock()
	if gauge, exists := mc.gauges[key]; exists {
		mc.mu.RUnlock()
		return gauge
	}
	mc.mu.RUnlock()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if gauge, exists := mc.gauges[key]; exists {
		return gauge
	}

	gauge := NewGauge(name, tags)
	mc.gauges[key] = gauge
	return gauge
}

// Histogram gets or creates a histogram
func (mc *MetricsCollector) Histogram(name string, tags map[string]string) *Histogram {
	key := mc.metricKey(name, tags)

	mc.mu.RLock()
	if histogram, exists := mc.histograms[key]; exists {
		mc.mu.RUnlock()
		return histogram
	}
	mc.mu.RUnlock()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if histogram, exists := mc.histograms[key]; exists {
		return histogram
	}

	histogram := NewHistogram(name, tags)
	mc.histograms[key] = histogram
	return histogram
}

// Timer gets or creates a timer
func (mc *MetricsCollector) Timer(name string, tags map[string]string) *Timer {
	key := mc.metricKey(name, tags)

	mc.mu.RLock()
	if timer, exists := mc.timers[key]; exists {
		mc.mu.RUnlock()
		return timer
	}
	mc.mu.RUnlock()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if timer, exists := mc.timers[key]; exists {
		return timer
	}

	timer := NewTimer(name, tags)
	mc.timers[key] = timer
	return timer
}

// GetAllMetrics returns all current metrics
func (mc *MetricsCollector) GetAllMetrics() []Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var metrics []Metric
	now := time.Now()

	// Add counters
	for _, counter := range mc.counters {
		metrics = append(metrics, Metric{
			Name:      counter.name,
			Type:      MetricTypeCounter,
			Value:     float64(counter.Value()),
			Unit:      "count",
			Timestamp: now,
			Tags:      counter.tags,
		})
	}

	// Add gauges
	for _, gauge := range mc.gauges {
		metrics = append(metrics, Metric{
			Name:      gauge.name,
			Type:      MetricTypeGauge,
			Value:     gauge.Value(),
			Unit:      "value",
			Timestamp: now,
			Tags:      gauge.tags,
		})
	}

	// Add histograms
	for _, histogram := range mc.histograms {
		metrics = append(metrics, Metric{
			Name:      histogram.name + "_count",
			Type:      MetricTypeHistogram,
			Value:     float64(histogram.Count()),
			Unit:      "count",
			Timestamp: now,
			Tags:      histogram.tags,
		})

		metrics = append(metrics, Metric{
			Name:      histogram.name + "_sum",
			Type:      MetricTypeHistogram,
			Value:     histogram.Sum(),
			Unit:      "ms",
			Timestamp: now,
			Tags:      histogram.tags,
		})

		metrics = append(metrics, Metric{
			Name:      histogram.name + "_mean",
			Type:      MetricTypeHistogram,
			Value:     histogram.Mean(),
			Unit:      "ms",
			Timestamp: now,
			Tags:      histogram.tags,
		})

		// Add percentiles
		for _, p := range []float64{50, 90, 95, 99} {
			metrics = append(metrics, Metric{
				Name:      histogram.name + "_p" + fmt.Sprintf("%.0f", p),
				Type:      MetricTypeHistogram,
				Value:     histogram.Percentile(p),
				Unit:      "ms",
				Timestamp: now,
				Tags:      histogram.tags,
			})
		}
	}

	return metrics
}

// GetSystemMetrics returns system-level metrics
func (mc *MetricsCollector) GetSystemMetrics() []Metric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	uptime := now.Sub(mc.startTime)

	return []Metric{
		{
			Name:        "system_memory_alloc",
			Type:        MetricTypeGauge,
			Value:       float64(m.Alloc),
			Unit:        "bytes",
			Timestamp:   now,
			Description: "Bytes allocated and still in use",
		},
		{
			Name:        "system_memory_total_alloc",
			Type:        MetricTypeCounter,
			Value:       float64(m.TotalAlloc),
			Unit:        "bytes",
			Timestamp:   now,
			Description: "Cumulative bytes allocated",
		},
		{
			Name:        "system_memory_sys",
			Type:        MetricTypeGauge,
			Value:       float64(m.Sys),
			Unit:        "bytes",
			Timestamp:   now,
			Description: "Total bytes obtained from OS",
		},
		{
			Name:        "system_gc_runs",
			Type:        MetricTypeCounter,
			Value:       float64(m.NumGC),
			Unit:        "count",
			Timestamp:   now,
			Description: "Number of GC runs",
		},
		{
			Name:        "system_goroutines",
			Type:        MetricTypeGauge,
			Value:       float64(runtime.NumGoroutine()),
			Unit:        "count",
			Timestamp:   now,
			Description: "Number of goroutines",
		},
		{
			Name:        "system_uptime",
			Type:        MetricTypeGauge,
			Value:       uptime.Seconds(),
			Unit:        "seconds",
			Timestamp:   now,
			Description: "Application uptime",
		},
	}
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counters = make(map[string]*Counter)
	mc.gauges = make(map[string]*Gauge)
	mc.histograms = make(map[string]*Histogram)
	mc.timers = make(map[string]*Timer)
	mc.startTime = time.Now()
}

// metricKey generates a unique key for a metric with tags
func (mc *MetricsCollector) metricKey(name string, tags map[string]string) string {
	if len(tags) == 0 {
		return name
	}

	key := name
	for k, v := range tags {
		key += ":" + k + "=" + v
	}
	return key
}

// Global metrics collector instance
var defaultCollector = NewMetricsCollector()

// Default functions for convenience
func DefaultCounter(name string, tags map[string]string) *Counter {
	return defaultCollector.Counter(name, tags)
}

func DefaultGauge(name string, tags map[string]string) *Gauge {
	return defaultCollector.Gauge(name, tags)
}

func DefaultHistogram(name string, tags map[string]string) *Histogram {
	return defaultCollector.Histogram(name, tags)
}

func DefaultTimer(name string, tags map[string]string) *Timer {
	return defaultCollector.Timer(name, tags)
}

func GetAllMetrics() []Metric {
	return defaultCollector.GetAllMetrics()
}

func GetSystemMetrics() []Metric {
	return defaultCollector.GetSystemMetrics()
}

func ResetMetrics() {
	defaultCollector.Reset()
}

package database

import (
	"fmt"
	"time"

	"github.com/Vedant9500/WTF/internal/cache"
	"github.com/Vedant9500/WTF/internal/metrics"
)

// MonitoredDatabase wraps Database with performance monitoring
type MonitoredDatabase struct {
	*CachedDatabase
	monitor *metrics.PerformanceMonitor
}

// NewMonitoredDatabase creates a new database with performance monitoring
func NewMonitoredDatabase(db *Database) *MonitoredDatabase {
	cachedDB := NewCachedDatabase(db)
	return &MonitoredDatabase{
		CachedDatabase: cachedDB,
		monitor:        metrics.NewPerformanceMonitor(),
	}
}

// SearchWithMonitoring performs search with performance monitoring
func (mdb *MonitoredDatabase) SearchWithMonitoring(query string, limit int) []SearchResult {
	start := time.Now()
	
	// Check if result will come from cache
	searchCache := mdb.cacheManager.GetSearchCache()
	cacheOptions := mdb.convertToCacheOptions(SearchOptions{Limit: limit})
	_, cacheHit := searchCache.Get(query, cacheOptions)
	
	// Perform the search
	results := mdb.SearchWithCache(query, limit)
	
	// Record metrics
	duration := time.Since(start)
	mdb.monitor.RecordSearchOperation(duration, len(results), cacheHit, len(query))
	
	return results
}

// SearchWithOptionsAndMonitoring performs search with options and monitoring
func (mdb *MonitoredDatabase) SearchWithOptionsAndMonitoring(query string, options SearchOptions) []SearchResult {
	start := time.Now()
	
	// Check cache hit
	searchCache := mdb.cacheManager.GetSearchCache()
	cacheOptions := mdb.convertToCacheOptions(options)
	_, cacheHit := searchCache.Get(query, cacheOptions)
	
	// Perform the search
	results := mdb.SearchWithOptionsAndCache(query, options)
	
	// Record metrics
	duration := time.Since(start)
	mdb.monitor.RecordSearchOperation(duration, len(results), cacheHit, len(query))
	
	return results
}

// LoadDatabaseWithMonitoring loads database with performance monitoring
func (mdb *MonitoredDatabase) LoadDatabaseWithMonitoring(commands []Command) error {
	start := time.Now()
	
	// Update the database
	mdb.UpdateDatabase(commands)
	
	// Record metrics
	duration := time.Since(start)
	mdb.monitor.RecordDatabaseOperation("load", duration, true)
	
	return nil
}

// GetPerformanceReport returns a performance report
func (mdb *MonitoredDatabase) GetPerformanceReport() metrics.PerformanceReport {
	return mdb.monitor.GetPerformanceReport()
}

// EnableMonitoring enables or disables performance monitoring
func (mdb *MonitoredDatabase) EnableMonitoring(enabled bool) {
	mdb.monitor.Enable(enabled)
}

// IsMonitoringEnabled returns whether monitoring is enabled
func (mdb *MonitoredDatabase) IsMonitoringEnabled() bool {
	return mdb.monitor.IsEnabled()
}

// RecordMemoryUsage records current memory usage
func (mdb *MonitoredDatabase) RecordMemoryUsage() {
	mdb.monitor.RecordMemoryUsage()
}

// BenchmarkSearch performs a benchmark of search operations
func (mdb *MonitoredDatabase) BenchmarkSearch(queries []string, iterations int) []metrics.BenchmarkResult {
	benchmarker := metrics.NewBenchmarker(mdb.monitor)
	results := make([]metrics.BenchmarkResult, len(queries))
	
	for i, query := range queries {
		results[i] = benchmarker.BenchmarkFunction(
			"search_"+query,
			func() {
				mdb.SearchWithMonitoring(query, 10)
			},
			iterations,
		)
	}
	
	return results
}

// ProfileSearchMemory profiles memory usage during search
func (mdb *MonitoredDatabase) ProfileSearchMemory(query string) metrics.MemoryProfile {
	benchmarker := metrics.NewBenchmarker(mdb.monitor)
	
	return benchmarker.ProfileMemory("search_memory_"+query, func() {
		mdb.SearchWithMonitoring(query, 10)
	})
}

// convertToCacheOptions converts SearchOptions to cache.SearchOptions
func (mdb *MonitoredDatabase) convertToCacheOptions(options SearchOptions) cache.SearchOptions {
	return cache.SearchOptions{
		Limit:          options.Limit,
		ContextBoosts:  options.ContextBoosts,
		PipelineOnly:   options.PipelineOnly,
		PipelineBoost:  options.PipelineBoost,
		UseFuzzy:       options.UseFuzzy,
		FuzzyThreshold: options.FuzzyThreshold,
		UseNLP:         options.UseNLP,
	}
}

// SearchPerformanceAnalyzer analyzes search performance patterns
type SearchPerformanceAnalyzer struct {
	database *MonitoredDatabase
	results  []SearchAnalysisResult
}

// SearchAnalysisResult contains analysis of search performance
type SearchAnalysisResult struct {
	Query           string        `json:"query"`
	AverageTime     time.Duration `json:"average_time"`
	MinTime         time.Duration `json:"min_time"`
	MaxTime         time.Duration `json:"max_time"`
	CacheHitRatio   float64       `json:"cache_hit_ratio"`
	ResultCount     int           `json:"result_count"`
	Iterations      int           `json:"iterations"`
}

// NewSearchPerformanceAnalyzer creates a new search performance analyzer
func NewSearchPerformanceAnalyzer(database *MonitoredDatabase) *SearchPerformanceAnalyzer {
	return &SearchPerformanceAnalyzer{
		database: database,
		results:  make([]SearchAnalysisResult, 0),
	}
}

// AnalyzeQuery analyzes the performance of a specific query
func (spa *SearchPerformanceAnalyzer) AnalyzeQuery(query string, iterations int) SearchAnalysisResult {
	times := make([]time.Duration, iterations)
	var totalResults int
	cacheHits := 0
	
	// Clear cache to get accurate measurements
	spa.database.InvalidateCache()
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		// Check if this will be a cache hit (after first iteration)
		if i > 0 {
			searchCache := spa.database.cacheManager.GetSearchCache()
			cacheOptions := spa.database.convertToCacheOptions(SearchOptions{Limit: 10})
			if _, hit := searchCache.Get(query, cacheOptions); hit {
				cacheHits++
			}
		}
		
		results := spa.database.SearchWithMonitoring(query, 10)
		times[i] = time.Since(start)
		
		if i == 0 {
			totalResults = len(results)
		}
	}
	
	// Calculate statistics
	var totalTime time.Duration
	minTime := times[0]
	maxTime := times[0]
	
	for _, t := range times {
		totalTime += t
		if t < minTime {
			minTime = t
		}
		if t > maxTime {
			maxTime = t
		}
	}
	
	result := SearchAnalysisResult{
		Query:         query,
		AverageTime:   totalTime / time.Duration(iterations),
		MinTime:       minTime,
		MaxTime:       maxTime,
		CacheHitRatio: float64(cacheHits) / float64(iterations-1), // Exclude first iteration
		ResultCount:   totalResults,
		Iterations:    iterations,
	}
	
	spa.results = append(spa.results, result)
	return result
}

// AnalyzeQueries analyzes multiple queries
func (spa *SearchPerformanceAnalyzer) AnalyzeQueries(queries []string, iterations int) []SearchAnalysisResult {
	results := make([]SearchAnalysisResult, len(queries))
	
	for i, query := range queries {
		results[i] = spa.AnalyzeQuery(query, iterations)
	}
	
	return results
}

// GetResults returns all analysis results
func (spa *SearchPerformanceAnalyzer) GetResults() []SearchAnalysisResult {
	return spa.results
}

// GenerateReport generates a performance analysis report
func (spa *SearchPerformanceAnalyzer) GenerateReport() string {
	if len(spa.results) == 0 {
		return "No analysis results available"
	}
	
	report := "Search Performance Analysis Report\n"
	report += "==================================\n\n"
	
	for _, result := range spa.results {
		report += fmt.Sprintf("Query: %s\n", result.Query)
		report += fmt.Sprintf("  Average Time: %v\n", result.AverageTime)
		report += fmt.Sprintf("  Min Time: %v\n", result.MinTime)
		report += fmt.Sprintf("  Max Time: %v\n", result.MaxTime)
		report += fmt.Sprintf("  Cache Hit Ratio: %.2f%%\n", result.CacheHitRatio*100)
		report += fmt.Sprintf("  Result Count: %d\n", result.ResultCount)
		report += fmt.Sprintf("  Iterations: %d\n\n", result.Iterations)
	}
	
	return report
}
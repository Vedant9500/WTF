package database

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkSearch benchmarks the basic search functionality
func BenchmarkSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000) // 1000 commands for realistic testing

	queries := []string{
		"git commit",
		"find files",
		"tar compress",
		"docker run",
		"npm install",
		"mkdir directory",
		"grep search",
		"curl download",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		results := db.Search(query, 10)
		_ = results // Prevent optimization
	}
}

// BenchmarkSearchWithOptions benchmarks search with various options
func BenchmarkSearchWithOptions(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	db.BuildUniversalIndex() // Build universal index for benchmarking

	options := SearchOptions{
		Limit:         10,
		ContextBoosts: map[string]float64{"git": 2.0, "docker": 1.5},
		PipelineOnly:  false,
		UseFuzzy:      false,
		UseNLP:        false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results := db.SearchUniversal("git commit", options)
		_ = results
	}
}

// BenchmarkCalculateScore benchmarks the scoring algorithm
func BenchmarkCalculateScore(b *testing.B) {
	cmd := &Command{
		Command:          "git commit -m 'message'",
		Description:      "commit changes with message",
		Keywords:         []string{"git", "commit", "version-control"},
		CommandLower:     "git commit -m 'message'",
		DescriptionLower: "commit changes with message",
		KeywordsLower:    []string{"git", "commit", "version-control"},
	}

	queryWords := []string{"git", "commit"}
	contextBoosts := map[string]float64{"git": 2.0}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		score := calculateScore(cmd, queryWords, contextBoosts)
		_ = score
	}
}

// BenchmarkCalculateWordScore benchmarks individual word scoring
func BenchmarkCalculateWordScore(b *testing.B) {
	cmd := &Command{
		Command:          "git commit -m 'message'",
		Description:      "commit changes with message",
		Keywords:         []string{"git", "commit", "version-control"},
		Tags:             []string{"vcs", "development"},
		CommandLower:     "git commit -m 'message'",
		DescriptionLower: "commit changes with message",
		KeywordsLower:    []string{"git", "commit", "version-control"},
		TagsLower:        []string{"vcs", "development"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		score := calculateWordScore("git", cmd)
		_ = score
	}
}

// BenchmarkStringOperations benchmarks string operations in hot paths
func BenchmarkStringOperations(b *testing.B) {
	text := "git commit -m 'initial commit'"

	b.Run("ToLower", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := strings.ToLower(text)
			_ = result
		}
	})

	b.Run("Fields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := strings.Fields(text)
			_ = result
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := strings.Contains(text, "commit")
			_ = result
		}
	})
}

// BenchmarkMemoryAllocations benchmarks memory allocation patterns
func BenchmarkMemoryAllocations(b *testing.B) {
	db := createBenchmarkDatabase(100)

	b.Run("SearchResults", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]SearchResult, 0, 50)
			for j := 0; j < 10; j++ {
				results = append(results, SearchResult{
					Command: &db.Commands[j%len(db.Commands)],
					Score:   float64(j),
				})
			}
			_ = results
		}
	})

	b.Run("QueryWords", func(b *testing.B) {
		query := "git commit -m message"
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			words := strings.Fields(strings.ToLower(query))
			_ = words
		}
	})
}

// BenchmarkLargeDatabase benchmarks performance with large datasets
func BenchmarkLargeDatabase(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			db := createBenchmarkDatabase(size)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				results := db.Search("git commit", 10)
				_ = results
			}
		})
	}
}

// BenchmarkConcurrentSearch benchmarks concurrent search operations
func BenchmarkConcurrentSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	queries := []string{
		"git commit", "find files", "tar compress", "docker run",
		"npm install", "mkdir directory", "grep search", "curl download",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			query := queries[i%len(queries)]
			results := db.Search(query, 10)
			_ = results
			i++
		}
	})
}

// createBenchmarkDatabase creates a database with specified number of commands for benchmarking
func createBenchmarkDatabase(size int) *Database {
	commands := make([]Command, size)

	// Base command templates for realistic data
	templates := []struct {
		command     string
		description string
		keywords    []string
		tags        []string
	}{
		{"git commit -m '%s'", "commit changes with message", []string{"git", "commit", "version-control"}, []string{"vcs", "development"}},
		{"find . -name '%s'", "find files by name", []string{"find", "search", "files"}, []string{"filesystem", "search"}},
		{"tar -czf %s.tar.gz .", "create compressed archive", []string{"tar", "compress", "archive"}, []string{"compression", "backup"}},
		{"docker run -d %s", "run docker container", []string{"docker", "container", "run"}, []string{"containerization", "deployment"}},
		{"npm install %s", "install npm package", []string{"npm", "install", "package"}, []string{"nodejs", "package-manager"}},
		{"mkdir -p %s", "create directory recursively", []string{"mkdir", "directory", "create"}, []string{"filesystem", "directory"}},
		{"grep -r '%s' .", "search text in files", []string{"grep", "search", "text"}, []string{"search", "text-processing"}},
		{"curl -O %s", "download file with curl", []string{"curl", "download", "http"}, []string{"network", "download"}},
		{"ssh user@%s", "connect via SSH", []string{"ssh", "remote", "connect"}, []string{"network", "remote-access"}},
		{"cp -r %s %s", "copy files recursively", []string{"cp", "copy", "files"}, []string{"filesystem", "file-operations"}},
	}

	for i := 0; i < size; i++ {
		template := templates[i%len(templates)]
		suffix := fmt.Sprintf("item%d", i)

		cmd := Command{
			Command:     fmt.Sprintf(template.command, suffix),
			Description: template.description,
			Keywords:    template.keywords,
			Tags:        template.tags,
		}

		// Pre-compute lowercased fields for performance
		cmd.CommandLower = strings.ToLower(cmd.Command)
		cmd.DescriptionLower = strings.ToLower(cmd.Description)
		cmd.KeywordsLower = make([]string, len(cmd.Keywords))
		for j, keyword := range cmd.Keywords {
			cmd.KeywordsLower[j] = strings.ToLower(keyword)
		}
		cmd.TagsLower = make([]string, len(cmd.Tags))
		for j, tag := range cmd.Tags {
			cmd.TagsLower[j] = strings.ToLower(tag)
		}

		commands[i] = cmd
	}

	db := &Database{Commands: commands}
	db.BuildUniversalIndex() // Build universal index for benchmarking
	return db
}

// BenchmarkSearchResultsAllocation benchmarks different allocation strategies
func BenchmarkSearchResultsAllocation(b *testing.B) {
	db := createBenchmarkDatabase(1000)

	b.Run("PreAllocated", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]SearchResult, 0, 50) // Pre-allocate capacity
			for j := 0; j < 10 && j < len(db.Commands); j++ {
				results = append(results, SearchResult{
					Command: &db.Commands[j],
					Score:   float64(j),
				})
			}
			_ = results
		}
	})

	b.Run("GrowingSlice", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var results []SearchResult // No pre-allocation
			for j := 0; j < 10 && j < len(db.Commands); j++ {
				results = append(results, SearchResult{
					Command: &db.Commands[j],
					Score:   float64(j),
				})
			}
			_ = results
		}
	})
}

// BenchmarkMemoryProfile runs a memory-intensive search operation for profiling
func BenchmarkMemoryProfile(b *testing.B) {
	db := createBenchmarkDatabase(5000)

	queries := []string{
		"git commit message",
		"find text files recursively",
		"compress archive with tar",
		"docker container management",
		"npm package installation",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, query := range queries {
			results := db.SearchUniversal(query, SearchOptions{
				Limit:         20,
				ContextBoosts: map[string]float64{"git": 2.0, "docker": 1.5},
				UseFuzzy:      false,
				UseNLP:        false,
			})
			_ = results
		}
	}
}

// BenchmarkOptimizedSearch benchmarks the optimized search functionality
func BenchmarkOptimizedSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000)

	queries := []string{
		"git commit",
		"find files",
		"tar compress",
		"docker run",
		"npm install",
		"mkdir directory",
		"grep search",
		"curl download",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		results := db.SearchUniversal(query, SearchOptions{Limit: 10})
		_ = results
	}
}

// BenchmarkSearchComparison compares original vs optimized search
func BenchmarkSearchComparison(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	query := "git commit"

	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := db.Search(query, 10)
			_ = results
		}
	})

	b.Run("Universal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := db.SearchUniversal(query, SearchOptions{Limit: 10})
			_ = results
		}
	})
}

// BenchmarkQueryParsing compares query parsing methods
func BenchmarkQueryParsing(b *testing.B) {
	query := "git commit -m message"

	b.Run("StringsFields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			words := strings.Fields(strings.ToLower(query))
			_ = words
		}
	})

	b.Run("StringsFieldsFunc", func(b *testing.B) {
		words := make([]string, 0, 10)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			words = words[:0]
			queryLower := strings.ToLower(query)
			words = strings.Fields(queryLower)
			_ = words
		}
	})
}

// BenchmarkObjectPools benchmarks object pool performance
func BenchmarkObjectPools(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]SearchResult, 0, 50)
			for j := 0; j < 10; j++ {
				results = append(results, SearchResult{Score: float64(j)})
			}
			_ = results
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]SearchResult, 0, 50) // Simulate pool allocation
			for j := 0; j < 10; j++ {
				results = append(results, SearchResult{Score: float64(j)})
			}
			_ = results // Simulate pool return
		}
	})
}

// BenchmarkBatchSearch benchmarks batch search operations
func BenchmarkBatchSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	queries := []string{
		"git commit",
		"find files",
		"tar compress",
		"docker run",
		"npm install",
	}

	b.Run("Individual", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, query := range queries {
				results := db.SearchUniversal(query, SearchOptions{Limit: 10})
				_ = results
			}
		}
	})

	b.Run("BatchSimulation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate batch processing with universal search
			allResults := make([][]SearchResult, 0, len(queries))
			for _, query := range queries {
				results := db.SearchUniversal(query, SearchOptions{Limit: 10})
				allResults = append(allResults, results)
			}
			_ = allResults
		}
	})
}

// BenchmarkCachedSearch benchmarks search with caching enabled
func BenchmarkCachedSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	cachedDB := NewCachedDatabase(db)

	queries := []string{
		"git commit",
		"find files",
		"tar compress",
		"docker run",
		"npm install",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		results := cachedDB.SearchWithCache(query, 10)
		_ = results
	}
}

// BenchmarkCacheComparison compares cached vs non-cached search
func BenchmarkCacheComparison(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	cachedDB := NewCachedDatabase(db)
	query := "git commit"

	b.Run("WithoutCache", func(b *testing.B) {
		cachedDB.EnableCache(false)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := cachedDB.SearchWithCache(query, 10)
			_ = results
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cachedDB.EnableCache(true)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := cachedDB.SearchWithCache(query, 10)
			_ = results
		}
	})
}

// BenchmarkCacheHitRatio benchmarks cache performance with different hit ratios
func BenchmarkCacheHitRatio(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	cachedDB := NewCachedDatabase(db)

	// Queries with different repetition patterns
	queries := []string{
		"git commit", "git commit", "git commit", // High repetition
		"find files", "find files", // Medium repetition
		"tar compress", // Low repetition
		"docker run",
		"npm install",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		results := cachedDB.SearchWithCache(query, 10)
		_ = results
	}
}

// BenchmarkCacheMemoryUsage benchmarks memory usage with caching
func BenchmarkCacheMemoryUsage(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	cachedDB := NewCachedDatabase(db)

	// Use many different queries to fill cache
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("query%d", i%100) // 100 different queries
		results := cachedDB.SearchWithCache(query, 10)
		_ = results
	}
}

// BenchmarkMonitoredSearch benchmarks search with performance monitoring
func BenchmarkMonitoredSearch(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	monitoredDB := NewMonitoredDatabase(db)

	queries := []string{
		"git commit",
		"find files",
		"tar compress",
		"docker run",
		"npm install",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		results := monitoredDB.SearchWithMonitoring(query, 10)
		_ = results
	}
}

// BenchmarkMonitoringOverhead compares search with and without monitoring
func BenchmarkMonitoringOverhead(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	cachedDB := NewCachedDatabase(db)
	monitoredDB := NewMonitoredDatabase(db)
	query := "git commit"

	b.Run("WithoutMonitoring", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := cachedDB.SearchWithCache(query, 10)
			_ = results
		}
	})

	b.Run("WithMonitoring", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := monitoredDB.SearchWithMonitoring(query, 10)
			_ = results
		}
	})
}

// BenchmarkPerformanceReport benchmarks performance report generation
func BenchmarkPerformanceReport(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	monitoredDB := NewMonitoredDatabase(db)

	// Generate some metrics
	for i := 0; i < 100; i++ {
		monitoredDB.SearchWithMonitoring("test query", 10)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		report := monitoredDB.GetPerformanceReport()
		_ = report
	}
}

// BenchmarkSearchAnalysis benchmarks search performance analysis
func BenchmarkSearchAnalysis(b *testing.B) {
	db := createBenchmarkDatabase(1000)
	monitoredDB := NewMonitoredDatabase(db)
	analyzer := NewSearchPerformanceAnalyzer(monitoredDB)

	queries := []string{"git commit", "find files", "tar compress"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		result := analyzer.AnalyzeQuery(query, 10)
		_ = result
	}
}

package database

import (
	"testing"

	"github.com/Vedant9500/WTF/internal/testutil"
)

// BenchmarkSearch benchmarks the search functionality
func BenchmarkSearch(b *testing.B) {
	// Create a large test database
	db := testutil.CreateLargeDatabase(1000)

	// Build the index once
	db.BuildUniversalIndex()

	testQueries := []string{
		"copy files",
		"git commit",
		"find text",
		"compress directory",
		"network configuration",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := testQueries[i%len(testQueries)]
		results := db.SearchUniversal(query, SearchOptions{
			Limit:    10,
			UseNLP:   true,
			UseFuzzy: true,
		})

		// Prevent optimization from eliminating the search
		_ = len(results)
	}
}

// BenchmarkSearchWithoutNLP benchmarks search without NLP processing
func BenchmarkSearchWithoutNLP(b *testing.B) {
	db := testutil.CreateLargeDatabase(1000)
	db.BuildUniversalIndex()

	testQueries := []string{
		"copy files",
		"git commit",
		"find text",
		"compress directory",
		"network configuration",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := testQueries[i%len(testQueries)]
		results := db.SearchUniversal(query, SearchOptions{
			Limit:    10,
			UseNLP:   false,
			UseFuzzy: false,
		})

		_ = len(results)
	}
}

// BenchmarkIndexBuilding benchmarks the index building process
func BenchmarkIndexBuilding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		db := testutil.CreateLargeDatabase(500)
		b.StartTimer()

		db.BuildUniversalIndex()
	}
}

// BenchmarkDatabaseLoad benchmarks database loading
func BenchmarkDatabaseLoad(b *testing.B) {
	// Create a test database file
	tempDir, cleanup := testutil.CreateTempDir()
	defer cleanup()

	testDB := testutil.CreateLargeDatabase(100)
	dbPath := tempDir + "/benchmark.yml"

	err := testutil.SaveDatabase(testDB, dbPath)
	if err != nil {
		b.Fatalf("Failed to save test database: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := LoadDatabase(dbPath)
		if err != nil {
			b.Fatalf("Failed to load database: %v", err)
		}
	}
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	db := testutil.CreateLargeDatabase(2000)

	// Force garbage collection before starting
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate realistic usage pattern
		db.BuildUniversalIndex()

		// Perform multiple searches
		for j := 0; j < 10; j++ {
			results := db.SearchUniversal("test query", SearchOptions{
				Limit:    10,
				UseNLP:   true,
				UseFuzzy: true,
			})
			_ = len(results)
		}
	}
}

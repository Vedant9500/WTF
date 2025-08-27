package database

import (
	"testing"
)

// createLargeTestDatabase creates a large database for benchmarking without testutil
func createLargeTestDatabase(size int) *Database {
	commands := make([]Command, size)
	
	// Create diverse test commands
	templates := []struct {
		cmd, desc string
		keywords  []string
		platform  []string
	}{
		{"git commit -m '%s'", "commit changes with message", []string{"git", "commit", "message"}, []string{"linux", "macos", "windows"}},
		{"find . -name '%s'", "find files by name", []string{"find", "search", "files"}, []string{"linux", "macos"}},
		{"tar -czf %s.tar.gz %s", "compress directory", []string{"tar", "compress", "archive"}, []string{"linux", "macos"}},
		{"docker run -d %s", "run docker container", []string{"docker", "container", "run"}, []string{"linux", "macos", "windows"}},
		{"npm install %s", "install npm package", []string{"npm", "install", "package"}, []string{"linux", "macos", "windows"}},
		{"mkdir -p %s", "create directory", []string{"mkdir", "directory", "create"}, []string{"linux", "macos"}},
		{"grep -r '%s' .", "search for text recursively", []string{"grep", "search", "text"}, []string{"linux", "macos"}},
		{"curl -O %s", "download file", []string{"curl", "download", "http"}, []string{"linux", "macos"}},
		{"ps aux | grep %s", "find process", []string{"ps", "process", "find"}, []string{"linux", "macos"}},
		{"netstat -tulpn", "show network connections", []string{"netstat", "network", "connections"}, []string{"linux", "macos"}},
	}
	
	for i := 0; i < size; i++ {
		template := templates[i%len(templates)]
		commands[i] = Command{
			Command:     template.cmd,
			Description: template.desc,
			Keywords:    template.keywords,
			Platform:    template.platform,
			Pipeline:    i%3 == 0, // every 3rd command is pipeline
		}
	}
	
	return &Database{Commands: commands}
}

// BenchmarkSearchLarge benchmarks the search functionality with large dataset
func BenchmarkSearchLarge(b *testing.B) {
	// Create a large test database
	db := createLargeTestDatabase(1000)

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
	// Create a large test database
	db := createLargeTestDatabase(1000)
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
		db := createLargeTestDatabase(500)
		b.StartTimer()

		db.BuildUniversalIndex()
	}
}

// BenchmarkDatabaseLoad benchmarks database loading
func BenchmarkDatabaseLoad(b *testing.B) {
	// Skip this benchmark since it requires file I/O setup without testutil
	b.Skip("Skipping file I/O benchmark to avoid testutil dependency")
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	db := createLargeTestDatabase(2000)

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

package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Vedant9500/WTF/internal/database"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// WithTimeout runs a test function with a timeout
func (th *TestHelper) WithTimeout(timeout time.Duration, testFunc func()) {
	done := make(chan bool, 1)

	go func() {
		testFunc()
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		th.t.Fatalf("Test timed out after %v", timeout)
	}
}

// ExpectPanic expects a function to panic
func (th *TestHelper) ExpectPanic(testFunc func()) {
	defer func() {
		if r := recover(); r == nil {
			th.t.Error("Expected function to panic, but it didn't")
		}
	}()
	testFunc()
}

// ExpectNoPanic expects a function not to panic
func (th *TestHelper) ExpectNoPanic(testFunc func()) {
	defer func() {
		if r := recover(); r != nil {
			th.t.Errorf("Expected function not to panic, but it panicked with: %v", r)
		}
	}()
	testFunc()
}

// FileHelper provides file-related testing utilities
type FileHelper struct {
	tempFiles []string
	tempDirs  []string
}

// NewFileHelper creates a new file helper
func NewFileHelper() *FileHelper {
	return &FileHelper{
		tempFiles: make([]string, 0),
		tempDirs:  make([]string, 0),
	}
}

// CreateTempFile creates a temporary file and tracks it for cleanup
func (fh *FileHelper) CreateTempFile(content string) string {
	tempFile, cleanup := NewTestFixtures().CreateTempFile(content)
	fh.tempFiles = append(fh.tempFiles, tempFile)
	// Store cleanup function reference (in real implementation, we'd track this)
	_ = cleanup
	return tempFile
}

// CreateTempDir creates a temporary directory and tracks it for cleanup
func (fh *FileHelper) CreateTempDir() string {
	tempDir, cleanup := NewTestFixtures().CreateTempDir()
	fh.tempDirs = append(fh.tempDirs, tempDir)
	// Store cleanup function reference (in real implementation, we'd track this)
	_ = cleanup
	return tempDir
}

// Cleanup removes all temporary files and directories
func (fh *FileHelper) Cleanup() {
	for _, file := range fh.tempFiles {
		os.Remove(file)
	}
	for _, dir := range fh.tempDirs {
		os.RemoveAll(dir)
	}
	fh.tempFiles = fh.tempFiles[:0]
	fh.tempDirs = fh.tempDirs[:0]
}

// DatabaseTestHelper provides database-specific testing utilities
type DatabaseTestHelper struct {
	testDB TestDatabase
}

// NewDatabaseTestHelper creates a new database test helper
func NewDatabaseTestHelper() *DatabaseTestHelper {
	return &DatabaseTestHelper{
		testDB: NewTestDatabase(),
	}
}

// CreateTestDatabase creates a test database with custom commands
func (dth *DatabaseTestHelper) CreateTestDatabase(commands []database.Command) *database.Database {
	return dth.testDB.CreateTestDB(commands)
}

// CreateMinimalDatabase creates a minimal test database
func (dth *DatabaseTestHelper) CreateMinimalDatabase() *database.Database {
	return dth.testDB.CreateMinimalDB()
}

// CreateLargeDatabase creates a large test database for performance testing
func (dth *DatabaseTestHelper) CreateLargeDatabase() *database.Database {
	return dth.testDB.CreateLargeDB()
}

// CreateEmptyDatabase creates an empty database for edge case testing
func (dth *DatabaseTestHelper) CreateEmptyDatabase() *database.Database {
	return dth.testDB.CreateEmptyDB()
}

// ValidateSearchResults validates search results against expected criteria
func (dth *DatabaseTestHelper) ValidateSearchResults(t *testing.T, results []database.SearchResult, query TestQuery) {
	t.Helper()

	// Check result count
	AssertResultCount(t, results, query.ExpectedResults)

	if len(results) == 0 && query.ExpectedResults == 0 {
		return // No results expected and none found - test passes
	}

	// Check each result
	for _, result := range results {
		// Check score range
		AssertScoreRange(t, result.Score, query.MinScore, query.MaxScore)

		// Check that result contains expected keywords
		if len(query.ShouldContain) > 0 {
			AssertContainsKeywords(t, result.Command, query.ShouldContain)
		}

		// Check that result doesn't contain unwanted keywords
		if len(query.ShouldNotContain) > 0 {
			AssertDoesNotContainKeywords(t, result.Command, query.ShouldNotContain)
		}
	}
}

// BenchmarkHelper provides benchmarking utilities
type BenchmarkHelper struct {
	db *database.Database
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(db *database.Database) *BenchmarkHelper {
	return &BenchmarkHelper{db: db}
}

// BenchmarkSearch benchmarks search operations
func (bh *BenchmarkHelper) BenchmarkSearch(b *testing.B, query string, limit int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bh.db.Search(query, limit)
	}
}

// BenchmarkSearchWithOptions benchmarks search with options
func (bh *BenchmarkHelper) BenchmarkSearchWithOptions(b *testing.B, query string, options database.SearchOptions) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bh.db.SearchWithOptions(query, options)
	}
}

// MemoryHelper provides memory usage testing utilities
type MemoryHelper struct{}

// NewMemoryHelper creates a new memory helper
func NewMemoryHelper() *MemoryHelper {
	return &MemoryHelper{}
}

// MeasureMemoryUsage measures memory usage of a function
func (mh *MemoryHelper) MeasureMemoryUsage(testFunc func()) (beforeMem, afterMem uint64) {
	// This is a simplified version - in a real implementation,
	// we would use runtime.MemStats to measure actual memory usage
	testFunc()
	return 0, 0 // Placeholder values
}

// SimpleTestDataGenerator generates simple test data (renamed to avoid conflict)
type SimpleTestDataGenerator struct{}

// NewSimpleTestDataGenerator creates a new simple test data generator
func NewSimpleTestDataGenerator() *SimpleTestDataGenerator {
	return &SimpleTestDataGenerator{}
}

// GenerateCommands generates a specified number of test commands
func (tdg *SimpleTestDataGenerator) GenerateCommands(count int) []database.Command {
	commands := make([]database.Command, count)

	baseCommands := []string{
		"git", "find", "grep", "tar", "zip", "curl", "wget", "ssh", "scp", "rsync",
		"ls", "cp", "mv", "rm", "mkdir", "rmdir", "cat", "less", "more", "head", "tail",
		"ps", "top", "htop", "kill", "killall", "jobs", "bg", "fg", "nohup",
		"chmod", "chown", "chgrp", "sudo", "su", "whoami", "id", "groups",
	}

	baseDescriptions := []string{
		"version control", "file operations", "text processing", "archive management",
		"network operations", "system monitoring", "process management", "permissions",
	}

	baseKeywords := []string{
		"file", "directory", "search", "process", "network", "system", "text", "archive",
		"permission", "user", "group", "monitor", "manage", "create", "delete", "copy",
	}

	for i := 0; i < count; i++ {
		cmdIndex := i % len(baseCommands)
		descIndex := i % len(baseDescriptions)

		commands[i] = database.Command{
			Command:     baseCommands[cmdIndex] + " test" + string(rune('A'+i%26)),
			Description: baseDescriptions[descIndex] + " for testing",
			Keywords:    []string{baseKeywords[i%len(baseKeywords)], "test"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    i%5 == 0, // Every 5th command is a pipeline command
		}
	}

	return commands
}

// GenerateQueries generates test queries for various scenarios
func (tdg *SimpleTestDataGenerator) GenerateQueries(count int) []TestQuery {
	queries := make([]TestQuery, count)

	baseQueries := []string{
		"git commit", "find files", "search text", "compress archive", "download file",
		"list directory", "copy file", "move file", "delete file", "create directory",
		"process list", "system monitor", "change permission", "network connection",
	}

	for i := 0; i < count; i++ {
		queryIndex := i % len(baseQueries)

		queries[i] = TestQuery{
			Query:            baseQueries[queryIndex],
			ExpectedResults:  1 + i%3, // 1-3 expected results
			MinScore:         float64(i%10 + 1),
			MaxScore:         float64(i%10 + 20),
			ShouldContain:    []string{baseQueries[queryIndex][:3]}, // First 3 chars as keyword
			ShouldNotContain: []string{},
		}
	}

	return queries
}

// PathHelper provides path-related utilities for testing
type PathHelper struct{}

// NewPathHelper creates a new path helper
func NewPathHelper() *PathHelper {
	return &PathHelper{}
}

// GetTestDataDir returns the path to the test data directory
func (ph *PathHelper) GetTestDataDir() string {
	return GetTestDataPath()
}

// EnsureTestDataDir ensures the test data directory exists
func (ph *PathHelper) EnsureTestDataDir() error {
	testDataDir := ph.GetTestDataDir()
	return os.MkdirAll(testDataDir, 0755)
}

// CreateTestDataFile creates a test data file with the given content
func (ph *PathHelper) CreateTestDataFile(filename, content string) (string, error) {
	testDataDir := ph.GetTestDataDir()
	if err := ph.EnsureTestDataDir(); err != nil {
		return "", err
	}

	filePath := filepath.Join(testDataDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return filePath, err
}

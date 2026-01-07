package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Vedant9500/WTF/internal/database"
)

// TestDatabase interface for creating test database instances
type TestDatabase interface {
	CreateTestDB(commands []database.Command) *database.Database
	CreateMinimalDB() *database.Database
	CreateLargeDB() *database.Database
	CreateEmptyDB() *database.Database
}

// TestFixtures manages test data and fixtures
type TestFixtures interface {
	GetSampleCommands() []database.Command
	GetTestQueries() []TestQuery
	CreateTempDir() (string, func())
	CreateTempFile(content string) (string, func())
}

// TestQuery represents a test query with expected results
type TestQuery struct {
	Query            string
	ExpectedResults  int
	MinScore         float64
	MaxScore         float64
	ShouldContain    []string
	ShouldNotContain []string
}

// DefaultTestDatabase implements TestDatabase interface
type DefaultTestDatabase struct{}

// DefaultTestFixtures implements TestFixtures interface
type DefaultTestFixtures struct{}

// NewTestDatabase creates a new test database helper
func NewTestDatabase() TestDatabase {
	return &DefaultTestDatabase{}
}

// NewTestFixtures creates a new test fixtures helper
func NewTestFixtures() TestFixtures {
	return &DefaultTestFixtures{}
}

// CreateTestDB creates a database with the provided commands
func (td *DefaultTestDatabase) CreateTestDB(commands []database.Command) *database.Database {
	// Populate cached lowercased fields for performance
	for i := range commands {
		commands[i].CommandLower = strings.ToLower(commands[i].Command)
		commands[i].DescriptionLower = strings.ToLower(commands[i].Description)
		commands[i].KeywordsLower = make([]string, len(commands[i].Keywords))
		for j, keyword := range commands[i].Keywords {
			commands[i].KeywordsLower[j] = strings.ToLower(keyword)
		}
	}

	return &database.Database{
		Commands: commands,
	}
}

// CreateMinimalDB creates a minimal database with basic commands for testing
func (td *DefaultTestDatabase) CreateMinimalDB() *database.Database {
	commands := []database.Command{
		{
			Command:     "git commit -m 'message'",
			Description: "commit changes with message",
			Keywords:    []string{"git", "commit", "message"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "find . -name '*.txt'",
			Description: "find text files",
			Keywords:    []string{"find", "files", "text"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "tar -czf archive.tar.gz .",
			Description: "create compressed archive",
			Keywords:    []string{"tar", "compress", "archive"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
	}
	return td.CreateTestDB(commands)
}

// CreateLargeDB creates a larger database for performance testing
func (td *DefaultTestDatabase) CreateLargeDB() *database.Database {
	var commands []database.Command
	commands = append(commands, td.createGitCommands()...)
	commands = append(commands, td.createFileCommands()...)
	commands = append(commands, td.createArchiveCommands()...)
	commands = append(commands, td.createPipelineCommands()...)
	commands = append(commands, td.createNetworkCommands()...)
	commands = append(commands, td.createSystemCommands()...)

	return td.CreateTestDB(commands)
}

func (td *DefaultTestDatabase) createGitCommands() []database.Command {
	return []database.Command{
		{
			Command:     "git init",
			Description: "initialize a new git repository",
			Keywords:    []string{"git", "init", "repository"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "git clone <url>",
			Description: "clone a remote repository",
			Keywords:    []string{"git", "clone", "remote"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "git add .",
			Description: "add all files to staging area",
			Keywords:    []string{"git", "add", "staging"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "git commit -m 'message'",
			Description: "commit changes with message",
			Keywords:    []string{"git", "commit", "message"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "git push origin main",
			Description: "push changes to remote repository",
			Keywords:    []string{"git", "push", "remote"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
	}
}

func (td *DefaultTestDatabase) createFileCommands() []database.Command {
	return []database.Command{
		{
			Command:     "find . -name '*.txt'",
			Description: "find text files",
			Keywords:    []string{"find", "files", "text"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "grep -r 'pattern' .",
			Description: "search for pattern in files recursively",
			Keywords:    []string{"grep", "search", "pattern"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "ls -la",
			Description: "list files with detailed information",
			Keywords:    []string{"ls", "list", "files"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "cp source destination",
			Description: "copy files or directories",
			Keywords:    []string{"cp", "copy", "files"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "mv source destination",
			Description: "move or rename files",
			Keywords:    []string{"mv", "move", "rename"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
	}
}

func (td *DefaultTestDatabase) createArchiveCommands() []database.Command {
	return []database.Command{
		{
			Command:     "tar -czf archive.tar.gz .",
			Description: "create compressed archive",
			Keywords:    []string{"tar", "compress", "archive"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "tar -xzf archive.tar.gz",
			Description: "extract compressed archive",
			Keywords:    []string{"tar", "extract", "archive"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "zip -r archive.zip .",
			Description: "create zip archive",
			Keywords:    []string{"zip", "archive", "compress"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "unzip archive.zip",
			Description: "extract zip archive",
			Keywords:    []string{"unzip", "extract", "archive"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
	}
}

func (td *DefaultTestDatabase) createPipelineCommands() []database.Command {
	return []database.Command{
		{
			Command:     "ps aux | grep process",
			Description: "find running processes",
			Keywords:    []string{"ps", "grep", "process"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
		},
		{
			Command:     "cat file.txt | sort | uniq",
			Description: "sort and remove duplicates from file",
			Keywords:    []string{"cat", "sort", "uniq"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
		},
	}
}

func (td *DefaultTestDatabase) createNetworkCommands() []database.Command {
	return []database.Command{
		{
			Command:     "curl -O https://example.com/file",
			Description: "download file from URL",
			Keywords:    []string{"curl", "download", "http"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "wget https://example.com/file",
			Description: "download file using wget",
			Keywords:    []string{"wget", "download", "http"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "ssh user@host",
			Description: "connect to remote host via SSH",
			Keywords:    []string{"ssh", "remote", "connect"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
	}
}

func (td *DefaultTestDatabase) createSystemCommands() []database.Command {
	return []database.Command{
		{
			Command:     "top",
			Description: "display running processes",
			Keywords:    []string{"top", "processes", "system"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "htop",
			Description: "interactive process viewer",
			Keywords:    []string{"htop", "processes", "interactive"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
	}
}

// CreateEmptyDB creates an empty database for testing edge cases
func (td *DefaultTestDatabase) CreateEmptyDB() *database.Database {
	return &database.Database{
		Commands: []database.Command{},
	}
}

// GetSampleCommands returns a standard set of commands for testing
func (tf *DefaultTestFixtures) GetSampleCommands() []database.Command {
	return []database.Command{
		{
			Command:     "git commit -m 'message'",
			Description: "commit changes with message",
			Keywords:    []string{"git", "commit", "message"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
		},
		{
			Command:     "find . -name '*.txt'",
			Description: "find text files",
			Keywords:    []string{"find", "files", "text"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "tar -czf archive.tar.gz .",
			Description: "create compressed archive",
			Keywords:    []string{"tar", "compress", "archive"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "grep -r 'pattern' .",
			Description: "search for pattern in files recursively",
			Keywords:    []string{"grep", "search", "pattern"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
		},
		{
			Command:     "ps aux | grep process",
			Description: "find running processes",
			Keywords:    []string{"ps", "grep", "process"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
		},
	}
}

// GetTestQueries returns a set of test queries with expected results
func (tf *DefaultTestFixtures) GetTestQueries() []TestQuery {
	return []TestQuery{
		{
			Query:            "git commit",
			ExpectedResults:  1,
			MinScore:         10.0,
			MaxScore:         50.0,
			ShouldContain:    []string{"git", "commit"},
			ShouldNotContain: []string{"find", "tar"},
		},
		{
			Query:            "find files",
			ExpectedResults:  1,
			MinScore:         5.0,
			MaxScore:         30.0,
			ShouldContain:    []string{"find", "files"},
			ShouldNotContain: []string{"git", "commit"},
		},
		{
			Query:            "compress archive",
			ExpectedResults:  1,
			MinScore:         5.0,
			MaxScore:         30.0,
			ShouldContain:    []string{"tar", "compress"},
			ShouldNotContain: []string{"git", "find"},
		},
		{
			Query:            "search pattern",
			ExpectedResults:  1,
			MinScore:         5.0,
			MaxScore:         30.0,
			ShouldContain:    []string{"grep", "search"},
			ShouldNotContain: []string{"git", "tar"},
		},
		{
			Query:            "nonexistent command",
			ExpectedResults:  0,
			MinScore:         0.0,
			MaxScore:         0.0,
			ShouldContain:    []string{},
			ShouldNotContain: []string{},
		},
	}
}

// CreateTempDir creates a temporary directory for testing
func (tf *DefaultTestFixtures) CreateTempDir() (dir string, cleanupFn func()) {
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// CreateTempFile creates a temporary file with the given content
func (tf *DefaultTestFixtures) CreateTempFile(content string) (path string, cleanupFn func()) {
	tempFile, err := os.CreateTemp("", "wtf_test_*.txt")
	if err != nil {
		panic("Failed to create temp file: " + err.Error())
	}

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		panic("Failed to write to temp file: " + err.Error())
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		panic("Failed to close temp file: " + err.Error())
	}

	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return tempFile.Name(), cleanup
}

// Helper functions for common test operations

// AssertCommandExists checks if a command exists in the results
func AssertCommandExists(t *testing.T, results []database.SearchResult, expectedCommand string) {
	t.Helper()
	for _, result := range results {
		if result.Command.Command == expectedCommand {
			return
		}
	}
	t.Errorf("Expected command '%s' not found in results", expectedCommand)
}

// AssertScoreRange checks if a score is within the expected range
func AssertScoreRange(t *testing.T, score, minScore, maxScore float64) {
	t.Helper()
	if score < minScore || score > maxScore {
		t.Errorf("Score %f is not within expected range [%f, %f]", score, minScore, maxScore)
	}
}

// AssertResultCount checks if the number of results matches expected count
func AssertResultCount(t *testing.T, results []database.SearchResult, expectedCount int) {
	t.Helper()
	if len(results) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(results))
	}
}

// AssertContainsKeywords checks if a command contains expected keywords
func AssertContainsKeywords(t *testing.T, cmd *database.Command, keywords []string) {
	t.Helper()
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))
	for _, keyword := range keywords {
		if !strings.Contains(cmdText, strings.ToLower(keyword)) {
			t.Errorf("Command '%s' does not contain expected keyword '%s'", cmd.Command, keyword)
		}
	}
}

// AssertDoesNotContainKeywords checks if a command does not contain unwanted keywords
func AssertDoesNotContainKeywords(t *testing.T, cmd *database.Command, keywords []string) {
	t.Helper()
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))
	for _, keyword := range keywords {
		if strings.Contains(cmdText, strings.ToLower(keyword)) {
			t.Errorf("Command '%s' contains unwanted keyword '%s'", cmd.Command, keyword)
		}
	}
}

// CreateTestDatabaseFromYAML creates a test database from YAML content
func CreateTestDatabaseFromYAML(yamlContent string) (*database.Database, error) {
	tempFile, cleanup := NewTestFixtures().CreateTempFile(yamlContent)
	defer cleanup()
	_ = tempFile // Use the tempFile variable to avoid unused variable error

	// This would typically use the loader package, but for now we'll create manually
	// In a real implementation, this would parse the YAML and create the database
	return NewTestDatabase().CreateMinimalDB(), nil
}

// GetTestDataPath returns the path to test data files
func GetTestDataPath() string {
	// Get the current working directory and construct path to test data
	wd, err := os.Getwd()
	if err != nil {
		return "testdata"
	}
	return filepath.Join(wd, "testdata")
}

// SetupTestEnvironment sets up a complete test environment
func SetupTestEnvironment(t *testing.T) (resultDB *database.Database, cleanupFn func()) {
	t.Helper()

	// Create test database
	testDB := NewTestDatabase()
	db := testDB.CreateLargeDB()

	// Create temporary directory for test files
	tempDir, cleanup := NewTestFixtures().CreateTempDir()

	// Set up any environment variables if needed
	originalDir, _ := os.Getwd()
	_ = os.Chdir(tempDir)

	cleanupFunc := func() {
		_ = os.Chdir(originalDir)
		cleanup()
	}

	return db, cleanupFunc
}

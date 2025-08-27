//go:build integration
// +build integration

package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Vedant9500/WTF/internal/testutil"
)

// TestEndToEndSearchWorkflow tests the complete search workflow from CLI to results
func TestEndToEndSearchWorkflow(t *testing.T) {
	// Setup test database
	tempDir, cleanup := testutil.CreateTempDir()
	defer cleanup()

	testDB := testutil.CreateTestDatabase(testutil.GetSampleCommands())
	dbPath := tempDir + "/test-commands.yml"

	err := testutil.SaveDatabase(testDB, dbPath)
	if err != nil {
		t.Fatalf("Failed to save test database: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name:           "Basic search",
			args:           []string{"search", "copy files", "--database", dbPath},
			expectedOutput: []string{"copy", "files"},
			shouldFail:     false,
		},
		{
			name:           "JSON output format",
			args:           []string{"search", "copy files", "--format", "json", "--database", dbPath},
			expectedOutput: []string{`"command"`, `"description"`},
			shouldFail:     false,
		},
		{
			name:           "Verbose output",
			args:           []string{"search", "copy files", "--verbose", "--database", dbPath},
			expectedOutput: []string{"Loaded", "commands", "Searching for"},
			shouldFail:     false,
		},
		{
			name:           "Limited results",
			args:           []string{"search", "copy files", "--limit", "2", "--database", dbPath},
			expectedOutput: []string{"copy"},
			shouldFail:     false,
		},
		{
			name:           "Platform filtering",
			args:           []string{"search", "copy files", "--platform", "windows", "--database", dbPath},
			expectedOutput: []string{"copy"},
			shouldFail:     false,
		},
		{
			name:           "Invalid database path",
			args:           []string{"search", "copy files", "--database", "/nonexistent/path.yml"},
			expectedOutput: []string{"database", "not found"},
			shouldFail:     true,
		},
		{
			name:           "Empty query",
			args:           []string{"search", "", "--database", dbPath},
			expectedOutput: []string{"empty", "query"},
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			buf := new(bytes.Buffer)

			// Create root command for testing
			rootCmd := NewRootCommand()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			// Check if error expectation matches
			if tt.shouldFail && err == nil {
				t.Errorf("Expected command to fail, but it succeeded")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected command to succeed, but it failed: %v", err)
			}

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(strings.ToLower(output), strings.ToLower(expected)) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

// TestSearchPerformanceRegression tests that search performance doesn't degrade
func TestSearchPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a larger test database
	largeDB := testutil.CreateLargeDatabase(1000) // 1000 commands
	tempDir, cleanup := testutil.CreateTempDir()
	defer cleanup()

	dbPath := tempDir + "/large-commands.yml"
	err := testutil.SaveDatabase(largeDB, dbPath)
	if err != nil {
		t.Fatalf("Failed to save large test database: %v", err)
	}

	// Test queries that should complete within reasonable time
	testQueries := []string{
		"copy files",
		"git commit",
		"docker run",
		"find text in files",
		"compress directory",
	}

	for _, query := range testQueries {
		t.Run("Query: "+query, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCommand()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{"search", query, "--database", dbPath})

			// Measure execution time
			start := time.Now()
			err := rootCmd.Execute()
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Search failed: %v", err)
			}

			// Performance threshold: searches should complete within 500ms
			if duration > 500*time.Millisecond {
				t.Errorf("Search took too long: %v (threshold: 500ms)", duration)
			}

			t.Logf("Query '%s' completed in %v", query, duration)
		})
	}
}

// TestConcurrentSearches tests that multiple concurrent searches work correctly
func TestConcurrentSearches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	// Setup test database
	tempDir, cleanup := testutil.CreateTempDir()
	defer cleanup()

	testDB := testutil.CreateTestDatabase(testutil.GetSampleCommands())
	dbPath := tempDir + "/concurrent-test.yml"

	err := testutil.SaveDatabase(testDB, dbPath)
	if err != nil {
		t.Fatalf("Failed to save test database: %v", err)
	}

	// Run multiple searches concurrently
	const numWorkers = 10
	const searchesPerWorker = 5

	errors := make(chan error, numWorkers*searchesPerWorker)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < searchesPerWorker; j++ {
				buf := new(bytes.Buffer)
				rootCmd := NewRootCommand()
				rootCmd.SetOut(buf)
				rootCmd.SetErr(buf)
				rootCmd.SetArgs([]string{"search", "copy files", "--database", dbPath})

				err := rootCmd.Execute()
				if err != nil {
					errors <- err
				} else {
					errors <- nil
				}
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numWorkers*searchesPerWorker; i++ {
		err := <-errors
		if err == nil {
			successCount++
		} else {
			t.Errorf("Concurrent search failed: %v", err)
		}
	}

	expectedSuccess := numWorkers * searchesPerWorker
	if successCount != expectedSuccess {
		t.Errorf("Expected %d successful searches, got %d", expectedSuccess, successCount)
	}

	t.Logf("Successfully completed %d concurrent searches", successCount)
}

// TestConfigurationValidation tests that configuration validation works end-to-end
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() string
		args        []string
		shouldFail  bool
		expectedErr string
	}{
		{
			name: "Valid configuration",
			setupConfig: func() string {
				tempDir, _ := testutil.CreateTempDir()
				testDB := testutil.CreateTestDatabase(testutil.GetSampleCommands())
				dbPath := tempDir + "/valid-config.yml"
				testutil.SaveDatabase(testDB, dbPath)
				return dbPath
			},
			args:       []string{"search", "copy files"},
			shouldFail: false,
		},
		{
			name: "Invalid database path",
			setupConfig: func() string {
				return "/nonexistent/invalid.yml"
			},
			args:        []string{"search", "copy files", "--database"},
			shouldFail:  true,
			expectedErr: "database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setupConfig()

			// Add database path to args if it's valid
			args := tt.args
			if !tt.shouldFail {
				args = append(args, "--database", dbPath)
			} else {
				args = append(args, dbPath)
			}

			buf := new(bytes.Buffer)
			rootCmd := NewRootCommand()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(args)

			err := rootCmd.Execute()

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded")
				}
				if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected command to succeed, but it failed: %v", err)
				}
			}
		})
	}
}

// TestNLPIntegration tests that NLP processing works correctly in the full pipeline
func TestNLPIntegration(t *testing.T) {
	// Setup test database with NLP-testable commands
	tempDir, cleanup := testutil.CreateTempDir()
	defer cleanup()

	// Create test database with commands that should match NLP queries
	nlpTestCommands := []testutil.Command{
		{
			Command:     "ipconfig",
			Description: "Display and manage the network configuration of Windows.",
			Keywords:    []string{"ipconfig", "network", "windows", "ip", "configuration"},
			Platform:    []string{"windows"},
		},
		{
			Command:     "git commit",
			Description: "Create a new commit with staged changes.",
			Keywords:    []string{"git", "commit", "create", "changes"},
			Platform:    []string{"cross-platform"},
		},
		{
			Command:     "docker run",
			Description: "Run a command in a new container.",
			Keywords:    []string{"docker", "run", "container", "execute"},
			Platform:    []string{"cross-platform"},
		},
	}

	testDB := testutil.CreateTestDatabase(nlpTestCommands)
	dbPath := tempDir + "/nlp-test.yml"

	err := testutil.SaveDatabase(testDB, dbPath)
	if err != nil {
		t.Fatalf("Failed to save NLP test database: %v", err)
	}

	nlpTests := []struct {
		name          string
		query         string
		expectedFirst string // First result should contain this
		description   string
	}{
		{
			name:          "Natural language IP query",
			query:         "what is the command to manage ip in windows",
			expectedFirst: "ipconfig",
			description:   "Should find ipconfig for IP management queries",
		},
		{
			name:          "Natural language git query",
			query:         "how to create a git commit",
			expectedFirst: "git commit",
			description:   "Should find git commit for commit creation queries",
		},
		{
			name:          "Natural language docker query",
			query:         "run container with docker",
			expectedFirst: "docker run",
			description:   "Should find docker run for container execution queries",
		},
	}

	for _, tt := range nlpTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCommand()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{"search", tt.query, "--database", dbPath, "--format", "json"})

			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("NLP search failed: %v", err)
			}

			output := buf.String()

			// Check that the expected command appears in the results
			if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.expectedFirst)) {
				t.Errorf("%s: Expected output to contain '%s', got: %s", tt.description, tt.expectedFirst, output)
			}

			t.Logf("NLP query '%s' successfully found '%s'", tt.query, tt.expectedFirst)
		})
	}
}

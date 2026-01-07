// Package testutil provides comprehensive testing utilities and fixtures.
//
// This package contains shared testing infrastructure including:
//   - Pre-configured test databases with realistic command data
//   - Test data generators for various scenarios
//   - Helper functions for common test operations
//   - Fixtures for consistent test data across packages
//
// The test databases provided include commands from various categories
// (git, filesystem, compression) to enable comprehensive search testing.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Command represents a database command for testing
// This is a local copy to avoid import cycles with the database package
type Command struct {
	Command     string   `yaml:"command" json:"command"`
	Description string   `yaml:"description" json:"description"`
	Keywords    []string `yaml:"keywords" json:"keywords"`
	Niche       string   `yaml:"niche,omitempty" json:"niche,omitempty"`
	Platform    []string `yaml:"platform,omitempty" json:"platform,omitempty"`
	Pipeline    bool     `yaml:"pipeline" json:"pipeline"`
}

// Database represents a test database structure
// This is a local copy to avoid import cycles with the database package
type Database struct {
	Commands []Command `yaml:"commands" json:"commands"`
}

// GetSampleCommands returns a set of sample commands for testing
func GetSampleCommands() []Command {
	return []Command{
		{
			Command:     "copy",
			Description: "Copy files",
			Keywords:    []string{"copy", "files", "duplicate"},
			Platform:    []string{"windows"},
		},
		{
			Command:     "cp",
			Description: "Copy files and directories",
			Keywords:    []string{"copy", "files", "duplicate"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "ipconfig",
			Description: "Display and manage the network configuration of Windows.",
			Keywords:    []string{"ipconfig", "network", "windows", "ip", "configuration"},
			Platform:    []string{"windows-cmd", "powershell"},
		},
	}
}

// CreateTestDatabase creates a test database with the provided commands
func CreateTestDatabase(commands []Command) *Database {
	return &Database{
		Commands: commands,
	}
}

// CreateLargeDatabase creates a test database with the specified number of commands
func CreateLargeDatabase(count int) *Database {
	sampleCommands := GetSampleCommands()
	commands := make([]Command, count)

	for i := 0; i < count; i++ {
		// Cycle through sample commands and modify them slightly
		base := sampleCommands[i%len(sampleCommands)]
		commands[i] = Command{
			Command:     base.Command + " " + fmt.Sprintf("variant-%d", i),
			Description: base.Description + fmt.Sprintf(" (variant %d)", i),
			Keywords:    base.Keywords,
			Platform:    base.Platform,
		}
	}

	return &Database{
		Commands: commands,
	}
}

// SaveDatabase saves a database to a YAML file
func SaveDatabase(db *Database, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(db.Commands)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir() (dir string, cleanupFn func()) {
	tempDir, err := os.MkdirTemp("", "wtf-test-*")
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// CreateDefaultTestDatabase creates a comprehensive test database with realistic sample commands.
//
// This function returns a database populated with commands from various categories
// including git, filesystem operations, compression, and pipeline commands.
// The commands are designed to test different aspects of the search functionality:
//   - Different command structures and complexity
//   - Various keyword combinations
//   - Platform-specific and cross-platform commands
//   - Pipeline and non-pipeline commands
//   - Different niche categories
//
// This is the primary test database used for most search functionality tests.
func CreateDefaultTestDatabase() *Database {
	return &Database{
		Commands: []Command{
			{
				Command:     "git commit -m 'message'",
				Description: "commit changes with message",
				Keywords:    []string{"git", "commit", "message", "version-control"},
				Niche:       "git",
				Platform:    []string{"linux", "macos", "windows"},
				Pipeline:    false,
			},
			{
				Command:     "find . -name '*.txt'",
				Description: "find text files in current directory",
				Keywords:    []string{"find", "files", "text", "search"},
				Niche:       "filesystem",
				Platform:    []string{"linux", "macos"},
				Pipeline:    false,
			},
			{
				Command:     "tar -czf archive.tar.gz .",
				Description: "create compressed tar archive",
				Keywords:    []string{"tar", "compress", "archive", "gzip"},
				Niche:       "compression",
				Platform:    []string{"linux", "macos"},
				Pipeline:    true,
			},
			{
				Command:     "grep -r 'pattern' .",
				Description: "search for pattern in files recursively",
				Keywords:    []string{"grep", "search", "pattern", "text"},
				Niche:       "search",
				Platform:    []string{"linux", "macos"},
				Pipeline:    true,
			},
			{
				Command:     "docker run image",
				Description: "run a docker container",
				Keywords:    []string{"docker", "run", "container", "image"},
				Niche:       "docker",
				Platform:    []string{"linux", "macos", "windows"},
				Pipeline:    false,
			},
		},
	}
}

// CreateMinimalTestDatabase creates a minimal test database with a single command.
//
// This function is useful for tests that need a simple database without the
// complexity of multiple commands. It contains only one basic command for
// testing core functionality without interference from other commands.
func CreateMinimalTestDatabase() *Database {
	return &Database{
		Commands: []Command{
			{
				Command:     "test command",
				Description: "test description",
				Keywords:    []string{"test"},
				Pipeline:    false,
			},
		},
	}
}

// CreateEmptyTestDatabase creates an empty test database with no commands.
//
// This function is useful for testing edge cases, error conditions, and
// scenarios where no search results should be found. It helps verify that
// the search functionality handles empty databases gracefully.
func CreateEmptyTestDatabase() *Database {
	return &Database{
		Commands: []Command{},
	}
}

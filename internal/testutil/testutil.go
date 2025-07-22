// Package testutil provides utilities for testing.
package testutil

import (
	"github.com/Vedant9500/WTF/internal/database"
)

// CreateTestDatabase creates a test database with sample commands
func CreateTestDatabase() *database.Database {
	return &database.Database{
		Commands: []database.Command{
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
				Pipeline:    false,
			},
			{
				Command:     "ls -la | grep '.txt'",
				Description: "list txt files with details",
				Keywords:    []string{"ls", "grep", "files", "list"},
				Niche:       "filesystem",
				Platform:    []string{"linux", "macos"},
				Pipeline:    true,
			},
			{
				Command:     "mkdir -p directory/path",
				Description: "create directory with parent directories",
				Keywords:    []string{"mkdir", "create", "directory", "folder"},
				Niche:       "filesystem",
				Platform:    []string{"linux", "macos", "windows"},
				Pipeline:    false,
			},
		},
	}
}

// CreateMinimalTestDatabase creates a minimal test database
func CreateMinimalTestDatabase() *database.Database {
	return &database.Database{
		Commands: []database.Command{
			{
				Command:     "test command",
				Description: "test description",
				Keywords:    []string{"test"},
				Pipeline:    false,
			},
		},
	}
}

// CreateEmptyTestDatabase creates an empty test database
func CreateEmptyTestDatabase() *database.Database {
	return &database.Database{
		Commands: []database.Command{},
	}
}

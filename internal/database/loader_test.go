package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDatabase(t *testing.T) {
	// Create a temporary test YAML file
	testYAML := `- command: "test command"
  description: "test description"
  keywords: ["test", "keyword"]
  platform: [linux, macos]
  pipeline: false
`

	// Create temporary file
	tmpDir, err := os.MkdirTemp("", "cmd-finder-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.yml")
	err = os.WriteFile(testFile, []byte(testYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading
	db, err := LoadDatabase(testFile)
	if err != nil {
		t.Fatalf("LoadDatabase failed: %v", err)
	}

	if db.Size() != 1 {
		t.Errorf("Expected 1 command, got %d", db.Size())
	}

	if len(db.Commands) == 0 {
		t.Fatal("No commands loaded")
	}

	cmd := db.Commands[0]
	if cmd.Command != "test command" {
		t.Errorf("Expected 'test command', got '%s'", cmd.Command)
	}

	if cmd.Description != "test description" {
		t.Errorf("Expected 'test description', got '%s'", cmd.Description)
	}
}

func TestLoadDatabaseFileNotFound(t *testing.T) {
	_, err := LoadDatabase("nonexistent.yml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestDatabaseSize(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{Command: "cmd1"},
			{Command: "cmd2"},
		},
	}

	if db.Size() != 2 {
		t.Errorf("Expected size 2, got %d", db.Size())
	}
}

func TestLoadDatabaseWithPersonal(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create main database file
	mainDBPath := filepath.Join(tempDir, "main.yml")
	mainDBContent := `- command: "ls -la"
  description: "list files with details"
  keywords: ["ls", "list", "files"]
- command: "cd /path"
  description: "change directory"
  keywords: ["cd", "directory"]`

	err := os.WriteFile(mainDBPath, []byte(mainDBContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main database file: %v", err)
	}

	// Create personal database file
	personalDBPath := filepath.Join(tempDir, "personal.yml")
	personalDBContent := `- command: "my custom command"
  description: "my custom description"
  keywords: ["custom", "personal"]
  niche: "custom"
  pipeline: false`

	err = os.WriteFile(personalDBPath, []byte(personalDBContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create personal database file: %v", err)
	}

	// Load combined database
	db, err := LoadDatabaseWithPersonal(mainDBPath, personalDBPath)
	if err != nil {
		t.Fatalf("LoadDatabaseWithPersonal failed: %v", err)
	}

	// Verify commands were loaded from both files
	expectedCount := 3 // 2 from main + 1 from personal
	if db.Size() != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, db.Size())
	}

	// Verify personal command is included
	found := false
	for _, cmd := range db.Commands {
		if cmd.Command == "my custom command" {
			found = true
			if cmd.Description != "my custom description" {
				t.Errorf("Expected description 'my custom description', got '%s'", cmd.Description)
			}
			if cmd.Niche != "custom" {
				t.Errorf("Expected niche 'custom', got '%s'", cmd.Niche)
			}
			break
		}
	}
	if !found {
		t.Error("Personal command not found in loaded database")
	}
}

func TestLoadDatabaseWithPersonal_NoPersonalFile(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create only main database file
	mainDBPath := filepath.Join(tempDir, "main.yml")
	mainDBContent := `- command: "ls -la"
  description: "list files with details"
  keywords: ["ls", "list", "files"]`

	err := os.WriteFile(mainDBPath, []byte(mainDBContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main database file: %v", err)
	}

	// Personal database path doesn't exist
	personalDBPath := filepath.Join(tempDir, "nonexistent.yml")

	// Load combined database - should succeed with just main database
	db, err := LoadDatabaseWithPersonal(mainDBPath, personalDBPath)
	if err != nil {
		t.Fatalf("LoadDatabaseWithPersonal failed when personal DB doesn't exist: %v", err)
	}

	// Should have only main database commands
	expectedCount := 1
	if db.Size() != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, db.Size())
	}
}

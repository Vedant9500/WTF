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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxResults != 5 {
		t.Errorf("Expected MaxResults 5, got %d", cfg.MaxResults)
	}

	if cfg.CacheEnabled != true {
		t.Error("Expected CacheEnabled to be true")
	}

	if cfg.DatabasePath != "assets/commands.yml" {
		t.Errorf("Expected DatabasePath 'assets/commands.yml', got '%s'", cfg.DatabasePath)
	}
}

func TestGetDatabasePath(t *testing.T) {
	// Create a temporary directory and file for testing
	tmpDir, err := os.MkdirTemp("", "cmd-finder-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.yml")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &Config{
		DatabasePath: testFile,
		MaxResults:   5,
	}

	path := cfg.GetDatabasePath()
	if path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, path)
	}
}

func TestGetDatabasePathFallback(t *testing.T) {
	cfg := &Config{
		DatabasePath: "nonexistent.yml",
		MaxResults:   5,
	}

	// Should return the configured path even if file doesn't exist
	path := cfg.GetDatabasePath()
	if path != "nonexistent.yml" {
		t.Errorf("Expected fallback to configured path, got '%s'", path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cmd-finder-config-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, "config", "cmd-finder")
	cfg := &Config{
		ConfigDir: configDir,
	}

	err = cfg.EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir failed: %v", err)
	}

	// Check if directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

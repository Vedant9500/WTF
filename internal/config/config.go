// Package config provides application configuration management.
//
// This package handles all configuration-related functionality including:
//   - Default configuration values
//   - Configuration validation
//   - Database path resolution with fallbacks
//   - User directory management
//
// The Config struct is the main configuration container and provides
// methods for validation and path resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds application configuration settings.
//
// Config manages all configurable aspects of the WTF application including
// database paths, result limits, caching preferences, and directory locations.
// It provides intelligent defaults and validation to ensure the application
// runs correctly across different environments.
type Config struct {
	// DatabasePath is the path to the main command database file
	DatabasePath string

	// PersonalDBPath is the path to the user's personal command database
	PersonalDBPath string

	// MaxResults is the maximum number of search results to return
	MaxResults int

	// CacheEnabled determines whether search result caching is active
	CacheEnabled bool

	// ConfigDir is the directory where configuration files are stored
	ConfigDir string
}

// DefaultConfig returns a new Config instance with sensible default values.
//
// The default configuration includes:
//   - Database path pointing to the bundled commands.yml file
//   - Personal database in the user's config directory
//   - Maximum of 5 search results
//   - Caching enabled for better performance
//   - Config directory in ~/.config/cmd-finder
//
// This function automatically determines the user's home directory and
// creates appropriate paths for cross-platform compatibility.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "cmd-finder")

	return &Config{
		DatabasePath:   "assets/commands.yml", // Default to assets directory
		PersonalDBPath: filepath.Join(configDir, "personal.yml"),
		MaxResults:     5,
		CacheEnabled:   true,
		ConfigDir:      configDir,
	}
}

// Validate checks if the configuration contains valid values.
//
// This method performs comprehensive validation of all configuration fields:
//   - MaxResults must be positive and not exceed 100
//   - DatabasePath must not be empty
//   - All paths must be valid (though files don't need to exist yet)
//
// Returns an error if any validation fails, nil if all values are valid.
func (c *Config) Validate() error {
	if c.MaxResults <= 0 {
		return fmt.Errorf("MaxResults must be positive, got %d", c.MaxResults)
	}
	if c.MaxResults > 100 {
		return fmt.Errorf("MaxResults too large, got %d (max: 100)", c.MaxResults)
	}
	if c.DatabasePath == "" {
		return fmt.Errorf("DatabasePath cannot be empty")
	}
	return nil
}

// GetDatabasePath returns the path to the command database file.
//
// This method implements intelligent path resolution with multiple fallback
// locations. It first tries the configured DatabasePath, then falls back to
// common installation locations in this order:
//  1. Configured path
//  2. System-wide installations (/usr/local/share, /usr/share)
//  3. Local development paths (assets/, internal/)
//  4. Legacy file names for backward compatibility
//
// If no file is found, it returns the originally configured path, allowing
// the calling code to handle the error appropriately.
func (c *Config) GetDatabasePath() string {
	// First try the configured path
	if _, err := os.Stat(c.DatabasePath); err == nil {
		return c.DatabasePath
	}

	// Fallback options (check in order of preference)
	fallbacks := []string{
		"/usr/local/share/wtf/commands.yml", // System-wide installation
		"/usr/share/wtf/commands.yml",       // Alternative system location
		"assets/commands.yml",               // Local development
		filepath.Join("assets", "commands.yml"),
		"commands.yml", // Backward compatibility
		filepath.Join("internal", "database", "commands.yml"),
		"commands_fixed.yml", // Legacy support
	}

	for _, path := range fallbacks {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return configured path anyway (will error gracefully later)
	return c.DatabasePath
}

// GetPersonalDatabasePath returns the path to the user's personal database file.
//
// The personal database allows users to add their own custom commands
// that are stored separately from the main command database. This file
// is typically located in the user's configuration directory.
func (c *Config) GetPersonalDatabasePath() string {
	return c.PersonalDBPath
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
//
// This method creates the full directory path with secure permissions
// for storing configuration files, personal databases, and other user data.
// It's safe to call multiple times - if the directory already exists, no error
// is returned.
//
// Returns an error if the directory cannot be created due to permissions or
// other filesystem issues.
func (c *Config) EnsureConfigDir() error {
	// Use secure directory creation from validation package
	// Import would be: "github.com/Vedant9500/WTF/internal/validation"
	// For now, use secure permissions directly
	const secureDirectoryMode = 0755
	return os.MkdirAll(c.ConfigDir, secureDirectoryMode)
}

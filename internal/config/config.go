package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	DatabasePath   string
	PersonalDBPath string
	MaxResults     int
	CacheEnabled   bool
	ConfigDir      string
}

// DefaultConfig returns default configuration
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

// Validate checks if the configuration is valid
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

// GetDatabasePath returns the database path, checking if file exists
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

// GetPersonalDatabasePath returns the path to the personal database file
func (c *Config) GetPersonalDatabasePath() string {
	return c.PersonalDBPath
}

// EnsureConfigDir creates the config directory if it doesn't exist
func (c *Config) EnsureConfigDir() error {
	return os.MkdirAll(c.ConfigDir, 0755)
}

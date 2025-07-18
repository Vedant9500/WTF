package config

import (
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

// GetDatabasePath returns the database path, checking if file exists
func (c *Config) GetDatabasePath() string {
	// First try the configured path
	if _, err := os.Stat(c.DatabasePath); err == nil {
		return c.DatabasePath
	}

	// Fallback options
	fallbacks := []string{
		"assets/commands.yml", // New organized location
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

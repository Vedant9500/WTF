// Package recovery provides error recovery mechanisms for the WTF application
package recovery

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/database"
	"github.com/Vedant9500/WTF/internal/errors"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// DatabaseRecovery handles database loading with fallback mechanisms
type DatabaseRecovery struct {
	retryConfig RetryConfig
}

// NewDatabaseRecovery creates a new database recovery instance
func NewDatabaseRecovery(config RetryConfig) *DatabaseRecovery {
	return &DatabaseRecovery{
		retryConfig: config,
	}
}

// LoadDatabaseWithFallback attempts to load the database with multiple fallback strategies
func (dr *DatabaseRecovery) LoadDatabaseWithFallback(primaryPath, personalPath string) (*database.Database, error) {
	// Try to load the primary database with retry
	db, err := dr.loadWithRetry(primaryPath, personalPath)
	if err == nil {
		return db, nil
	}

	// Store the primary error for reporting
	primaryErr := err

	// Try fallback strategies
	fallbackStrategies := []struct {
		name string
		fn   func() (*database.Database, error)
	}{
		{
			name: "embedded default database",
			fn:   dr.loadEmbeddedDatabase,
		},
		{
			name: "backup database",
			fn:   func() (*database.Database, error) { return dr.loadBackupDatabase(primaryPath) },
		},
		{
			name: "minimal database",
			fn:   dr.createMinimalDatabase,
		},
	}

	for _, strategy := range fallbackStrategies {
		if db, err := strategy.fn(); err == nil {
			// Create a warning error that includes the original failure and recovery info
			recoveryErr := errors.NewAppError(
				errors.ErrorTypeDatabase,
				fmt.Sprintf("primary database failed, using %s", strategy.name),
				primaryErr,
			).WithUserMessage(
				fmt.Sprintf("Warning: Could not load the main database, using %s instead.\n\nSome commands may be missing. To fix this:\n• Check the database file at '%s'\n• Run 'wtf setup' to reinitialize the database\n• Restore from backup if available", strategy.name, primaryPath),
			).WithContext("fallback_strategy", strategy.name).
				WithContext("primary_path", primaryPath).
				WithSuggestions(
					"Run 'wtf setup' to reinitialize the database",
					"Check if the database file exists and is readable",
					"Restore from a backup if available",
				)

			// Return the database with a warning (not an error)
			fmt.Printf("Warning: %s\n", recoveryErr.Error())
			return db, nil
		}
	}

	// If all fallback strategies failed, return the original error
	return nil, errors.NewAppError(
		errors.ErrorTypeDatabase,
		"all database loading strategies failed",
		primaryErr,
	).WithUserMessage(
		"Failed to load any database. The application cannot function without a command database.\n\nPlease:\n• Check the database file exists\n• Run 'wtf setup' to create a new database\n• Ensure you have proper file permissions",
	).WithSuggestions(
		"Run 'wtf setup' to create a new database",
		"Check file permissions on the database directory",
		"Verify the database file is not corrupted",
	)
}

// loadWithRetry attempts to load the database with exponential backoff retry
func (dr *DatabaseRecovery) loadWithRetry(primaryPath, personalPath string) (*database.Database, error) {
	var lastErr error

	for attempt := 1; attempt <= dr.retryConfig.MaxAttempts; attempt++ {
		db, err := database.LoadDatabaseWithPersonal(primaryPath, personalPath)
		if err == nil {
			return db, nil
		}

		lastErr = err

		// Don't retry for certain types of errors
		if !dr.shouldRetry(err) {
			break
		}

		// Don't sleep on the last attempt
		if attempt < dr.retryConfig.MaxAttempts {
			delay := dr.calculateDelay(attempt)
			time.Sleep(delay)
		}
	}

	return nil, lastErr
}

// shouldRetry determines if an error is worth retrying
func (dr *DatabaseRecovery) shouldRetry(err error) bool {
	// Don't retry for file not found or permission errors
	if os.IsNotExist(err) || os.IsPermission(err) {
		return false
	}

	// Check if it's an AppError with specific types that shouldn't be retried
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case errors.ErrorTypePermission, errors.ErrorTypeValidation:
			return false
		}
	}

	// Retry for other errors (network issues, temporary file locks, etc.)
	return true
}

// calculateDelay calculates the delay for exponential backoff
func (dr *DatabaseRecovery) calculateDelay(attempt int) time.Duration {
	delay := float64(dr.retryConfig.BaseDelay) * math.Pow(dr.retryConfig.BackoffFactor, float64(attempt-1))

	if delay > float64(dr.retryConfig.MaxDelay) {
		delay = float64(dr.retryConfig.MaxDelay)
	}

	return time.Duration(delay)
}

// loadEmbeddedDatabase loads a minimal embedded database as fallback
func (dr *DatabaseRecovery) loadEmbeddedDatabase() (*database.Database, error) {
	// Create a minimal set of essential commands
	essentialCommands := []database.Command{
		{
			Command:     "ls",
			Description: "List directory contents",
			Keywords:    []string{"list", "directory", "files"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "dir",
			Description: "List directory contents",
			Keywords:    []string{"list", "directory", "files"},
			Platform:    []string{"windows"},
		},
		{
			Command:     "cd",
			Description: "Change directory",
			Keywords:    []string{"change", "directory", "navigate"},
			Platform:    []string{"cross-platform"},
		},
		{
			Command:     "pwd",
			Description: "Print working directory",
			Keywords:    []string{"current", "directory", "path"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "mkdir",
			Description: "Create directory",
			Keywords:    []string{"create", "directory", "folder"},
			Platform:    []string{"cross-platform"},
		},
		{
			Command:     "rm",
			Description: "Remove files and directories",
			Keywords:    []string{"delete", "remove", "files"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "del",
			Description: "Delete files",
			Keywords:    []string{"delete", "remove", "files"},
			Platform:    []string{"windows"},
		},
		{
			Command:     "cp",
			Description: "Copy files and directories",
			Keywords:    []string{"copy", "files", "duplicate"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "copy",
			Description: "Copy files",
			Keywords:    []string{"copy", "files", "duplicate"},
			Platform:    []string{"windows"},
		},
		{
			Command:     "mv",
			Description: "Move/rename files and directories",
			Keywords:    []string{"move", "rename", "files"},
			Platform:    []string{"linux", "macos"},
		},
		{
			Command:     "move",
			Description: "Move files",
			Keywords:    []string{"move", "files"},
			Platform:    []string{"windows"},
		},
	}

	// Populate lowercased cache fields for performance
	for i := range essentialCommands {
		cmd := &essentialCommands[i]
		cmd.CommandLower = strings.ToLower(cmd.Command)
		cmd.DescriptionLower = strings.ToLower(cmd.Description)
		cmd.KeywordsLower = make([]string, len(cmd.Keywords))
		for j, kw := range cmd.Keywords {
			cmd.KeywordsLower[j] = strings.ToLower(kw)
		}
		cmd.TagsLower = make([]string, len(cmd.Tags))
		for j, tag := range cmd.Tags {
			cmd.TagsLower[j] = strings.ToLower(tag)
		}
	}

	return &database.Database{
		Commands: essentialCommands,
	}, nil
}

// loadBackupDatabase attempts to load from a backup file
func (dr *DatabaseRecovery) loadBackupDatabase(primaryPath string) (*database.Database, error) {
	backupPath := primaryPath + ".backup"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup database not found at %s", backupPath)
	}

	return database.LoadDatabase(backupPath)
}

// createMinimalDatabase creates a minimal database with just basic commands
func (dr *DatabaseRecovery) createMinimalDatabase() (*database.Database, error) {
	// This is a last resort - create a very minimal database
	minimalCommands := []database.Command{
		{
			Command:     "help",
			Description: "Show help information",
			Keywords:    []string{"help", "assistance"},
			Platform:    []string{"cross-platform"},
		},
		{
			Command:     "wtf setup",
			Description: "Set up WTF database",
			Keywords:    []string{"setup", "initialize", "configure"},
			Platform:    []string{"cross-platform"},
		},
	}

	// Populate lowercased cache fields
	for i := range minimalCommands {
		cmd := &minimalCommands[i]
		cmd.CommandLower = strings.ToLower(cmd.Command)
		cmd.DescriptionLower = strings.ToLower(cmd.Description)
		cmd.KeywordsLower = make([]string, len(cmd.Keywords))
		for j, kw := range cmd.Keywords {
			cmd.KeywordsLower[j] = strings.ToLower(kw)
		}
		cmd.TagsLower = make([]string, len(cmd.Tags))
		for j, tag := range cmd.Tags {
			cmd.TagsLower[j] = strings.ToLower(tag)
		}
	}

	return &database.Database{
		Commands: minimalCommands,
	}, nil
}

// SearchRecovery handles search operation failures with graceful degradation
type SearchRecovery struct{}

// NewSearchRecovery creates a new search recovery instance
func NewSearchRecovery() *SearchRecovery {
	return &SearchRecovery{}
}

// RecoverFromSearchFailure provides graceful degradation when search fails
func (sr *SearchRecovery) RecoverFromSearchFailure(query string, originalErr error, db *database.Database) ([]database.SearchResult, error) {
	// Try simpler search strategies as fallback
	fallbackStrategies := []struct {
		name string
		fn   func(string, *database.Database) ([]database.SearchResult, error)
	}{
		{
			name: "basic keyword search",
			fn:   sr.basicKeywordSearch,
		},
		{
			name: "single word search",
			fn:   sr.singleWordSearch,
		},
		{
			name: "partial match search",
			fn:   sr.partialMatchSearch,
		},
	}

	for _, strategy := range fallbackStrategies {
		if results, err := strategy.fn(query, db); err == nil && len(results) > 0 {
			// Create a warning about the degraded search
			fmt.Printf("Warning: Search had issues, using %s instead. Some results may be missing.\n\n", strategy.name)
			return results, nil
		}
	}

	// If all strategies failed, return helpful suggestions
	suggestions := []string{
		"Try using simpler keywords",
		"Check for typos in your query",
		"Use more general terms",
		"Try searching for individual words",
	}

	return nil, errors.NewSearchFailedError(query, originalErr).
		WithSuggestions(suggestions...)
}

// basicKeywordSearch performs a simple keyword-based search
func (sr *SearchRecovery) basicKeywordSearch(query string, db *database.Database) ([]database.SearchResult, error) {
	// Simple implementation - just look for exact matches in command names
	var results []database.SearchResult
	queryLower := strings.ToLower(query)

	for i := range db.Commands {
		cmd := &db.Commands[i]
		if strings.Contains(cmd.CommandLower, queryLower) {
			results = append(results, database.SearchResult{
				Command: cmd,
				Score:   1.0,
			})
		}
	}

	return results, nil
}

// singleWordSearch searches using only the first word of the query
func (sr *SearchRecovery) singleWordSearch(query string, db *database.Database) ([]database.SearchResult, error) {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil, fmt.Errorf("no words in query")
	}

	firstWord := words[0]
	var results []database.SearchResult

	for i := range db.Commands {
		cmd := &db.Commands[i]
		if strings.Contains(cmd.CommandLower, firstWord) ||
			strings.Contains(cmd.DescriptionLower, firstWord) {
			results = append(results, database.SearchResult{
				Command: cmd,
				Score:   0.8,
			})
		}
	}

	return results, nil
}

// partialMatchSearch performs partial matching on command names
func (sr *SearchRecovery) partialMatchSearch(query string, db *database.Database) ([]database.SearchResult, error) {
	var results []database.SearchResult
	queryLower := strings.ToLower(query)

	for i := range db.Commands {
		cmd := &db.Commands[i]
		// Check if any part of the query matches any part of the command
		for _, word := range strings.Fields(queryLower) {
			if len(word) >= 2 && (strings.Contains(cmd.CommandLower, word) ||
				strings.Contains(cmd.DescriptionLower, word)) {
				results = append(results, database.SearchResult{
					Command: cmd,
					Score:   0.6,
				})
				break
			}
		}
	}

	return results, nil
}

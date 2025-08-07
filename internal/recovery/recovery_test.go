package recovery

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Vedant9500/WTF/internal/database"
	appErrors "github.com/Vedant9500/WTF/internal/errors"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}

	if config.BaseDelay != 100*time.Millisecond {
		t.Errorf("Expected BaseDelay to be 100ms, got %v", config.BaseDelay)
	}

	if config.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor to be 2.0, got %f", config.BackoffFactor)
	}
}

func TestCalculateDelay(t *testing.T) {
	dr := NewDatabaseRecovery(DefaultRetryConfig())

	tests := []struct {
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{1, 100 * time.Millisecond, 100 * time.Millisecond},
		{2, 200 * time.Millisecond, 200 * time.Millisecond},
		{3, 400 * time.Millisecond, 400 * time.Millisecond},
		{10, 5 * time.Second, 5 * time.Second}, // Should be capped at MaxDelay
	}

	for _, tt := range tests {
		delay := dr.calculateDelay(tt.attempt)
		if delay < tt.expectedMin || delay > tt.expectedMax {
			t.Errorf("For attempt %d, expected delay between %v and %v, got %v",
				tt.attempt, tt.expectedMin, tt.expectedMax, delay)
		}
	}
}

func TestShouldRetry(t *testing.T) {
	dr := NewDatabaseRecovery(DefaultRetryConfig())

	tests := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{
			name:        "file not found",
			err:         os.ErrNotExist,
			shouldRetry: false,
		},
		{
			name:        "permission denied",
			err:         os.ErrPermission,
			shouldRetry: false,
		},
		{
			name:        "validation error",
			err:         appErrors.NewAppError(appErrors.ErrorTypeValidation, "test", nil),
			shouldRetry: false,
		},
		{
			name:        "permission app error",
			err:         appErrors.NewAppError(appErrors.ErrorTypePermission, "test", nil),
			shouldRetry: false,
		},
		{
			name:        "database error",
			err:         appErrors.NewAppError(appErrors.ErrorTypeDatabase, "test", nil),
			shouldRetry: true,
		},
		{
			name:        "generic error",
			err:         errors.New("generic error"),
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dr.shouldRetry(tt.err)
			if result != tt.shouldRetry {
				t.Errorf("Expected shouldRetry to be %v for %s, got %v",
					tt.shouldRetry, tt.name, result)
			}
		})
	}
}

func TestLoadEmbeddedDatabase(t *testing.T) {
	dr := NewDatabaseRecovery(DefaultRetryConfig())

	db, err := dr.loadEmbeddedDatabase()
	if err != nil {
		t.Fatalf("Expected no error loading embedded database, got: %v", err)
	}

	if len(db.Commands) == 0 {
		t.Error("Expected embedded database to have commands")
	}

	// Check that essential commands are present
	essentialCommands := []string{"ls", "dir", "cd", "mkdir"}
	found := make(map[string]bool)

	for _, cmd := range db.Commands {
		for _, essential := range essentialCommands {
			if cmd.Command == essential {
				found[essential] = true
			}
		}
	}

	for _, essential := range essentialCommands {
		if !found[essential] {
			t.Errorf("Expected to find essential command '%s' in embedded database", essential)
		}
	}

	// Verify that lowercased fields are populated
	for _, cmd := range db.Commands {
		if cmd.CommandLower == "" {
			t.Errorf("Expected CommandLower to be populated for command '%s'", cmd.Command)
		}
		if cmd.DescriptionLower == "" {
			t.Errorf("Expected DescriptionLower to be populated for command '%s'", cmd.Command)
		}
	}
}

func TestCreateMinimalDatabase(t *testing.T) {
	dr := NewDatabaseRecovery(DefaultRetryConfig())

	db, err := dr.createMinimalDatabase()
	if err != nil {
		t.Fatalf("Expected no error creating minimal database, got: %v", err)
	}

	if len(db.Commands) == 0 {
		t.Error("Expected minimal database to have commands")
	}

	// Check that help command is present
	helpFound := false
	for _, cmd := range db.Commands {
		if strings.Contains(cmd.Command, "help") {
			helpFound = true
			break
		}
	}

	if !helpFound {
		t.Error("Expected to find help command in minimal database")
	}
}

func TestSearchRecoveryBasicKeywordSearch(t *testing.T) {
	sr := NewSearchRecovery()

	// Create a test database
	testDB := &database.Database{
		Commands: []database.Command{
			{
				Command:          "ls",
				CommandLower:     "ls",
				Description:      "List files",
				DescriptionLower: "list files",
			},
			{
				Command:          "mkdir",
				CommandLower:     "mkdir",
				Description:      "Create directory",
				DescriptionLower: "create directory",
			},
		},
	}

	results, err := sr.basicKeywordSearch("ls", testDB)
	if err != nil {
		t.Fatalf("Expected no error in basic keyword search, got: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].Command.Command != "ls" {
		t.Errorf("Expected to find 'ls' command, got '%s'", results[0].Command.Command)
	}
}

func TestSearchRecoverySingleWordSearch(t *testing.T) {
	sr := NewSearchRecovery()

	// Create a test database
	testDB := &database.Database{
		Commands: []database.Command{
			{
				Command:          "ls -la",
				CommandLower:     "ls -la",
				Description:      "List files with details",
				DescriptionLower: "list files with details",
			},
			{
				Command:          "mkdir test",
				CommandLower:     "mkdir test",
				Description:      "Create test directory",
				DescriptionLower: "create test directory",
			},
		},
	}

	results, err := sr.singleWordSearch("list files", testDB)
	if err != nil {
		t.Fatalf("Expected no error in single word search, got: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result from single word search")
	}

	// Should find the ls command because "list" is in the description
	found := false
	for _, result := range results {
		if result.Command.Command == "ls -la" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'ls -la' command in single word search results")
	}
}

func TestSearchRecoveryPartialMatchSearch(t *testing.T) {
	sr := NewSearchRecovery()

	// Create a test database
	testDB := &database.Database{
		Commands: []database.Command{
			{
				Command:          "git commit",
				CommandLower:     "git commit",
				Description:      "Commit changes",
				DescriptionLower: "commit changes",
			},
			{
				Command:          "git push",
				CommandLower:     "git push",
				Description:      "Push to remote",
				DescriptionLower: "push to remote",
			},
		},
	}

	results, err := sr.partialMatchSearch("git", testDB)
	if err != nil {
		t.Fatalf("Expected no error in partial match search, got: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Both commands should be found since they both contain "git"
	commands := make(map[string]bool)
	for _, result := range results {
		commands[result.Command.Command] = true
	}

	if !commands["git commit"] || !commands["git push"] {
		t.Error("Expected to find both git commands in partial match search")
	}
}

func TestRecoverFromSearchFailure(t *testing.T) {
	sr := NewSearchRecovery()

	// Create a test database
	testDB := &database.Database{
		Commands: []database.Command{
			{
				Command:          "ls",
				CommandLower:     "ls",
				Description:      "List files",
				DescriptionLower: "list files",
			},
		},
	}

	// Test recovery with a query that should find results
	results, err := sr.RecoverFromSearchFailure("ls", errors.New("search failed"), testDB)
	if err != nil {
		t.Fatalf("Expected successful recovery, got error: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected recovery to find results")
	}

	// Test recovery with a query that won't find results
	results, err = sr.RecoverFromSearchFailure("nonexistent", errors.New("search failed"), testDB)
	if err == nil {
		t.Error("Expected error when recovery fails to find results")
	}

	// Check that the error has suggestions
	if appErr, ok := err.(*appErrors.AppError); ok {
		if len(appErr.Suggestions) == 0 {
			t.Error("Expected error to have suggestions")
		}
	}
}

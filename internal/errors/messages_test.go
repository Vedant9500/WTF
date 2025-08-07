package errors

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestUserFriendlyErrorMessages(t *testing.T) {
	tests := []struct {
		name                string
		errorFunc           func() error
		expectedInMsg       string
		expectedSuggestions int
	}{
		{
			name: "database not found error",
			errorFunc: func() error {
				return NewDatabaseNotFoundError("/nonexistent/path", os.ErrNotExist)
			},
			expectedInMsg:       "The command database file could not be found",
			expectedSuggestions: 3,
		},
		{
			name: "database parse error",
			errorFunc: func() error {
				return NewDatabaseParseError("/path/to/db.yml", errors.New("yaml: unmarshal error"))
			},
			expectedInMsg:       "contains invalid data",
			expectedSuggestions: 3,
		},
		{
			name: "query too long error",
			errorFunc: func() error {
				return NewQueryTooLongError(500, 200)
			},
			expectedInMsg:       "Your search query is too long",
			expectedSuggestions: 3,
		},
		{
			name: "query empty error",
			errorFunc: func() error {
				return NewQueryEmptyError()
			},
			expectedInMsg:       "Please provide a search query",
			expectedSuggestions: 3,
		},
		{
			name: "invalid limit error",
			errorFunc: func() error {
				return NewLimitInvalidError(150, 100)
			},
			expectedInMsg:       "The result limit must be between",
			expectedSuggestions: 3,
		},
		{
			name: "no results error with suggestions",
			errorFunc: func() error {
				return NewNoResultsError("nonexistent command", []string{"existing command", "another command"})
			},
			expectedInMsg:       "No commands found matching",
			expectedSuggestions: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()

			// Check if it's a user-friendly error
			if !IsUserFriendlyError(err) {
				t.Errorf("Expected user-friendly error, got: %T", err)
			}

			// Check user message content
			userMsg := GetUserFriendlyMessage(err)
			if !strings.Contains(userMsg, tt.expectedInMsg) {
				t.Errorf("Expected user message to contain '%s', got: %s", tt.expectedInMsg, userMsg)
			}

			// Check suggestions
			suggestions := GetErrorSuggestions(err)
			if len(suggestions) != tt.expectedSuggestions {
				t.Errorf("Expected %d suggestions, got %d: %v", tt.expectedSuggestions, len(suggestions), suggestions)
			}
		})
	}
}

func TestAppErrorChaining(t *testing.T) {
	originalErr := os.ErrNotExist
	appErr := NewDatabaseNotFoundError("/test/path", originalErr)

	// Test error unwrapping
	if appErr.Unwrap() != originalErr {
		t.Errorf("Expected unwrapped error to be %v, got %v", originalErr, appErr.Unwrap())
	}

	// Test context
	if appErr.Context["file_path"] != "/test/path" {
		t.Errorf("Expected file_path context to be '/test/path', got %v", appErr.Context["file_path"])
	}
}

func TestGetUserFriendlyMessageFallback(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "permission denied",
			err:         os.ErrPermission,
			expectedMsg: "Permission denied",
		},
		{
			name:        "generic error",
			err:         errors.New("some generic error"),
			expectedMsg: "An error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetUserFriendlyMessage(tt.err)
			if !strings.Contains(msg, tt.expectedMsg) {
				t.Errorf("Expected message to contain '%s', got: %s", tt.expectedMsg, msg)
			}
		})
	}
}

func TestErrorTemplates(t *testing.T) {
	// Test that all error templates are properly formatted
	for key, template := range errorTemplates {
		if strings.TrimSpace(template) == "" {
			t.Errorf("Error template '%s' is empty", key)
		}

		// Check that templates don't have trailing newlines that would cause formatting issues
		if strings.HasSuffix(template, "\n\n") {
			t.Errorf("Error template '%s' has excessive trailing newlines", key)
		}
	}
}

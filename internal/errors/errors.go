package errors

import (
	"fmt"
	"strings"
)

// DatabaseError represents database-related errors
type DatabaseError struct {
	Path  string
	Op    string
	Cause error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database %s failed for '%s': %v", e.Op, e.Path, e.Cause)
}

// Unwrap returns the underlying error for error chain support
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError creates a new database error (legacy function, use NewDatabaseXXXError functions for better UX)
func NewDatabaseError(op, path string, cause error) *DatabaseError {
	return &DatabaseError{
		Op:    op,
		Path:  path,
		Cause: cause,
	}
}

// NewDatabaseErrorWithContext creates a user-friendly database error based on the operation and cause
func NewDatabaseErrorWithContext(op, path string, cause error) error {
	if cause == nil {
		return NewAppError(ErrorTypeDatabase, fmt.Sprintf("database %s failed for %s", op, path), nil).
			WithContext("operation", op).
			WithContext("file_path", path)
	}
	
	// Determine the specific error type and create appropriate user-friendly error
	errStr := cause.Error()
	switch {
	case strings.Contains(errStr, "no such file or directory"):
		return NewDatabaseNotFoundError(path, cause)
	case strings.Contains(errStr, "permission denied"):
		return NewDatabasePermissionError(path, cause)
	case strings.Contains(errStr, "yaml:") || strings.Contains(errStr, "unmarshal"):
		return NewDatabaseParseError(path, cause)
	default:
		return NewAppError(ErrorTypeDatabase, fmt.Sprintf("database %s failed for %s", op, path), cause).
			WithUserMessage(fmt.Sprintf("Failed to %s database file at '%s': %v", op, path, cause)).
			WithContext("operation", op).
			WithContext("file_path", path).
			WithSuggestions(
				"Check if the file exists and is accessible",
				"Verify file permissions",
				"Try running 'wtf setup' to reinitialize",
			)
	}
}

// SearchError represents search-related errors
type SearchError struct {
	Query string
	Cause error
}

func (e *SearchError) Error() string {
	return fmt.Sprintf("search failed for query '%s': %v", e.Query, e.Cause)
}

// Unwrap returns the underlying error for error chain support
func (e *SearchError) Unwrap() error {
	return e.Cause
}

// NewSearchError creates a new search error
func NewSearchError(query string, cause error) *SearchError {
	return &SearchError{
		Query: query,
		Cause: cause,
	}
}

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeDatabase    ErrorType = "database"
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeSearch      ErrorType = "search"
	ErrorTypeConfig      ErrorType = "config"
	ErrorTypeNetwork     ErrorType = "network"
	ErrorTypeFileSystem  ErrorType = "filesystem"
	ErrorTypePermission  ErrorType = "permission"
)

// AppError represents application-specific errors with context and user-friendly messages
type AppError struct {
	Type        ErrorType              `json:"type"`
	Message     string                 `json:"message"`
	UserMessage string                 `json:"user_message"`
	Cause       error                  `json:"-"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain support
func (e *AppError) Unwrap() error {
	return e.Cause
}

// GetTechnicalDetails returns detailed technical information for debugging
func (e *AppError) GetTechnicalDetails() string {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("Type: %s\n", e.Type))
	details.WriteString(fmt.Sprintf("Message: %s\n", e.Message))
	if e.UserMessage != "" {
		details.WriteString(fmt.Sprintf("User Message: %s\n", e.UserMessage))
	}
	if e.Cause != nil {
		details.WriteString(fmt.Sprintf("Cause: %v\n", e.Cause))
	}
	if len(e.Context) > 0 {
		details.WriteString("Context:\n")
		for k, v := range e.Context {
			details.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}
	return details.String()
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, message string, cause error) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithUserMessage adds a user-friendly message to the error
func (e *AppError) WithUserMessage(userMessage string) *AppError {
	e.UserMessage = userMessage
	return e
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSuggestions adds helpful suggestions to the error
func (e *AppError) WithSuggestions(suggestions ...string) *AppError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// Error message templates for common scenarios
var errorTemplates = map[string]string{
	"database_not_found": "The command database file could not be found at '%s'.\n\nThis usually happens when:\n• The database file is missing or moved\n• You're running WTF from the wrong directory\n• The database path is incorrectly configured",
	"database_parse_error": "The command database file at '%s' contains invalid data.\n\nThis could be due to:\n• Corrupted YAML syntax\n• Invalid command structure\n• File encoding issues",
	"database_permission": "Permission denied when trying to access the database file at '%s'.\n\nTo fix this:\n• Check file permissions with 'ls -la %s'\n• Ensure you have read access to the file\n• Try running with appropriate permissions",
	"query_too_long": "Your search query is too long (%d characters). Please keep queries under %d characters.\n\nTip: Try using more specific keywords instead of full sentences.",
	"query_empty": "Please provide a search query.\n\nExample: wtf \"compress a directory\"",
	"query_invalid_chars": "Your search query contains invalid characters.\n\nPlease use only letters, numbers, spaces, and common punctuation.",
	"limit_invalid": "The result limit must be between 1 and %d, but you specified %d.\n\nExample: wtf --limit 10 \"your query\"",
	"config_invalid": "The configuration file contains invalid settings.\n\nPlease check your configuration and ensure all values are valid.",
	"search_failed": "Search operation failed for query '%s'.\n\nThis might be due to:\n• Database corruption\n• Memory issues\n• Invalid search parameters",
	"no_results": "No commands found matching '%s'.\n\nTry:\n• Using different keywords\n• Checking for typos\n• Being more specific or more general",
}

// Common error creation functions with user-friendly messages

// NewDatabaseNotFoundError creates a user-friendly database not found error
func NewDatabaseNotFoundError(path string, cause error) *AppError {
	return NewAppError(ErrorTypeDatabase, fmt.Sprintf("database file not found: %s", path), cause).
		WithUserMessage(fmt.Sprintf(errorTemplates["database_not_found"], path)).
		WithContext("file_path", path).
		WithSuggestions(
			"Run 'wtf setup' to initialize the database",
			"Check if you're in the correct directory",
			"Verify the database file exists with 'ls -la "+path+"'",
		)
}

// NewDatabaseParseError creates a user-friendly database parse error
func NewDatabaseParseError(path string, cause error) *AppError {
	return NewAppError(ErrorTypeDatabase, fmt.Sprintf("failed to parse database: %s", path), cause).
		WithUserMessage(fmt.Sprintf(errorTemplates["database_parse_error"], path)).
		WithContext("file_path", path).
		WithSuggestions(
			"Check the YAML syntax in the database file",
			"Restore from backup if available",
			"Run 'wtf setup' to recreate the database",
		)
}

// NewDatabasePermissionError creates a user-friendly database permission error
func NewDatabasePermissionError(path string, cause error) *AppError {
	return NewAppError(ErrorTypePermission, fmt.Sprintf("permission denied: %s", path), cause).
		WithUserMessage(fmt.Sprintf(errorTemplates["database_permission"], path, path)).
		WithContext("file_path", path).
		WithSuggestions(
			"Check file permissions with 'ls -la "+path+"'",
			"Ensure you have read access to the file",
			"Try running with sudo if appropriate",
		)
}

// NewQueryTooLongError creates a user-friendly query too long error
func NewQueryTooLongError(queryLength, maxLength int) *AppError {
	return NewAppError(ErrorTypeValidation, fmt.Sprintf("query too long: %d characters", queryLength), nil).
		WithUserMessage(fmt.Sprintf(errorTemplates["query_too_long"], queryLength, maxLength)).
		WithContext("query_length", queryLength).
		WithContext("max_length", maxLength).
		WithSuggestions(
			"Use more specific keywords",
			"Break complex queries into simpler ones",
			"Focus on the main action you want to perform",
		)
}

// NewQueryEmptyError creates a user-friendly empty query error
func NewQueryEmptyError() *AppError {
	return NewAppError(ErrorTypeValidation, "empty query provided", nil).
		WithUserMessage(errorTemplates["query_empty"]).
		WithSuggestions(
			"wtf \"compress a directory\"",
			"wtf \"find files by name\"",
			"wtf \"git commit changes\"",
		)
}

// NewQueryInvalidCharsError creates a user-friendly invalid characters error
func NewQueryInvalidCharsError(invalidChars string) *AppError {
	return NewAppError(ErrorTypeValidation, fmt.Sprintf("invalid characters in query: %s", invalidChars), nil).
		WithUserMessage(errorTemplates["query_invalid_chars"]).
		WithContext("invalid_characters", invalidChars).
		WithSuggestions(
			"Remove special characters like <>|&",
			"Use quotes around phrases if needed",
			"Stick to alphanumeric characters and spaces",
		)
}

// NewLimitInvalidError creates a user-friendly invalid limit error
func NewLimitInvalidError(limit, maxLimit int) *AppError {
	return NewAppError(ErrorTypeValidation, fmt.Sprintf("invalid limit: %d", limit), nil).
		WithUserMessage(fmt.Sprintf(errorTemplates["limit_invalid"], maxLimit, limit)).
		WithContext("provided_limit", limit).
		WithContext("max_limit", maxLimit).
		WithSuggestions(
			fmt.Sprintf("Use --limit %d for more results", min(maxLimit, 10)),
			"Omit --limit to use the default",
			fmt.Sprintf("Maximum allowed limit is %d", maxLimit),
		)
}

// NewSearchFailedError creates a user-friendly search failed error
func NewSearchFailedError(query string, cause error) *AppError {
	return NewAppError(ErrorTypeSearch, fmt.Sprintf("search failed for: %s", query), cause).
		WithUserMessage(fmt.Sprintf(errorTemplates["search_failed"], query)).
		WithContext("query", query).
		WithSuggestions(
			"Try a simpler query",
			"Check if the database is corrupted",
			"Restart the application",
		)
}

// NewNoResultsError creates a user-friendly no results error with suggestions
func NewNoResultsError(query string, suggestions []string) *AppError {
	err := NewAppError(ErrorTypeSearch, fmt.Sprintf("no results for: %s", query), nil).
		WithUserMessage(fmt.Sprintf(errorTemplates["no_results"], query)).
		WithContext("query", query)
	
	if len(suggestions) > 0 {
		err = err.WithSuggestions(suggestions...)
		// Add "Did you mean" suggestions to the user message
		userMsg := err.UserMessage + "\n\nDid you mean:\n"
		for _, suggestion := range suggestions {
			userMsg += fmt.Sprintf("• %s\n", suggestion)
		}
		err.UserMessage = userMsg
	}
	
	return err
}

// NewConfigError creates a user-friendly configuration error
func NewConfigError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeConfig, message, cause).
		WithUserMessage(errorTemplates["config_invalid"]).
		WithSuggestions(
			"Check your configuration file syntax",
			"Reset to default configuration",
			"Refer to the documentation for valid settings",
		)
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsUserFriendlyError checks if an error is a user-friendly AppError
func IsUserFriendlyError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetUserFriendlyMessage extracts a user-friendly message from any error
func GetUserFriendlyMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Error()
	}
	
	// For non-AppError types, provide generic user-friendly messages
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "no such file or directory"):
		return "File not found. Please check the file path and try again."
	case strings.Contains(errStr, "permission denied"):
		return "Permission denied. Please check file permissions or run with appropriate privileges."
	case strings.Contains(errStr, "connection refused"):
		return "Connection failed. Please check your network connection and try again."
	case strings.Contains(errStr, "timeout"):
		return "Operation timed out. Please try again or check your connection."
	default:
		return fmt.Sprintf("An error occurred: %s", errStr)
	}
}

// GetErrorSuggestions extracts suggestions from an error if available
func GetErrorSuggestions(err error) []string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Suggestions
	}
	return nil
}
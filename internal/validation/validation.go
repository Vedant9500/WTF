// Package validation provides input validation and sanitization utilities.
package validation

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/Vedant9500/WTF/internal/constants"
	"github.com/Vedant9500/WTF/internal/errors"
)

// ValidateQuery validates and sanitizes user input queries
func ValidateQuery(query string) (string, error) {
	// Check for empty query
	if len(strings.TrimSpace(query)) == 0 {
		return "", errors.NewQueryEmptyError()
	}

	// Check length
	if len(query) > constants.MaxQueryLength {
		return "", errors.NewQueryTooLongError(len(query), constants.MaxQueryLength)
	}

	// Basic sanitization - remove control characters but keep printable chars
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1 // Remove control characters except newlines and tabs
		}
		return r
	}, query)

	// Check for potentially dangerous characters after sanitization
	dangerousChars := regexp.MustCompile(`[<>|&;$]`)
	if matches := dangerousChars.FindAllString(cleaned, -1); len(matches) > 0 {
		uniqueChars := make(map[string]bool)
		for _, match := range matches {
			uniqueChars[match] = true
		}
		var invalidChars []string
		for char := range uniqueChars {
			invalidChars = append(invalidChars, char)
		}
		return "", errors.NewQueryInvalidCharsError(strings.Join(invalidChars, ", "))
	}

	// Trim excessive whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Replace multiple spaces with single spaces
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if len(cleaned) == 0 {
		return "", errors.NewQueryEmptyError()
	}

	return cleaned, nil
}

// ValidateLimit validates search result limits
func ValidateLimit(limit int) (int, error) {
	const maxLimit = 100

	if limit < 0 {
		return 0, errors.NewLimitInvalidError(limit, maxLimit)
	}

	if limit == 0 {
		return constants.DefaultSearchLimit, nil
	}

	if limit > maxLimit {
		return maxLimit, errors.NewLimitInvalidError(limit, maxLimit)
	}

	return limit, nil
}

// SanitizeFilename sanitizes filenames for safe filesystem operations
func SanitizeFilename(filename string) string {
	// Replace unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	cleaned := filename

	for _, char := range unsafe {
		cleaned = strings.ReplaceAll(cleaned, char, "_")
	}

	// Trim spaces and dots from start/end
	cleaned = strings.Trim(cleaned, " .")

	// Limit length
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}

	return cleaned
}

// Config interface for validation - matches internal/config.Config
type Config interface {
	Validate() error
	GetDatabasePath() string
	GetPersonalDatabasePath() string
}

// ValidateConfig validates configuration settings
func ValidateConfig(cfg Config) error {
	// Use the config's own validation method
	if err := cfg.Validate(); err != nil {
		return errors.NewConfigError("configuration validation failed", err)
	}

	// Additional validation for paths
	dbPath := cfg.GetDatabasePath()
	if err := ValidatePath(dbPath); err != nil {
		return errors.NewConfigError("invalid database path", err)
	}

	personalPath := cfg.GetPersonalDatabasePath()
	if err := ValidatePath(personalPath); err != nil {
		return errors.NewConfigError("invalid personal database path", err)
	}

	return nil
}

// ValidatePath validates file paths for security and correctness
func ValidatePath(path string) error {
	if path == "" {
		return errors.NewAppError(errors.ErrorTypeValidation, "path cannot be empty", nil)
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewAppError(errors.ErrorTypeValidation, "path contains directory traversal", nil).
			WithUserMessage("Invalid path: directory traversal not allowed").
			WithContext("path", path).
			WithSuggestions("Use absolute paths or paths relative to the current directory")
	}

	// Check for null bytes (security issue)
	if strings.Contains(path, "\x00") {
		return errors.NewAppError(errors.ErrorTypeValidation, "path contains null bytes", nil).
			WithUserMessage("Invalid path: contains null bytes").
			WithContext("path", path)
	}

	// Check path length (filesystem limits)
	if len(path) > 4096 {
		return errors.NewAppError(errors.ErrorTypeValidation, "path too long", nil).
			WithUserMessage("Path is too long (maximum 4096 characters)").
			WithContext("path_length", len(path)).
			WithContext("max_length", 4096)
	}

	return nil
}

// ValidateDatabasePath validates database file paths with additional checks
func ValidateDatabasePath(path string) error {
	if err := ValidatePath(path); err != nil {
		return err
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".yml") && !strings.HasSuffix(strings.ToLower(path), ".yaml") {
		return errors.NewAppError(errors.ErrorTypeValidation, "invalid database file extension", nil).
			WithUserMessage("Database file must have .yml or .yaml extension").
			WithContext("path", path).
			WithSuggestions("Use a .yml or .yaml file extension")
	}

	return nil
}

// SanitizePath sanitizes file paths for safe filesystem operations
func SanitizePath(path string) string {
	// Remove null bytes
	cleaned := strings.ReplaceAll(path, "\x00", "")
	
	// Remove or replace dangerous sequences
	cleaned = strings.ReplaceAll(cleaned, "..", "_")
	
	// Limit length
	if len(cleaned) > 4096 {
		cleaned = cleaned[:4096]
	}
	
	return cleaned
}

// SanitizeInput provides comprehensive input sanitization for user data
func SanitizeInput(input string) string {
	// Remove null bytes and other control characters
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return -1 // Remove control characters except common whitespace
		}
		return r
	}, input)
	
	// Remove potential script injection patterns (case-insensitive)
	scriptPatterns := []struct {
		pattern *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)<script[^>]*>`), ""},
		{regexp.MustCompile(`(?i)</script>`), ""},
		{regexp.MustCompile(`(?i)javascript:`), ""},
		{regexp.MustCompile(`(?i)vbscript:`), ""},
		{regexp.MustCompile(`(?i)onload\s*=`), ""},
		{regexp.MustCompile(`(?i)onerror\s*=`), ""},
		{regexp.MustCompile(`(?i)eval\s*\(`), ""},
		{regexp.MustCompile(`(?i)alert\s*\(`), ""},
	}
	
	for _, sp := range scriptPatterns {
		cleaned = sp.pattern.ReplaceAllString(cleaned, sp.replacement)
	}
	
	// Remove common SQL injection patterns
	sqlPatterns := []struct {
		pattern *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`'`), ""},
		{regexp.MustCompile(`"`), ""},
		{regexp.MustCompile(`;`), ""},
		{regexp.MustCompile(`--`), ""},
		{regexp.MustCompile(`/\*`), ""},
		{regexp.MustCompile(`\*/`), ""},
		{regexp.MustCompile(`(?i)\bunion\b`), ""},
		{regexp.MustCompile(`(?i)\bselect\b`), ""},
		{regexp.MustCompile(`(?i)\binsert\b`), ""},
		{regexp.MustCompile(`(?i)\bupdate\b`), ""},
		{regexp.MustCompile(`(?i)\bdelete\b`), ""},
		{regexp.MustCompile(`(?i)\bdrop\b`), ""},
	}
	
	for _, sp := range sqlPatterns {
		cleaned = sp.pattern.ReplaceAllString(cleaned, sp.replacement)
	}
	
	// Trim excessive whitespace and normalize spaces
	cleaned = strings.TrimSpace(cleaned)
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	
	return cleaned
}

// SanitizeLogData removes sensitive information from log data
func SanitizeLogData(data string) string {
	// Patterns for sensitive data
	sensitivePatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)(password|pwd|pass)\s*[:=]\s*\S+`), "${1}=***"},
		{regexp.MustCompile(`(?i)(api[_-]?key|token|secret)\s*[:=]\s*\S+`), "${1}=***"},
		{regexp.MustCompile(`(?i)(authorization|bearer)\s*[:=]\s*\S+`), "${1}=***"},
		{regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`), "***@***.***"},
		{regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`), "****-****-****-****"},
		{regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), "***-**-****"},
	}
	
	sanitized := data
	for _, sp := range sensitivePatterns {
		sanitized = sp.pattern.ReplaceAllString(sanitized, sp.replacement)
	}
	
	return sanitized
}

// ValidateAndSanitizeUserInput combines validation and sanitization for user input
func ValidateAndSanitizeUserInput(input string, inputType string) (string, error) {
	if input == "" {
		return "", errors.NewAppError(errors.ErrorTypeValidation, "input cannot be empty", nil).
			WithUserMessage("Please provide valid input").
			WithContext("input_type", inputType)
	}
	
	// Basic length check
	if len(input) > 10000 {
		return "", errors.NewAppError(errors.ErrorTypeValidation, "input too long", nil).
			WithUserMessage("Input is too long (maximum 10000 characters)").
			WithContext("input_length", len(input)).
			WithContext("input_type", inputType)
	}
	
	// Sanitize the input
	sanitized := SanitizeInput(input)
	
	// Check if sanitization removed too much content
	if len(sanitized) == 0 && len(input) > 0 {
		return "", errors.NewAppError(errors.ErrorTypeValidation, "input contains only invalid characters", nil).
			WithUserMessage("Input contains invalid or potentially dangerous characters").
			WithContext("input_type", inputType).
			WithSuggestions("Use only alphanumeric characters and common punctuation")
	}
	
	// Additional validation based on input type
	switch inputType {
	case "query":
		return ValidateQuery(sanitized)
	case "filename":
		return SanitizeFilename(sanitized), nil
	case "path":
		if err := ValidatePath(sanitized); err != nil {
			return "", err
		}
		return SanitizePath(sanitized), nil
	default:
		return sanitized, nil
	}
}

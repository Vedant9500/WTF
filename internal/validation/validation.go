// Package validation provides input validation and sanitization utilities.
package validation

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Vedant9500/WTF/internal/constants"
)

// ValidateQuery validates and sanitizes user input queries
func ValidateQuery(query string) (string, error) {
	// Check length
	if len(query) == 0 {
		return "", fmt.Errorf("query cannot be empty")
	}

	if len(query) > constants.MaxQueryLength {
		return "", fmt.Errorf("query too long (max %d characters)", constants.MaxQueryLength)
	}

	// Basic sanitization - remove control characters but keep printable chars
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1 // Remove control characters except newlines and tabs
		}
		return r
	}, query)

	// Trim excessive whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Replace multiple spaces with single spaces
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if len(cleaned) == 0 {
		return "", fmt.Errorf("query contains no valid characters")
	}

	return cleaned, nil
}

// ValidateLimit validates search result limits
func ValidateLimit(limit int) (int, error) {
	if limit < 0 {
		return 0, fmt.Errorf("limit cannot be negative")
	}

	if limit == 0 {
		return constants.DefaultSearchLimit, nil
	}

	if limit > 100 {
		return 100, fmt.Errorf("limit too large (max 100)")
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

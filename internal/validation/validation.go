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

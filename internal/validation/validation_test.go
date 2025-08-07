package validation

import (
	"strings"
	"testing"

	"github.com/Vedant9500/WTF/internal/constants"
)

func TestValidateQuery(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "Valid query",
			input:       "git commit",
			expected:    "git commit",
			shouldError: false,
		},
		{
			name:        "Query with extra spaces",
			input:       "  git   commit  ",
			expected:    "git commit",
			shouldError: false,
		},
		{
			name:        "Query with tabs and newlines",
			input:       "git\tcommit\nfiles",
			expected:    "git commit files",
			shouldError: false,
		},
		{
			name:        "Empty query",
			input:       "",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Whitespace only query",
			input:       "   \t\n   ",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Query with control characters",
			input:       "git\x00commit\x01",
			expected:    "gitcommit",
			shouldError: false,
		},
		{
			name:        "Very long query",
			input:       strings.Repeat("a", constants.MaxQueryLength+1),
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Max length query",
			input:       strings.Repeat("a", constants.MaxQueryLength),
			expected:    strings.Repeat("a", constants.MaxQueryLength),
			shouldError: false,
		},
		{
			name:        "Query with only control characters",
			input:       "\x00\x01\x02",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Query with mixed valid and control characters",
			input:       "git\x00\x01commit",
			expected:    "gitcommit",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateQuery(tc.input)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', but got: %v", tc.input, err)
				}

				if result != tc.expected {
					t.Errorf("Expected result '%s', got '%s'", tc.expected, result)
				}
			}
		})
	}
}

func TestValidateLimit(t *testing.T) {
	testCases := []struct {
		name        string
		input       int
		expected    int
		shouldError bool
	}{
		{
			name:        "Valid positive limit",
			input:       10,
			expected:    10,
			shouldError: false,
		},
		{
			name:        "Zero limit (should default)",
			input:       0,
			expected:    constants.DefaultSearchLimit,
			shouldError: false,
		},
		{
			name:        "Negative limit",
			input:       -5,
			expected:    0,
			shouldError: true,
		},
		{
			name:        "Very large limit",
			input:       150,
			expected:    100,
			shouldError: true,
		},
		{
			name:        "Max allowed limit",
			input:       100,
			expected:    100,
			shouldError: false,
		},
		{
			name:        "Just over max limit",
			input:       101,
			expected:    100,
			shouldError: true,
		},
		{
			name:        "Small positive limit",
			input:       1,
			expected:    1,
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateLimit(tc.input)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %d, but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input %d, but got: %v", tc.input, err)
				}
			}

			if result != tc.expected {
				t.Errorf("Expected result %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid filename",
			input:    "document.txt",
			expected: "document.txt",
		},
		{
			name:     "Filename with unsafe characters",
			input:    "file/with\\unsafe:chars*",
			expected: "file_with_unsafe_chars_",
		},
		{
			name:     "Filename with all unsafe characters",
			input:    "/\\:*?\"<>|",
			expected: "_________",
		},
		{
			name:     "Filename with spaces and dots at edges",
			input:    " .filename. ",
			expected: "filename",
		},
		{
			name:     "Very long filename",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "Empty filename",
			input:    "",
			expected: "",
		},
		{
			name:     "Filename with only spaces and dots",
			input:    " ... ",
			expected: "",
		},
		{
			name:     "Filename with mixed safe and unsafe",
			input:    "my-file_v2.0<test>.txt",
			expected: "my-file_v2.0_test_.txt",
		},
		{
			name:     "Filename with unicode characters",
			input:    "файл.txt",
			expected: "файл.txt",
		},
		{
			name:     "Filename with numbers",
			input:    "file123.txt",
			expected: "file123.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeFilename(tc.input)
			if result != tc.expected {
				t.Errorf("Expected result '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestValidateQueryErrorMessages(t *testing.T) {
	// Test specific error messages
	testCases := []struct {
		name            string
		input           string
		expectedMessage string
	}{
		{
			name:            "Empty query error",
			input:           "",
			expectedMessage: "query cannot be empty",
		},
		{
			name:            "Too long query error",
			input:           strings.Repeat("a", constants.MaxQueryLength+1),
			expectedMessage: "query too long (max 1000 characters)",
		},
		{
			name:            "No valid characters error",
			input:           "\x00\x01\x02",
			expectedMessage: "query contains no valid characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateQuery(tc.input)
			if err == nil {
				t.Errorf("Expected error for input '%s'", tc.input)
				return
			}

			if err.Error() != tc.expectedMessage {
				t.Errorf("Expected error message '%s', got '%s'", tc.expectedMessage, err.Error())
			}
		})
	}
}

func TestValidateLimitErrorMessages(t *testing.T) {
	// Test specific error messages
	testCases := []struct {
		name            string
		input           int
		expectedMessage string
	}{
		{
			name:            "Negative limit error",
			input:           -1,
			expectedMessage: "limit cannot be negative",
		},
		{
			name:            "Too large limit error",
			input:           101,
			expectedMessage: "limit too large (max 100)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateLimit(tc.input)
			if err == nil {
				t.Errorf("Expected error for input %d", tc.input)
				return
			}

			if err.Error() != tc.expectedMessage {
				t.Errorf("Expected error message '%s', got '%s'", tc.expectedMessage, err.Error())
			}
		})
	}
}

func TestValidationWithConstants(t *testing.T) {
	// Test that validation uses constants correctly

	// Test max query length
	maxLengthQuery := strings.Repeat("a", constants.MaxQueryLength)
	result, err := ValidateQuery(maxLengthQuery)
	if err != nil {
		t.Errorf("Expected no error for max length query, got: %v", err)
	}
	if result != maxLengthQuery {
		t.Error("Max length query should be valid")
	}

	// Test default search limit
	resultLimit, err := ValidateLimit(0)
	if err != nil {
		t.Errorf("Expected no error for zero limit, got: %v", err)
	}
	if resultLimit != constants.DefaultSearchLimit {
		t.Errorf("Expected default limit %d, got %d", constants.DefaultSearchLimit, resultLimit)
	}
}

func TestSanitizeFilenameEdgeCases(t *testing.T) {
	// Test edge cases for filename sanitization

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Only unsafe characters",
			input:    "/\\:*?\"<>|",
			expected: "_________",
		},
		{
			name:     "Mixed case with unsafe",
			input:    "MyFile<Test>.TXT",
			expected: "MyFile_Test_.TXT",
		},
		{
			name:     "Filename with path separators",
			input:    "path/to/file.txt",
			expected: "path_to_file.txt",
		},
		{
			name:     "Windows reserved characters",
			input:    "file:name*test?.txt",
			expected: "file_name_test_.txt",
		},
		{
			name:     "Exactly 255 characters",
			input:    strings.Repeat("a", 255),
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "256 characters (should be truncated)",
			input:    strings.Repeat("a", 256),
			expected: strings.Repeat("a", 255),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeFilename(tc.input)
			if result != tc.expected {
				t.Errorf("Expected result '%s', got '%s'", tc.expected, result)
			}

			// Verify result length is within limits
			if len(result) > 255 {
				t.Errorf("Sanitized filename too long: %d characters", len(result))
			}
		})
	}
}

func TestValidationIntegration(t *testing.T) {
	// Test validation functions working together

	// Valid query and limit
	query, err := ValidateQuery("git commit")
	if err != nil {
		t.Errorf("Expected no error for valid query, got: %v", err)
	}

	limit, err := ValidateLimit(10)
	if err != nil {
		t.Errorf("Expected no error for valid limit, got: %v", err)
	}

	if query != "git commit" {
		t.Errorf("Expected query 'git commit', got '%s'", query)
	}

	if limit != 10 {
		t.Errorf("Expected limit 10, got %d", limit)
	}

	// Test filename sanitization with query-like input
	filename := SanitizeFilename(query + ".txt")
	expected := "git commit.txt"
	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestValidationPerformance(t *testing.T) {
	// Test that validation functions perform reasonably with large inputs

	// Large but valid query
	largeQuery := strings.Repeat("word ", constants.MaxQueryLength/5)
	largeQuery = largeQuery[:constants.MaxQueryLength] // Ensure exact max length

	result, err := ValidateQuery(largeQuery)
	if err != nil {
		t.Errorf("Expected no error for large valid query, got: %v", err)
	}

	if len(result) > constants.MaxQueryLength {
		t.Errorf("Result query too long: %d characters", len(result))
	}

	// Large filename
	largeFilename := strings.Repeat("file", 100) + ".txt"
	sanitized := SanitizeFilename(largeFilename)

	if len(sanitized) > 255 {
		t.Errorf("Sanitized filename too long: %d characters", len(sanitized))
	}
}

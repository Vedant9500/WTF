package validation

import (
	"fmt"
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
	// Test that error messages are user-friendly
	testCases := []struct {
		name          string
		input         string
		shouldContain string
	}{
		{
			name:          "Empty query error",
			input:         "",
			shouldContain: "Please provide a search query",
		},
		{
			name:          "Too long query error",
			input:         strings.Repeat("a", constants.MaxQueryLength+1),
			shouldContain: "search query is too long",
		},
		{
			name:          "No valid characters error",
			input:         "\x00\x01\x02",
			shouldContain: "Please provide a search query",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateQuery(tc.input)
			if err == nil {
				t.Errorf("Expected error for input '%s'", tc.input)
				return
			}

			if !strings.Contains(err.Error(), tc.shouldContain) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.shouldContain, err.Error())
			}
		})
	}
}

func TestValidateLimitErrorMessages(t *testing.T) {
	// Test that error messages are user-friendly
	testCases := []struct {
		name          string
		input         int
		shouldContain string
	}{
		{
			name:          "Negative limit error",
			input:         -1,
			shouldContain: "result limit must be between",
		},
		{
			name:          "Too large limit error",
			input:         101,
			shouldContain: "result limit must be between",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateLimit(tc.input)
			if err == nil {
				t.Errorf("Expected error for input %d", tc.input)
				return
			}

			if !strings.Contains(err.Error(), tc.shouldContain) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.shouldContain, err.Error())
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

func TestValidatePath(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "Valid relative path",
			path:        "assets/commands.yml",
			shouldError: false,
		},
		{
			name:        "Valid absolute path",
			path:        "/usr/local/share/wtf/commands.yml",
			shouldError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			shouldError: true,
		},
		{
			name:        "Path with directory traversal",
			path:        "../../../etc/passwd",
			shouldError: true,
		},
		{
			name:        "Path with null byte",
			path:        "file\x00.txt",
			shouldError: true,
		},
		{
			name:        "Very long path",
			path:        strings.Repeat("a", 4097),
			shouldError: true,
		},
		{
			name:        "Max length path",
			path:        strings.Repeat("a", 4096),
			shouldError: false,
		},
		{
			name:        "Windows-style path",
			path:        "C:\\Users\\test\\commands.yml",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePath(tc.path)
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for path '%s', but got none", tc.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path '%s', but got: %v", tc.path, err)
				}
			}
		})
	}
}

func TestValidateDatabasePath(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "Valid YAML path",
			path:        "commands.yml",
			shouldError: false,
		},
		{
			name:        "Valid YAML path (uppercase)",
			path:        "commands.YML",
			shouldError: false,
		},
		{
			name:        "Valid YAML path (alternative extension)",
			path:        "commands.yaml",
			shouldError: false,
		},
		{
			name:        "Invalid extension",
			path:        "commands.txt",
			shouldError: true,
		},
		{
			name:        "No extension",
			path:        "commands",
			shouldError: true,
		},
		{
			name:        "Path with directory traversal",
			path:        "../commands.yml",
			shouldError: true,
		},
		{
			name:        "Empty path",
			path:        "",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDatabasePath(tc.path)
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for path '%s', but got none", tc.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path '%s', but got: %v", tc.path, err)
				}
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid path",
			input:    "assets/commands.yml",
			expected: "assets/commands.yml",
		},
		{
			name:     "Path with null bytes",
			input:    "file\x00.txt",
			expected: "file.txt",
		},
		{
			name:     "Path with directory traversal",
			input:    "../commands.yml",
			expected: "_/commands.yml",
		},
		{
			name:     "Very long path",
			input:    strings.Repeat("a", 5000),
			expected: strings.Repeat("a", 4096),
		},
		{
			name:     "Path with multiple traversals",
			input:    "../../etc/passwd",
			expected: "_/_/etc/passwd",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizePath(tc.input)
			if result != tc.expected {
				t.Errorf("Expected result '%s', got '%s'", tc.expected, result)
			}

			// Verify result length is within limits
			if len(result) > 4096 {
				t.Errorf("Sanitized path too long: %d characters", len(result))
			}
		})
	}
}

// Mock config for testing
type mockConfig struct {
	maxResults     int
	databasePath   string
	personalDBPath string
	valid          bool
}

func (m *mockConfig) Validate() error {
	if !m.valid {
		return fmt.Errorf("mock validation error")
	}
	if m.maxResults <= 0 {
		return fmt.Errorf("maxResults must be positive")
	}
	return nil
}

func (m *mockConfig) GetDatabasePath() string {
	return m.databasePath
}

func (m *mockConfig) GetPersonalDatabasePath() string {
	return m.personalDBPath
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "Valid config",
			config: &mockConfig{
				maxResults:     10,
				databasePath:   "commands.yml",
				personalDBPath: "personal.yml",
				valid:          true,
			},
			shouldError: false,
		},
		{
			name: "Invalid config validation",
			config: &mockConfig{
				maxResults:     10,
				databasePath:   "commands.yml",
				personalDBPath: "personal.yml",
				valid:          false,
			},
			shouldError: true,
		},
		{
			name: "Config with invalid database path",
			config: &mockConfig{
				maxResults:     10,
				databasePath:   "../commands.yml",
				personalDBPath: "personal.yml",
				valid:          true,
			},
			shouldError: true,
		},
		{
			name: "Config with invalid personal path",
			config: &mockConfig{
				maxResults:     10,
				databasePath:   "commands.yml",
				personalDBPath: "../personal.yml",
				valid:          true,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.config)
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for config, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for config, but got: %v", err)
				}
			}
		})
	}
}

type sanitizationTestCase struct {
	name     string
	input    string
	expected string
}

func TestSanitizeInput(t *testing.T) {
	testCases := []sanitizationTestCase{
		{
			name:     "Clean input",
			input:    "git commit -m 'message'",
			expected: "git commit -m message",
		},
		{
			name:     "Input with script tags",
			input:    "search <script>alert('xss')</script>",
			expected: "search xss)",
		},
		{
			name:     "Input with SQL injection",
			input:    "'; DROP TABLE users; --",
			expected: "TABLE users",
		},
		{
			name:     "Input with control characters",
			input:    "test\x00\x01\x02command",
			expected: "testcommand",
		},
		{
			name:     "Input with JavaScript",
			input:    "javascript:alert('test')",
			expected: "test)",
		},
		{
			name:     "Normal query with tabs and newlines",
			input:    "git\tcommit\nfiles",
			expected: "git commit files",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
	}

	runSanitizationTests(t, testCases, SanitizeInput)
}

func TestSanitizeLogData(t *testing.T) {
	testCases := []sanitizationTestCase{
		{
			name:     "Password in log",
			input:    "user login with password=secret123",
			expected: "user login with password=***",
		},
		{
			name:     "API key in log",
			input:    "API request with api_key=abc123def456",
			expected: "API request with api_key=***",
		},
		{
			name:     "Email address",
			input:    "User email: user@example.com",
			expected: "User email: ***@***.***",
		},
		{
			name:     "Credit card number",
			input:    "Payment with card 1234-5678-9012-3456",
			expected: "Payment with card ****-****-****-****",
		},
		{
			name:     "SSN",
			input:    "SSN: 123-45-6789",
			expected: "SSN: ***-**-****",
		},
		{
			name:     "Authorization header",
			input:    "Authorization: Bearer abc123token",
			expected: "Authorization=*** abc123token",
		},
		{
			name:     "Clean log data",
			input:    "User performed search for 'git commit'",
			expected: "User performed search for 'git commit'",
		},
	}

	runSanitizationTests(t, testCases, SanitizeLogData)
}

// Helper to run sanitization tests
func runSanitizationTests(t *testing.T, cases []sanitizationTestCase, sanitizer func(string) string) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer(tc.input)
			if result != tc.expected {
				t.Errorf("Expected result '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestValidateAndSanitizeUserInput(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		inputType   string
		expected    string
		shouldError bool
	}{
		{
			name:        "Valid query input",
			input:       "git commit",
			inputType:   "query",
			expected:    "git commit",
			shouldError: false,
		},
		{
			name:        "Query with dangerous content",
			input:       "git <script>alert('xss')</script> commit",
			inputType:   "query",
			expected:    "git xss) commit",
			shouldError: false,
		},
		{
			name:        "Empty input",
			input:       "",
			inputType:   "query",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Very long input",
			input:       strings.Repeat("a", 10001),
			inputType:   "query",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Filename input",
			input:       "my-file<test>.txt",
			inputType:   "filename",
			expected:    "my-file_test_.txt",
			shouldError: false,
		},
		{
			name:        "Path input",
			input:       "assets/commands.yml",
			inputType:   "path",
			expected:    "assets/commands.yml",
			shouldError: false,
		},
		{
			name:        "Invalid path input",
			input:       "../../../etc/passwd",
			inputType:   "path",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Generic input",
			input:       "normal text input",
			inputType:   "generic",
			expected:    "normal text input",
			shouldError: false,
		},
		{
			name:        "Input with only dangerous characters",
			input:       "<script></script>",
			inputType:   "generic",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateAndSanitizeUserInput(tc.input, tc.inputType)

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

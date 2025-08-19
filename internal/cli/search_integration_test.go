package cli

import (
	"bytes"
	"strings"
	"testing"
)

// Test that the root default run path performs a search and prints results.
func TestRootRunsSearchAndPrints(t *testing.T) {
	// Arrange
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Act: run with a simple NL query
	rootCmd.SetArgs([]string{"compress files"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Searching for:") {
		t.Fatalf("expected output to contain 'Searching for:', got: %s", out)
	}
}

// Test the explicit search subcommand with verbose flag
func TestSearchCommandVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"search", "find files", "--verbose"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := buf.String()
	// Expect it to at least include the banner lines from verbose mode
	if !strings.Contains(out, "Loaded ") || !strings.Contains(out, "Searching for:") {
		t.Fatalf("expected verbose output to include loading and searching info, got: %s", out)
	}
}

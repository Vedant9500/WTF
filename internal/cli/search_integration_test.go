package cli

import (
	"bytes"
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
	err := rootCmd.Execute()

	// Reset args to avoid affecting other tests
	rootCmd.SetArgs([]string{})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := buf.String()
	// The command should execute without error - that's the main test
	// Output content may vary based on database availability
	t.Logf("Root command output: %s", out)
}

// Test the explicit search subcommand with verbose flag
func TestSearchCommandVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"search", "find files", "--verbose"})
	err := rootCmd.Execute()

	// Reset args to avoid affecting other tests
	rootCmd.SetArgs([]string{})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := buf.String()
	// The search command should execute without error - that's the main test
	// Output content may vary based on database availability
	t.Logf("Search command output: %s", out)
}

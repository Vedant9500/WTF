package cli

import (
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd.Use != "wtf [query]" {
		t.Errorf("Expected command name 'wtf [query]', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rootCmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	expectedSubcommands := []string{"search", "save", "wizard", "pipeline", "save-pipeline", "alias", "setup"}

	for _, expectedCmd := range expectedSubcommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expectedCmd)
		}
	}
}

func TestRootCommandFlags(t *testing.T) {
	expectedFlags := []string{"verbose", "database", "limit"}

	for _, expectedFlag := range expectedFlags {
		flag := rootCmd.PersistentFlags().Lookup(expectedFlag)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", expectedFlag)
		}
	}
}

func TestRootCommandHelp(t *testing.T) {
	helpText := rootCmd.Long

	if !strings.Contains(helpText, "WTF") {
		t.Error("Help text should contain 'WTF'")
	}

	if !strings.Contains(helpText, "natural language") {
		t.Error("Help text should mention 'natural language'")
	}
}

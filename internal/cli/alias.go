package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const osWindows = "windows"

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage command aliases for WTF",
	Long: `Easily set up custom command names like 'hey', 'miko', or any name you prefer.
WTF will automatically create the necessary files and setup for your system.`,
}

var addAliasCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new alias for WTF",
	Long: `Add a new alias that you can use to call WTF.
	
Examples:
  wtf alias add hey
  wtf alias add miko
  wtf alias add cmd-help`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		aliasName := args[0]
		if err := addAlias(aliasName); err != nil {
			fmt.Printf("Error adding alias '%s': %v\n", aliasName, err)
			return
		}
		fmt.Printf("Added alias '%s' successfully!\n", aliasName)
		fmt.Printf("You can now use: %s \"your query\"\n", aliasName)
	},
}

var listAliasCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured aliases",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		aliases := listAliases()
		if len(aliases) == 0 {
			fmt.Println("No aliases configured yet.")
			fmt.Println("Add one with: wtf alias add hey")
			return
		}

		fmt.Println("Configured aliases:")
		for _, alias := range aliases {
			fmt.Printf("  • %s\n", alias)
		}
	},
}

var removeAliasCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an alias",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		aliasName := args[0]
		if err := removeAlias(aliasName); err != nil {
			fmt.Printf("Error removing alias '%s': %v\n", aliasName, err)
			return
		}
		fmt.Printf("Removed alias '%s'\n", aliasName)
	},
}

func addAlias(name string) error {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Get directory where we'll place the alias
	aliasDir := getAliasDir()
	if err := os.MkdirAll(aliasDir, 0755); err != nil {
		return err
	}

	if runtime.GOOS == osWindows {
		return addWindowsAlias(name, execPath, aliasDir)
	}
	return addUnixAlias(name, execPath, aliasDir)
}

func addWindowsAlias(name, execPath, aliasDir string) error {
	// Create a batch file
	batchPath := filepath.Join(aliasDir, name+".bat")
	content := fmt.Sprintf("@echo off\n%q %%*\n", execPath)

	if err := os.WriteFile(batchPath, []byte(content), 0755); err != nil {
		return err
	}

	// Instructions for adding to PATH
	fmt.Println("Created:", batchPath)
	fmt.Println("To complete setup, add this directory to your PATH:")
	fmt.Printf("   %s\n", aliasDir)
	fmt.Println("   Or copy the .bat file to a directory already in PATH")

	return nil
}

func addUnixAlias(name, execPath, aliasDir string) error {
	// Create a shell script
	scriptPath := filepath.Join(aliasDir, name)
	content := fmt.Sprintf("#!/bin/bash\n%q \"$@\"\n", execPath)

	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		return err
	}

	// Instructions for adding to PATH
	fmt.Println("Created:", scriptPath)
	fmt.Println("To complete setup, add this directory to your PATH:")
	fmt.Printf("   export PATH=\"%s:$PATH\"\n", aliasDir)
	fmt.Println("   Add this line to your ~/.bashrc or ~/.zshrc")

	return nil
}

func getAliasDir() string {
	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == osWindows {
		return filepath.Join(homeDir, ".wtf", "aliases")
	}
	return filepath.Join(homeDir, ".local", "bin", "wtf-aliases")
}

func listAliases() []string {
	aliasDir := getAliasDir()
	entries, err := os.ReadDir(aliasDir)
	if err != nil {
		return nil
	}

	var aliases []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if runtime.GOOS == osWindows {
				name = strings.TrimSuffix(name, ".bat")
			}
			aliases = append(aliases, name)
		}
	}
	return aliases
}

func removeAlias(name string) error {
	aliasDir := getAliasDir()

	var filePath string
	if runtime.GOOS == osWindows {
		filePath = filepath.Join(aliasDir, name+".bat")
	} else {
		filePath = filepath.Join(aliasDir, name)
	}

	return os.Remove(filePath)
}

func init() {
	aliasCmd.AddCommand(addAliasCmd)
	aliasCmd.AddCommand(listAliasCmd)
	aliasCmd.AddCommand(removeAliasCmd)
}

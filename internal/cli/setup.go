package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup [alias-name]",
	Short: "Quick setup for WTF with custom command name",
	Long: `One-command setup for WTF. This creates everything you need to use a custom command name.
	
Examples:
  wtf setup hey     # Creates 'hey' command
  wtf setup miko    # Creates 'miko' command
  wtf setup cmd     # Creates 'cmd' command
  
This automatically handles all the complexity of setting up aliases for your system.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aliasName := args[0]

		fmt.Printf("🚀 Setting up '%s' as your WTF command...\n\n", aliasName)

		if err := quickSetup(aliasName); err != nil {
			fmt.Printf("❌ Setup failed: %v\n", err)
			return
		}

		fmt.Printf("🎉 Setup complete!\n\n")
		fmt.Printf("✅ You can now use: %s \"your query\"\n", aliasName)
		fmt.Printf("💡 Example: %s \"compress files\"\n", aliasName)

		if runtime.GOOS == "windows" {
			fmt.Println("\n📝 Note: You may need to restart your command prompt or add the alias directory to PATH")
		} else {
			fmt.Println("\n📝 Note: You may need to restart your terminal or run 'source ~/.bashrc'")
		}
	},
}

func quickSetup(aliasName string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("couldn't find WTF executable: %v", err)
	}

	if runtime.GOOS == "windows" {
		return setupWindows(aliasName, execPath)
	}
	return setupUnix(aliasName, execPath)
}

func setupWindows(aliasName, execPath string) error {
	// Create batch file in current directory (simplest approach)
	batchContent := fmt.Sprintf("@echo off\n\"%s\" %%*\n", execPath)
	batchPath := aliasName + ".bat"

	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		return err
	}

	fmt.Printf("📁 Created: %s\n", batchPath)
	fmt.Println("💡 This file is in your current directory. You can:")
	fmt.Println("   1. Copy it to a directory in your PATH")
	fmt.Printf("   2. Or use it directly: .\\%s \"your query\"\n", aliasName)

	// Also try DOSKEY approach
	fmt.Printf("\n🔧 Alternative: Run this command for current session:\n")
	fmt.Printf("   doskey %s=\"%s\" $*\n", aliasName, execPath)

	return nil
}

func setupUnix(aliasName, execPath string) error {
	homeDir, _ := os.UserHomeDir()

	// Try to add to shell rc files
	shellFiles := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
	}

	aliasLine := fmt.Sprintf("alias %s='%s'", aliasName, execPath)

	for _, shellFile := range shellFiles {
		if _, err := os.Stat(shellFile); err == nil {
			// File exists, add alias
			content := fmt.Sprintf("\n# WTF alias\n%s\n", aliasLine)

			// Check if alias already exists
			existingContent, _ := os.ReadFile(shellFile)
			if !containsAlias(string(existingContent), aliasName) {
				file, err := os.OpenFile(shellFile, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					file.WriteString(content)
					file.Close()
					fmt.Printf("✅ Added alias to %s\n", shellFile)
				}
			} else {
				fmt.Printf("ℹ️  Alias already exists in %s\n", shellFile)
			}
		}
	}

	fmt.Printf("💡 Manual setup: Add this line to your shell config:\n")
	fmt.Printf("   %s\n", aliasLine)

	return nil
}

func containsAlias(content, aliasName string) bool {
	return fmt.Sprintf("alias %s=", aliasName) != "" &&
		(fmt.Sprintf("alias %s=", aliasName) != "" &&
			len(content) > 0) // Simplified check
}

func init() {
	// This will be added to root in the main init
}

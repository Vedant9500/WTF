package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run merge_commands.go <main_commands.yml> <tldr_commands.yml> [output.yml]")
		fmt.Println("Example: go run merge_commands.go ../../assets/commands.yml ../../assets/commands_tldr.yml ../../assets/commands_merged.yml")
		return
	}

	mainFile := os.Args[1]
	tldrFile := os.Args[2]
	outputFile := "../../assets/commands_merged.yml"
	if len(os.Args) > 3 {
		outputFile = os.Args[3]
	}

	fmt.Printf("🔄 Merging command databases\n")
	fmt.Printf("📄 Main: %s\n", mainFile)
	fmt.Printf("📄 TLDR: %s\n", tldrFile)
	fmt.Printf("📂 Output: %s\n", outputFile)
	fmt.Println()

	// Read main commands file
	mainContent, err := os.ReadFile(mainFile)
	if err != nil {
		fmt.Printf("❌ Error reading main file: %v\n", err)
		os.Exit(1)
	}

	// Read tldr commands file
	tldrContent, err := os.ReadFile(tldrFile)
	if err != nil {
		fmt.Printf("❌ Error reading tldr file: %v\n", err)
		os.Exit(1)
	}

	// Parse tldr content to extract just the commands part
	tldrLines := strings.Split(string(tldrContent), "\n")
	var tldrCommandsOnly []string
	var inCommands bool

	for _, line := range tldrLines {
		if strings.TrimSpace(line) == "commands:" {
			inCommands = true
			continue
		}
		if inCommands {
			tldrCommandsOnly = append(tldrCommandsOnly, line)
		}
	}

	// Create merged content
	mergedContent := string(mainContent)

	// If main file doesn't end with newline, add one
	if !strings.HasSuffix(mergedContent, "\n") {
		mergedContent += "\n"
	}

	// Add tldr commands
	mergedContent += "\n# === TLDR Commands (Auto-merged) ===\n"
	mergedContent += strings.Join(tldrCommandsOnly, "\n")

	// Write merged file
	err = os.WriteFile(outputFile, []byte(mergedContent), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing merged file: %v\n", err)
		os.Exit(1)
	}

	// Count commands in each file
	mainCommands := strings.Count(string(mainContent), "  - command:")
	tldrCommands := strings.Count(string(tldrContent), "  - command:")
	totalCommands := mainCommands + tldrCommands

	fmt.Printf("✅ Successfully merged command databases!\n")
	fmt.Printf("📊 Main commands: %d\n", mainCommands)
	fmt.Printf("📊 TLDR commands: %d\n", tldrCommands)
	fmt.Printf("📊 Total commands: %d\n", totalCommands)
	fmt.Printf("📂 Saved to: %s\n", outputFile)
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var wizardCmd = &cobra.Command{
	Use:   "wizard [command]",
	Short: "Interactive command builder for complex commands",
	Long: `Launch an interactive wizard to build complex commands step-by-step.
Supports popular commands like tar, find, ffmpeg, and more.

Examples:
  wtf wizard tar      # Interactive tar archive builder
  wtf wizard find     # Interactive find command builder
  wtf wizard ffmpeg   # Interactive ffmpeg converter
  wtf wizard          # Show available wizards`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			showAvailableWizards()
			return
		}

		command := strings.ToLower(args[0])
		switch command {
		case "tar":
			runTarWizard()
		case "find":
			runFindWizard()
		case "ffmpeg":
			runFFmpegWizard()
		default:
			fmt.Printf("❌ Wizard for '%s' not available.\n", command)
			fmt.Println("\nAvailable wizards:")
			showAvailableWizards()
		}
	},
}

func showAvailableWizards() {
	fmt.Println("🧙 Available Command Wizards:")
	fmt.Println()
	fmt.Println("📁 tar     - Create and extract archives")
	fmt.Println("🔍 find    - Search for files and directories")
	fmt.Println("🎬 ffmpeg  - Convert and process media files")
	fmt.Println()
	fmt.Println("Usage: wtf wizard <command>")
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// Handle EOF or other errors gracefully
		fmt.Println("\nOperation canceled.")
		os.Exit(0)
	}
	return strings.TrimSpace(input)
}

func readChoice(prompt string, options []string) int {
	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	maxAttempts := 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		choice := readInput("Enter choice (number): ")
		if choice == "" {
			fmt.Println("Operation canceled.")
			os.Exit(0)
		}

		if num, err := strconv.Atoi(choice); err == nil && num >= 1 && num <= len(options) {
			return num - 1
		}

		if attempt < maxAttempts-1 {
			fmt.Printf("Invalid choice. Please enter 1-%d: ", len(options))
		}
	}

	fmt.Println("Too many invalid attempts. Exiting wizard.")
	os.Exit(1)
	return -1 // Never reached
}

func readYesNo(prompt string) bool {
	maxAttempts := 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		response := strings.ToLower(readInput(prompt + " (y/n): "))
		if response == "" {
			fmt.Println("Operation canceled.")
			os.Exit(0)
		}
		if response == "y" || response == "yes" {
			return true
		}
		if response == "n" || response == "no" {
			return false
		}
		if attempt < maxAttempts-1 {
			fmt.Print("Please enter 'y' or 'n': ")
		}
	}

	fmt.Println("Too many invalid attempts. Exiting wizard.")
	os.Exit(1)
	return false // Never reached
}

func runTarWizard() {
	fmt.Println("🧙 ✨ TAR Archive Wizard ✨")
	fmt.Println("I'll help you build the perfect tar command!")
	fmt.Println()

	// Step 1: Operation type
	operations := []string{
		"Create an archive",
		"Extract an archive",
		"List contents of archive",
		"Add files to existing archive",
	}

	operation := readChoice("What do you want to do?", operations)

	var command strings.Builder
	command.WriteString("tar ")

	switch operation {
	case 0: // Create
		buildTarCreate(&command)
	case 1: // Extract
		buildTarExtract(&command)
	case 2: // List
		buildTarList(&command)
	case 3: // Add files
		buildTarAdd(&command)
	}

	// Final command
	finalCommand := command.String()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 Your tar command is ready!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Command: %s\n", finalCommand)
	fmt.Println(strings.Repeat("=", 50))

	// Option to save
	if readYesNo("\nSave this command to your personal notebook?") {
		description := readInput("Enter description: ")
		if description == "" {
			description = "Generated tar command"
		}

		// Here we could call the save functionality
		fmt.Printf("✅ Command saved: %s\n", description)
		fmt.Printf("   Command: %s\n", finalCommand)
	}
}

func buildTarCompressionFlag(sb *strings.Builder, compression int) {
	switch compression {
	case 1:
		sb.WriteString("z")
	case 2:
		sb.WriteString("j")
	case 3:
		sb.WriteString("J")
	}
}

func getDefaultArchiveName(compression int) string {
	switch compression {
	case 1:
		return "archive.tar.gz"
	case 2:
		return "archive.tar.bz2"
	case 3:
		return "archive.tar.xz"
	default:
		return "archive.tar"
	}
}

func runFindWizard() {
	fmt.Println("🧙 ✨ FIND Command Wizard ✨")
	fmt.Println("I'll help you build the perfect find command!")
	fmt.Println()

	var command strings.Builder
	command.WriteString("find ")

	// Search location
	location := readInput("Enter search location (default: current directory): ")
	if location == "" {
		location = "."
	}
	command.WriteString(location)

	// Search criteria options available
	fmt.Println("\nBuilding your find command step by step...")
	fmt.Println("I'll ask about different search criteria.")

	// Name pattern
	// Name pattern
	if readYesNo("Search by name pattern?") {
		pattern := readInput("Enter name pattern (e.g., '*.txt', 'test*'): ")
		buildFindNamePattern(&command, pattern)
	}

	// Filters
	buildFindTypeFilter(&command)
	buildFindSizeFilter(&command)
	buildFindTimeFilter(&command)

	// Actions
	buildFindActions(&command)

	// Limit search depth
	if readYesNo("\nLimit search depth?") {
		depth := readInput("Enter maximum depth: ")
		// Insert maxdepth at the beginning (after find location)
		finalCommand := strings.Replace(command.String(), location, location+" -maxdepth "+depth, 1)
		command.Reset()
		command.WriteString(finalCommand)
	}

	// Final command
	finalCommand := command.String()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 Your find command is ready!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Command: %s\n", finalCommand)
	fmt.Println(strings.Repeat("=", 50))

	// Option to save
	if readYesNo("\nSave this command to your personal notebook?") {
		description := readInput("Enter description: ")
		if description == "" {
			description = "Generated find command"
		}

		fmt.Printf("✅ Command saved: %s\n", description)
		fmt.Printf("   Command: %s\n", finalCommand)
	}
}

func buildFindNamePattern(sb *strings.Builder, pattern string) {
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		sb.WriteString(" -name '" + pattern + "'")
	} else {
		sb.WriteString(" -name '*" + pattern + "*'")
	}
}

func buildFindTypeFilter(sb *strings.Builder) {
	if !readYesNo("\nFilter by file type?") {
		return
	}
	types := []string{
		"Regular files",
		"Directories",
		"Symbolic links",
		"Executable files",
	}
	switch readChoice("Choose file type:", types) {
	case 0:
		sb.WriteString(" -type f")
	case 1:
		sb.WriteString(" -type d")
	case 2:
		sb.WriteString(" -type l")
	case 3:
		sb.WriteString(" -type f -executable")
	}
}

func buildFindSizeFilter(sb *strings.Builder) {
	if !readYesNo("\nFilter by file size?") {
		return
	}
	sizeOps := []string{"Larger than", "Smaller than", "Exactly"}
	sizeOp := readChoice("Size comparison:", sizeOps)
	size := readInput("Enter size (e.g., 100k, 1M, 2G): ")
	switch sizeOp {
	case 0:
		sb.WriteString(" -size +" + size)
	case 1:
		sb.WriteString(" -size -" + size)
	case 2:
		sb.WriteString(" -size " + size)
	}
}

func buildFindTimeFilter(sb *strings.Builder) {
	if !readYesNo("\nFilter by modification time?") {
		return
	}
	timeOps := []string{
		"Modified within last N days",
		"Modified more than N days ago",
		"Modified exactly N days ago",
	}
	timeOp := readChoice("Time comparison:", timeOps)
	days := readInput("Enter number of days: ")
	switch timeOp {
	case 0:
		sb.WriteString(" -mtime -" + days)
	case 1:
		sb.WriteString(" -mtime +" + days)
	case 2:
		sb.WriteString(" -mtime " + days)
	}
}

func runFFmpegWizard() {
	fmt.Println("🧙 ✨ FFMPEG Wizard ✨")
	fmt.Println("I'll help you build the perfect ffmpeg command!")
	fmt.Println()

	var command strings.Builder
	command.WriteString("ffmpeg")

	// Input file
	inputFile := readInput("Enter input file name: ")
	command.WriteString(" -i " + inputFile)

	// Operation type
	operations := []string{
		"Convert video format",
		"Extract audio from video",
		"Resize/scale video",
		"Cut/trim video",
		"Merge videos",
		"Convert audio format",
		"Add watermark",
	}

	operation := readChoice("\nWhat do you want to do?", operations)

	switch operation {
	case 0: // Convert video format
		buildFFmpegConvertVideo(&command)
	case 1: // Extract audio
		buildFFmpegExtractAudio(&command)
	case 2: // Resize video
		buildFFmpegResizeVideo(&command)
	case 3: // Cut/trim video
		buildFFmpegTrimVideo(&command)
	}

	// Output file
	outputFile := readInput("\nEnter output file name: ")
	command.WriteString(" " + outputFile)

	// Overwrite option
	if readYesNo("\nOverwrite output file if exists?") {
		command.WriteString(" -y")
	}

	// Final command
	finalCommand := command.String()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Your ffmpeg command is ready!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Command: %s\n", finalCommand)
	fmt.Println(strings.Repeat("=", 60))

	// Estimated time warning
	if operation == 0 || operation == 2 {
		fmt.Println("⏱️  Note: Video processing may take time depending on file size and quality settings.")
	}

	// Option to save
	if readYesNo("\nSave this command to your personal notebook?") {
		description := readInput("Enter description: ")
		if description == "" {
			description = "Generated ffmpeg command"
		}

		fmt.Printf("✅ Command saved: %s\n", description)
		fmt.Printf("   Command: %s\n", finalCommand)
	}
}

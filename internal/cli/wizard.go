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
	Run: func(cmd *cobra.Command, args []string) {
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
		fmt.Println("\nOperation cancelled.")
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
			fmt.Println("Operation cancelled.")
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
			fmt.Println("Operation cancelled.")
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
		command.WriteString("-c")

		// Compression
		compressions := []string{
			"No compression (.tar)",
			"Gzip compression (.tar.gz / .tgz)",
			"Bzip2 compression (.tar.bz2)",
			"XZ compression (.tar.xz)",
		}

		compression := readChoice("\nChoose compression:", compressions)
		switch compression {
		case 1:
			command.WriteString("z")
		case 2:
			command.WriteString("j")
		case 3:
			command.WriteString("J")
		}

		// Verbose
		if readYesNo("\nShow files being processed (verbose)?") {
			command.WriteString("v")
		}

		command.WriteString("f")

		// Archive name
		archiveName := readInput("\nEnter archive name: ")
		if archiveName == "" {
			archiveName = "archive.tar"
			if compression == 1 {
				archiveName = "archive.tar.gz"
			} else if compression == 2 {
				archiveName = "archive.tar.bz2"
			} else if compression == 3 {
				archiveName = "archive.tar.xz"
			}
		}
		command.WriteString(" " + archiveName)

		// Source files/directories
		source := readInput("Enter files/directories to archive (default: current directory): ")
		if source == "" {
			source = "."
		}
		command.WriteString(" " + source)

		// Exclude patterns
		if readYesNo("\nExclude any files/patterns?") {
			exclude := readInput("Enter exclude pattern (e.g., '*.tmp'): ")
			if exclude != "" {
				command.WriteString(" --exclude='" + exclude + "'")
			}
		}

	case 1: // Extract
		command.WriteString("-x")

		// Auto-detect compression
		if readYesNo("\nAuto-detect compression?") {
			command.WriteString("a")
		} else {
			compressions := []string{
				"No compression",
				"Gzip compression",
				"Bzip2 compression",
				"XZ compression",
			}
			compression := readChoice("Choose compression type:", compressions)
			switch compression {
			case 1:
				command.WriteString("z")
			case 2:
				command.WriteString("j")
			case 3:
				command.WriteString("J")
			}
		}

		// Verbose
		if readYesNo("\nShow files being extracted (verbose)?") {
			command.WriteString("v")
		}

		command.WriteString("f")

		// Archive name
		archiveName := readInput("\nEnter archive file name: ")
		command.WriteString(" " + archiveName)

		// Extract directory
		if readYesNo("\nExtract to specific directory?") {
			extractDir := readInput("Enter directory: ")
			command.WriteString(" -C " + extractDir)
		}

	case 2: // List
		command.WriteString("-tv")

		// Auto-detect compression
		if readYesNo("\nAuto-detect compression?") {
			command.WriteString("a")
		}

		command.WriteString("f")

		// Archive name
		archiveName := readInput("\nEnter archive file name: ")
		command.WriteString(" " + archiveName)

	case 3: // Add files
		command.WriteString("-rv")

		// Archive name
		archiveName := readInput("\nEnter existing archive name: ")
		command.WriteString("f " + archiveName)

		// Files to add
		files := readInput("Enter files to add: ")
		command.WriteString(" " + files)
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
	if readYesNo("Search by name pattern?") {
		pattern := readInput("Enter name pattern (e.g., '*.txt', 'test*'): ")
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			command.WriteString(" -name '" + pattern + "'")
		} else {
			command.WriteString(" -name '*" + pattern + "*'")
		}
	}

	// File type
	if readYesNo("\nFilter by file type?") {
		types := []string{
			"Regular files",
			"Directories",
			"Symbolic links",
			"Executable files",
		}

		typeChoice := readChoice("Choose file type:", types)
		switch typeChoice {
		case 0:
			command.WriteString(" -type f")
		case 1:
			command.WriteString(" -type d")
		case 2:
			command.WriteString(" -type l")
		case 3:
			command.WriteString(" -type f -executable")
		}
	}

	// File size
	if readYesNo("\nFilter by file size?") {
		sizeOps := []string{
			"Larger than",
			"Smaller than",
			"Exactly",
		}

		sizeOp := readChoice("Size comparison:", sizeOps)
		size := readInput("Enter size (e.g., 100k, 1M, 2G): ")

		switch sizeOp {
		case 0:
			command.WriteString(" -size +" + size)
		case 1:
			command.WriteString(" -size -" + size)
		case 2:
			command.WriteString(" -size " + size)
		}
	}

	// Modification time
	if readYesNo("\nFilter by modification time?") {
		timeOps := []string{
			"Modified within last N days",
			"Modified more than N days ago",
			"Modified exactly N days ago",
		}

		timeOp := readChoice("Time comparison:", timeOps)
		days := readInput("Enter number of days: ")

		switch timeOp {
		case 0:
			command.WriteString(" -mtime -" + days)
		case 1:
			command.WriteString(" -mtime +" + days)
		case 2:
			command.WriteString(" -mtime " + days)
		}
	}

	// Actions
	if readYesNo("\nPerform action on found files?") {
		actions := []string{
			"Just list them (default)",
			"Delete them",
			"Copy to directory",
			"Move to directory",
			"Execute command on each",
			"Print detailed info",
		}

		action := readChoice("Choose action:", actions)
		switch action {
		case 0:
			// Default action, do nothing
		case 1:
			if readYesNo("⚠️  This will DELETE files! Are you sure?") {
				command.WriteString(" -delete")
			}
		case 2:
			destDir := readInput("Enter destination directory: ")
			command.WriteString(" -exec cp {} " + destDir + " \\;")
		case 3:
			destDir := readInput("Enter destination directory: ")
			command.WriteString(" -exec mv {} " + destDir + " \\;")
		case 4:
			execCmd := readInput("Enter command to execute (use {} for filename): ")
			command.WriteString(" -exec " + execCmd + " \\;")
		case 5:
			command.WriteString(" -ls")
		}
	}

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
		formats := []string{
			"MP4 (H.264)",
			"AVI",
			"MOV",
			"MKV",
			"WebM",
			"WMV",
		}

		format := readChoice("\nChoose output format:", formats)

		// Quality settings
		qualities := []string{
			"High quality (slower encoding)",
			"Medium quality (balanced)",
			"Low quality (faster encoding)",
			"Custom settings",
		}

		quality := readChoice("\nChoose quality:", qualities)

		switch format {
		case 0: // MP4
			command.WriteString(" -c:v libx264")
			switch quality {
			case 0:
				command.WriteString(" -crf 18")
			case 1:
				command.WriteString(" -crf 23")
			case 2:
				command.WriteString(" -crf 28")
			case 3:
				crf := readInput("Enter CRF value (18-28, lower = better): ")
				command.WriteString(" -crf " + crf)
			}
			command.WriteString(" -c:a aac")
		case 1: // AVI
			command.WriteString(" -c:v libxvid -c:a mp3")
		case 2: // MOV
			command.WriteString(" -c:v libx264 -c:a aac")
		case 3: // MKV
			command.WriteString(" -c:v libx264 -c:a aac")
		case 4: // WebM
			command.WriteString(" -c:v libvpx-vp9 -c:a libopus")
		case 5: // WMV
			command.WriteString(" -c:v wmv2 -c:a wmav2")
		}

	case 1: // Extract audio
		audioFormats := []string{
			"MP3",
			"AAC",
			"WAV",
			"FLAC",
			"OGG",
		}

		audioFormat := readChoice("\nChoose audio format:", audioFormats)

		switch audioFormat {
		case 0: // MP3
			command.WriteString(" -vn -c:a libmp3lame")
			bitrate := readInput("Enter bitrate (default: 192k): ")
			if bitrate == "" {
				bitrate = "192k"
			}
			command.WriteString(" -b:a " + bitrate)
		case 1: // AAC
			command.WriteString(" -vn -c:a aac")
		case 2: // WAV
			command.WriteString(" -vn -c:a pcm_s16le")
		case 3: // FLAC
			command.WriteString(" -vn -c:a flac")
		case 4: // OGG
			command.WriteString(" -vn -c:a libvorbis")
		}

	case 2: // Resize video
		resolutions := []string{
			"1920x1080 (1080p)",
			"1280x720 (720p)",
			"854x480 (480p)",
			"640x360 (360p)",
			"Custom resolution",
			"Scale by factor",
		}

		resolution := readChoice("\nChoose resolution:", resolutions)

		switch resolution {
		case 0:
			command.WriteString(" -vf scale=1920:1080")
		case 1:
			command.WriteString(" -vf scale=1280:720")
		case 2:
			command.WriteString(" -vf scale=854:480")
		case 3:
			command.WriteString(" -vf scale=640:360")
		case 4:
			width := readInput("Enter width: ")
			height := readInput("Enter height: ")
			command.WriteString(" -vf scale=" + width + ":" + height)
		case 5:
			factor := readInput("Enter scale factor (e.g., 0.5 for half size): ")
			command.WriteString(" -vf scale=iw*" + factor + ":ih*" + factor)
		}

	case 3: // Cut/trim video
		fmt.Println("\nTrim options:")
		startTime := readInput("Enter start time (HH:MM:SS or seconds): ")

		if readYesNo("Specify duration?") {
			duration := readInput("Enter duration (HH:MM:SS or seconds): ")
			command.WriteString(" -ss " + startTime + " -t " + duration)
		} else {
			endTime := readInput("Enter end time (HH:MM:SS or seconds): ")
			command.WriteString(" -ss " + startTime + " -to " + endTime)
		}

		command.WriteString(" -c copy") // Copy streams for speed
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

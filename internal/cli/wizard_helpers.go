package cli

import (
	"fmt"
	"strings"
)

func buildFFmpegConvertVideo(sb *strings.Builder) {
	formats := []string{"MP4 (H.264)", "AVI", "MOV", "MKV", "WebM", "WMV"}
	format := readChoice("\nChoose output format:", formats)
	qualities := []string{"High quality (slower encoding)", "Medium quality (balanced)", "Low quality (faster encoding)", "Custom settings"}
	quality := readChoice("\nChoose quality:", qualities)

	switch format {
	case 0: // MP4
		sb.WriteString(" -c:v libx264")
		switch quality {
		case 0:
			sb.WriteString(" -crf 18")
		case 1:
			sb.WriteString(" -crf 23")
		case 2:
			sb.WriteString(" -crf 28")
		case 3:
			crf := readInput("Enter CRF value (18-28, lower = better): ")
			sb.WriteString(" -crf " + crf)
		}
		sb.WriteString(" -c:a aac")
	case 1: // AVI
		sb.WriteString(" -c:v libxvid -c:a mp3")
	case 2: // MOV
		sb.WriteString(" -c:v libx264 -c:a aac")
	case 3: // MKV
		sb.WriteString(" -c:v libx264 -c:a aac")
	case 4: // WebM
		sb.WriteString(" -c:v libvpx-vp9 -c:a libopus")
	case 5: // WMV
		sb.WriteString(" -c:v wmv2 -c:a wmav2")
	}
}

func buildFFmpegExtractAudio(sb *strings.Builder) {
	audioFormats := []string{"MP3", "AAC", "WAV", "FLAC", "OGG"}
	switch readChoice("\nChoose audio format:", audioFormats) {
	case 0: // MP3
		sb.WriteString(" -vn -c:a libmp3lame")
		bitrate := readInput("Enter bitrate (default: 192k): ")
		if bitrate == "" {
			bitrate = "192k"
		}
		sb.WriteString(" -b:a " + bitrate)
	case 1: // AAC
		sb.WriteString(" -vn -c:a aac")
	case 2: // WAV
		sb.WriteString(" -vn -c:a pcm_s16le")
	case 3: // FLAC
		sb.WriteString(" -vn -c:a flac")
	case 4: // OGG
		sb.WriteString(" -vn -c:a libvorbis")
	}
}

func buildFFmpegResizeVideo(sb *strings.Builder) {
	resolutions := []string{"1920x1080 (1080p)", "1280x720 (720p)", "854x480 (480p)", "640x360 (360p)", "Custom resolution", "Scale by factor"}
	switch readChoice("\nChoose resolution:", resolutions) {
	case 0:
		sb.WriteString(" -vf scale=1920:1080")
	case 1:
		sb.WriteString(" -vf scale=1280:720")
	case 2:
		sb.WriteString(" -vf scale=854:480")
	case 3:
		sb.WriteString(" -vf scale=640:360")
	case 4:
		width := readInput("Enter width: ")
		height := readInput("Enter height: ")
		sb.WriteString(" -vf scale=" + width + ":" + height)
	case 5:
		factor := readInput("Enter scale factor (e.g., 0.5 for half size): ")
		sb.WriteString(" -vf scale=iw*" + factor + ":ih*" + factor)
	}
}

func buildFFmpegTrimVideo(sb *strings.Builder) {
	fmt.Println("\nTrim options:")
	startTime := readInput("Enter start time (HH:MM:SS or seconds): ")
	if readYesNo("Specify duration?") {
		duration := readInput("Enter duration (HH:MM:SS or seconds): ")
		sb.WriteString(" -ss " + startTime + " -t " + duration)
	} else {
		endTime := readInput("Enter end time (HH:MM:SS or seconds): ")
		sb.WriteString(" -ss " + startTime + " -to " + endTime)
	}
	sb.WriteString(" -c copy")
}

func buildTarCreate(sb *strings.Builder) {
	sb.WriteString("-c")

	// Compression
	compressions := []string{
		"No compression (.tar)",
		"Gzip compression (.tar.gz / .tgz)",
		"Bzip2 compression (.tar.bz2)",
		"XZ compression (.tar.xz)",
	}

	compression := readChoice("\nChoose compression:", compressions)
	buildTarCompressionFlag(sb, compression)

	// Verbose
	if readYesNo("\nShow files being processed (verbose)?") {
		sb.WriteString("v")
	}

	sb.WriteString("f")

	// Archive name
	archiveName := readInput("\nEnter archive name: ")
	if archiveName == "" {
		archiveName = getDefaultArchiveName(compression)
	}
	sb.WriteString(" " + archiveName)

	// Source files/directories
	source := readInput("Enter files/directories to archive (default: current directory): ")
	if source == "" {
		source = "."
	}
	sb.WriteString(" " + source)

	// Exclude patterns
	if readYesNo("\nExclude any files/patterns?") {
		exclude := readInput("Enter exclude pattern (e.g., '*.tmp'): ")
		if exclude != "" {
			sb.WriteString(" --exclude='" + exclude + "'")
		}
	}
}

func buildTarExtract(sb *strings.Builder) {
	sb.WriteString("-x")

	// Auto-detect compression
	if readYesNo("\nAuto-detect compression?") {
		sb.WriteString("a")
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
			sb.WriteString("z")
		case 2:
			sb.WriteString("j")
		case 3:
			sb.WriteString("J")
		}
	}

	// Verbose
	if readYesNo("\nShow files being extracted (verbose)?") {
		sb.WriteString("v")
	}

	sb.WriteString("f")

	// Archive name
	archiveName := readInput("\nEnter archive file name: ")
	sb.WriteString(" " + archiveName)

	// Extract directory
	if readYesNo("\nExtract to specific directory?") {
		extractDir := readInput("Enter directory: ")
		sb.WriteString(" -C " + extractDir)
	}
}

func buildTarList(sb *strings.Builder) {
	sb.WriteString("-tv")

	// Auto-detect compression
	if readYesNo("\nAuto-detect compression?") {
		sb.WriteString("a")
	}

	sb.WriteString("f")

	// Archive name
	archiveName := readInput("\nEnter archive file name: ")
	sb.WriteString(" " + archiveName)
}

func buildTarAdd(sb *strings.Builder) {
	sb.WriteString("-rv")

	// Archive name
	archiveName := readInput("\nEnter existing archive name: ")
	sb.WriteString("f " + archiveName)

	// Files to add
	files := readInput("Enter files to add: ")
	sb.WriteString(" " + files)
}

func buildFindActions(sb *strings.Builder) {
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
			if readYesNo("WARNING: This will DELETE files! Are you sure?") {
				sb.WriteString(" -delete")
			}
		case 2:
			destDir := readInput("Enter destination directory: ")
			sb.WriteString(" -exec cp {} " + destDir + " \\;")
		case 3:
			destDir := readInput("Enter destination directory: ")
			sb.WriteString(" -exec mv {} " + destDir + " \\;")
		case 4:
			execCmd := readInput("Enter command to execute (use {} for filename): ")
			sb.WriteString(" -exec " + execCmd + " \\;")
		case 5:
			sb.WriteString(" -ls")
		}
	}
}

package cli

import (
	"os"
	"runtime"
)

func isTerminalOutput() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func shouldUseColor(noColor bool) bool {
	if noColor {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return isTerminalOutput()
}

func shouldUseUnicodeSymbols() bool {
	if !isTerminalOutput() {
		return false
	}
	if runtime.GOOS == osWindows {
		return os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") != "" || os.Getenv("ConEmuANSI") == "ON"
	}
	return true
}

func truncateRunes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

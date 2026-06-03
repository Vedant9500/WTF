package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var aliasNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

func validateAliasName(name string) error {
	if !aliasNamePattern.MatchString(name) {
		return fmt.Errorf("invalid alias name %q: use 1-64 characters, start with a letter, and use only letters, numbers, '_' or '-'", name)
	}
	return nil
}

func validatePathWithinDir(baseDir, targetPath string) error {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to write alias outside %s", baseAbs)
	}
	return nil
}

func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
}

func shellQuotePattern(pattern string) string {
	return shellQuote(pattern)
}

func sanitizeUnsignedInt(input string) (string, bool) {
	if input == "" {
		return "", false
	}
	for _, r := range input {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return input, true
}

func sanitizeFindSize(input string) (string, bool) {
	if input == "" {
		return "", false
	}
	hasDigit := false
	for i, r := range input {
		if r >= '0' && r <= '9' {
			hasDigit = true
			continue
		}
		if i == len(input)-1 {
			switch r {
			case 'b', 'c', 'w', 'k', 'M', 'G':
				return input, hasDigit
			}
		}
		return "", false
	}
	return input, hasDigit
}

func sanitizeTimeValue(input string) (string, bool) {
	if input == "" {
		return "", false
	}
	for _, r := range input {
		if (r >= '0' && r <= '9') || r == ':' || r == '.' {
			continue
		}
		return "", false
	}
	return input, true
}

func isSafeFindExecCommand(input string) bool {
	if strings.TrimSpace(input) == "" {
		return false
	}
	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case strings.ContainsRune(" -_./{}:+,=@%", r):
		default:
			return false
		}
	}
	return true
}

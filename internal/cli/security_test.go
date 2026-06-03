package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Vedant9500/WTF/internal/database"
)

func TestValidateAliasNameRejectsUnsafeNames(t *testing.T) {
	invalid := []string{
		"../evil",
		"bad/name",
		"bad\\name",
		"bad;touch-x",
		"bad\nalias",
		"1starts-with-digit",
		"",
	}

	for _, name := range invalid {
		if err := validateAliasName(name); err == nil {
			t.Fatalf("expected alias %q to be rejected", name)
		}
	}

	valid := []string{"hey", "cmd-help", "miko_2"}
	for _, name := range valid {
		if err := validateAliasName(name); err != nil {
			t.Fatalf("expected alias %q to be valid: %v", name, err)
		}
	}
}

func TestValidatePathWithinDirRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	inside := filepath.Join(base, "safe")
	if err := validatePathWithinDir(base, inside); err != nil {
		t.Fatalf("expected inside path to be valid: %v", err)
	}

	outside := filepath.Join(base, "..", "outside")
	if err := validatePathWithinDir(base, outside); err == nil {
		t.Fatalf("expected outside path to be rejected")
	}
}

func TestShellQuoteEscapesSingleQuotes(t *testing.T) {
	got := shellQuote("a'b c")
	want := `'a'\''b c'`
	if got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}

func TestWizardBuildersQuoteShellArguments(t *testing.T) {
	var sb strings.Builder
	buildFindNamePattern(&sb, "Bob's *.txt")
	if !strings.Contains(sb.String(), `'Bob'\''s *.txt'`) {
		t.Fatalf("expected quoted pattern, got %q", sb.String())
	}

	var ff strings.Builder
	ff.WriteString("ffmpeg")
	input := "in file's.mp4"
	ff.WriteString(" -i " + shellQuote(input))
	if strings.Contains(ff.String(), " -i "+input) {
		t.Fatalf("expected quoted ffmpeg input, got %q", ff.String())
	}
}

func TestSetupUnixAliasLineQuotesExecutable(t *testing.T) {
	line := "alias hey=" + shellQuote("/tmp/wtf's/bin/wtf")
	if strings.Contains(line, "wtf's") {
		t.Fatalf("expected executable quote to be escaped, got %q", line)
	}
}

func TestSavePersonalDatabaseUsesPrivatePermissions(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("Windows does not reliably expose POSIX permission bits")
	}

	dbPath := filepath.Join(t.TempDir(), "nested", "personal.yml")
	entry := database.Command{Command: "echo hi", Description: "say hi", Keywords: []string{"echo"}}
	if err := saveToPersonalDatabase(dbPath, entry); err != nil {
		t.Fatalf("saveToPersonalDatabase failed: %v", err)
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if got := info.Mode().Perm(); got != personalDBFileMode {
		t.Fatalf("file mode = %v, want %v", got, os.FileMode(personalDBFileMode))
	}
}

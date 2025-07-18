package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semantic versioning pattern
	if len(Version) < 5 { // minimum: "1.0.0"
		t.Errorf("Version '%s' seems too short for semantic versioning", Version)
	}
}

func TestBuildInfo(t *testing.T) {
	info := BuildInfo()

	if info == "" {
		t.Error("BuildInfo should not be empty")
	}

	// Should contain version information
	if !containsSubstring(info, Version) {
		t.Errorf("BuildInfo should contain version '%s'", Version)
	}

	if !containsSubstring(info, "WTF") {
		t.Error("BuildInfo should contain 'WTF'")
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

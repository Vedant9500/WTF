package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer_AnalyzeDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	analyzer := NewAnalyzer()

	tests := []struct {
		name           string
		files          []string
		expectedTypes  []ProjectType
		expectedLang   string
		expectedGit    bool
		expectedDocker bool
	}{
		{
			name:          "Git repository",
			files:         []string{".git/"},
			expectedTypes: []ProjectType{ProjectTypeGit},
			expectedGit:   true,
		},
		{
			name:           "Docker project",
			files:          []string{"Dockerfile"},
			expectedTypes:  []ProjectType{ProjectTypeDocker},
			expectedDocker: true,
		},
		{
			name:          "Node.js project",
			files:         []string{"package.json"},
			expectedTypes: []ProjectType{ProjectTypeNode},
			expectedLang:  "javascript",
		},
		{
			name:          "Python project",
			files:         []string{"requirements.txt"},
			expectedTypes: []ProjectType{ProjectTypePython},
			expectedLang:  "python",
		},
		{
			name:          "Go project",
			files:         []string{"go.mod"},
			expectedTypes: []ProjectType{ProjectTypeGo},
			expectedLang:  "go",
		},
		{
			name:           "Multi-type project",
			files:          []string{".git/", "Dockerfile", "package.json"},
			expectedTypes:  []ProjectType{ProjectTypeGit, ProjectTypeDocker, ProjectTypeNode},
			expectedLang:   "javascript",
			expectedGit:    true,
			expectedDocker: true,
		},
		{
			name:          "Empty directory",
			files:         []string{},
			expectedTypes: []ProjectType{ProjectTypeGeneric},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create test files
			for _, file := range tt.files {
				filePath := filepath.Join(testDir, file)
				if file[len(file)-1] == '/' {
					// It's a directory
					err = os.MkdirAll(filePath, 0755)
				} else {
					// It's a file
					dir := filepath.Dir(filePath)
					err = os.MkdirAll(dir, 0755)
					if err != nil {
						t.Fatalf("Failed to create directory %s: %v", dir, err)
					}
					file, err := os.Create(filePath)
					if err == nil {
						file.Close()
					}
				}
				if err != nil {
					t.Fatalf("Failed to create test file/dir %s: %v", file, err)
				}
			}

			// Analyze the directory
			ctx, err := analyzer.AnalyzeDirectory(testDir)
			if err != nil {
				t.Fatalf("AnalyzeDirectory failed: %v", err)
			}

			// Verify results
			if len(ctx.ProjectTypes) != len(tt.expectedTypes) {
				t.Errorf("Expected %d project types, got %d", len(tt.expectedTypes), len(ctx.ProjectTypes))
			}

			for _, expectedType := range tt.expectedTypes {
				found := false
				for _, actualType := range ctx.ProjectTypes {
					if actualType == expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected project type %s not found", expectedType)
				}
			}

			if ctx.Language != tt.expectedLang {
				t.Errorf("Expected language %s, got %s", tt.expectedLang, ctx.Language)
			}

			if ctx.HasGit != tt.expectedGit {
				t.Errorf("Expected HasGit %v, got %v", tt.expectedGit, ctx.HasGit)
			}

			if ctx.HasDocker != tt.expectedDocker {
				t.Errorf("Expected HasDocker %v, got %v", tt.expectedDocker, ctx.HasDocker)
			}
		})
	}
}

func TestContext_GetContextBoosts(t *testing.T) {
	tests := []struct {
		name          string
		projectTypes  []ProjectType
		expectedBoost map[string]float64
	}{
		{
			name:         "Git project",
			projectTypes: []ProjectType{ProjectTypeGit},
			expectedBoost: map[string]float64{
				"git":      2.0,
				"commit":   1.5,
				"branch":   1.5,
				"checkout": 1.5,
			},
		},
		{
			name:         "Docker project",
			projectTypes: []ProjectType{ProjectTypeDocker},
			expectedBoost: map[string]float64{
				"docker":    2.0,
				"container": 1.8,
				"image":     1.5,
			},
		},
		{
			name:         "Multi-type project",
			projectTypes: []ProjectType{ProjectTypeGit, ProjectTypeNode},
			expectedBoost: map[string]float64{
				"git":  2.0,
				"npm":  2.0,
				"node": 1.8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				ProjectTypes: tt.projectTypes,
			}

			boosts := ctx.GetContextBoosts()

			for keyword, expectedBoost := range tt.expectedBoost {
				if actualBoost, exists := boosts[keyword]; !exists {
					t.Errorf("Expected boost for keyword %s not found", keyword)
				} else if actualBoost != expectedBoost {
					t.Errorf("Expected boost %f for keyword %s, got %f", expectedBoost, keyword, actualBoost)
				}
			}
		})
	}
}

func TestContext_GetContextDescription(t *testing.T) {
	tests := []struct {
		name         string
		projectTypes []ProjectType
		expected     string
	}{
		{
			name:         "Git repository",
			projectTypes: []ProjectType{ProjectTypeGit},
			expected:     "Git repository",
		},
		{
			name:         "Docker project",
			projectTypes: []ProjectType{ProjectTypeDocker},
			expected:     "Docker project",
		},
		{
			name:         "Multi-type project",
			projectTypes: []ProjectType{ProjectTypeGit, ProjectTypeNode},
			expected:     "Git repository, Node.js project",
		},
		{
			name:         "Generic directory",
			projectTypes: []ProjectType{ProjectTypeGeneric},
			expected:     "generic directory",
		},
		{
			name:         "Empty project types",
			projectTypes: []ProjectType{},
			expected:     "generic directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				ProjectTypes: tt.projectTypes,
			}

			description := ctx.GetContextDescription()
			if description != tt.expected {
				t.Errorf("Expected description %s, got %s", tt.expected, description)
			}
		})
	}
}

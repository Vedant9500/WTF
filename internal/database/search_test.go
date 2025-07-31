package database

import (
	"testing"
)

func TestSearch(t *testing.T) {
	// Create test database
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit -m 'message'",
				Description:      "commit changes with message",
				Keywords:         []string{"git", "commit", "message"},
				CommandLower:     "git commit -m 'message'",
				DescriptionLower: "commit changes with message",
				KeywordsLower:    []string{"git", "commit", "message"},
			},
			{
				Command:          "find . -name '*.txt'",
				Description:      "find text files",
				Keywords:         []string{"find", "files", "text"},
				CommandLower:     "find . -name '*.txt'",
				DescriptionLower: "find text files",
				KeywordsLower:    []string{"find", "files", "text"},
			},
			{
				Command:          "tar -czf archive.tar.gz .",
				Description:      "create compressed archive",
				Keywords:         []string{"tar", "compress", "archive"},
				CommandLower:     "tar -czf archive.tar.gz .",
				DescriptionLower: "create compressed archive",
				KeywordsLower:    []string{"tar", "compress", "archive"},
			},
		},
	}

	// Test search functionality
	results := db.Search("git commit", 5)

	if len(results) == 0 {
		t.Error("Expected at least one result for 'git commit'")
	}

	// First result should be the git command
	if results[0].Command.Command != "git commit -m 'message'" {
		t.Errorf("Expected git command first, got '%s'", results[0].Command.Command)
	}

	// Test that score is reasonable
	if results[0].Score <= 0 {
		t.Errorf("Expected positive score, got %f", results[0].Score)
	}
}

func TestSearchLimit(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "cmd1",
				Description:      "test",
				Keywords:         []string{"test"},
				CommandLower:     "cmd1",
				DescriptionLower: "test",
				KeywordsLower:    []string{"test"},
			},
			{
				Command:          "cmd2",
				Description:      "test",
				Keywords:         []string{"test"},
				CommandLower:     "cmd2",
				DescriptionLower: "test",
				KeywordsLower:    []string{"test"},
			},
			{
				Command:          "cmd3",
				Description:      "test",
				Keywords:         []string{"test"},
				CommandLower:     "cmd3",
				DescriptionLower: "test",
				KeywordsLower:    []string{"test"},
			},
		},
	}

	results := db.Search("test", 2)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit",
				Description:      "commit",
				Keywords:         []string{"git"},
				CommandLower:     "git commit",
				DescriptionLower: "commit",
				KeywordsLower:    []string{"git"},
			},
		},
	}

	results := db.Search("nonexistent", 5)

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestCalculateScore(t *testing.T) {
	cmd := &Command{
		Command:     "git commit",
		Description: "commit changes",
		Keywords:    []string{"git", "version-control"},
		// Populate cached lowercased fields
		CommandLower:     "git commit",
		DescriptionLower: "commit changes",
		KeywordsLower:    []string{"git", "version-control"},
	}

	queryWords := []string{"git", "commit"}
	score := calculateScore(cmd, queryWords, nil) // No context boosts for basic test

	if score <= 0 {
		t.Errorf("Expected positive score, got %f", score)
	}

	// Based on actual scoring algorithm:
	// "git": matches in command (10.0) + matches in keywords (4.0) = 14.0
	// "commit": matches in command (10.0) = 10.0
	// Total should be 24.0
	expectedScore := 24.0
	if score < expectedScore {
		t.Errorf("Expected score >= %f, got %f", expectedScore, score)
	}
}

func TestCalculateScoreWithContext(t *testing.T) {
	cmd := &Command{
		Command:     "git commit -m 'message'",
		Description: "commit changes",
		Keywords:    []string{"git", "version-control"},
		// Populate cached lowercased fields
		CommandLower:     "git commit -m 'message'",
		DescriptionLower: "commit changes",
		KeywordsLower:    []string{"git", "version-control"},
	}

	queryWords := []string{"git", "commit"}

	// Test without context boosts
	scoreWithoutContext := calculateScore(cmd, queryWords, nil)

	// Test with context boosts (simulating Git repository)
	contextBoosts := map[string]float64{
		"git":    2.0,
		"commit": 1.5,
	}
	scoreWithContext := calculateScore(cmd, queryWords, contextBoosts)

	if scoreWithContext <= scoreWithoutContext {
		t.Errorf("Expected context boost to increase score. Without: %f, With: %f",
			scoreWithoutContext, scoreWithContext)
	}

	// Based on actual scoring algorithm:
	// "git": matches in command (10.0) + matches in keywords (4.0) = 14.0 * 2.0 = 28.0
	// "commit": matches in command (10.0) = 10.0 * 1.5 = 15.0
	// Total should be 43.0
	expectedMinScore := 43.0
	if scoreWithContext < expectedMinScore {
		t.Errorf("Expected context-boosted score >= %f, got %f", expectedMinScore, scoreWithContext)
	}
}

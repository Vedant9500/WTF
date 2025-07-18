package database

import (
	"testing"
)

func TestSearch(t *testing.T) {
	// Create test database
	db := &Database{
		Commands: []Command{
			{
				Command:     "git commit -m 'message'",
				Description: "commit changes with message",
				Keywords:    []string{"git", "commit", "message"},
			},
			{
				Command:     "find . -name '*.txt'",
				Description: "find text files",
				Keywords:    []string{"find", "files", "text"},
			},
			{
				Command:     "tar -czf archive.tar.gz .",
				Description: "create compressed archive",
				Keywords:    []string{"tar", "compress", "archive"},
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
			{Command: "cmd1", Description: "test", Keywords: []string{"test"}},
			{Command: "cmd2", Description: "test", Keywords: []string{"test"}},
			{Command: "cmd3", Description: "test", Keywords: []string{"test"}},
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
			{Command: "git commit", Description: "commit", Keywords: []string{"git"}},
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
	}

	queryWords := []string{"git", "commit"}
	score := calculateScore(cmd, queryWords, nil) // No context boosts for basic test

	if score <= 0 {
		t.Errorf("Expected positive score, got %f", score)
	}

	// Git should match in command (10) and keywords (3) = at least 13
	// Commit should match in command (10) = at least 10
	// Total should be at least 23
	if score < 23 {
		t.Errorf("Expected score >= 23, got %f", score)
	}
}

func TestCalculateScoreWithContext(t *testing.T) {
	cmd := &Command{
		Command:     "git commit -m 'message'",
		Description: "commit changes",
		Keywords:    []string{"git", "version-control"},
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

	// The git keyword should get a 2.0x boost and commit should get 1.5x boost
	// Original git score: 13 (10 from command + 3 from keyword)
	// Boosted git score: 13 * 2.0 = 26
	// Original commit score: 10 (from command)
	// Boosted commit score: 10 * 1.5 = 15
	// Total should be around 41
	expectedMinScore := 40.0
	if scoreWithContext < expectedMinScore {
		t.Errorf("Expected context-boosted score >= %f, got %f", expectedMinScore, scoreWithContext)
	}
}

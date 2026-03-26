package database

import (
	"strings"
	"testing"
)

func buildPlatformStatusDB() *Database {
	return &Database{
		Commands: []Command{
			{
				Command:          "systemctl status nginx",
				Description:      "linux service status",
				Keywords:         []string{"status", "service", "linux"},
				CommandLower:     "systemctl status nginx",
				DescriptionLower: "linux service status",
				KeywordsLower:    []string{"status", "service", "linux"},
				Platform:         []string{"linux"},
			},
			{
				Command:          "Get-Service -Name nginx",
				Description:      "windows service status",
				Keywords:         []string{"status", "service", "windows"},
				CommandLower:     "get-service -name nginx",
				DescriptionLower: "windows service status",
				KeywordsLower:    []string{"status", "service", "windows"},
				Platform:         []string{"windows"},
			},
			{
				Command:          "git status",
				Description:      "show repository status",
				Keywords:         []string{"status", "git"},
				CommandLower:     "git status",
				DescriptionLower: "show repository status",
				KeywordsLower:    []string{"status", "git"},
				Platform:         []string{"cross-platform"},
			},
		},
	}
}

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

func TestSearchUniversalBasic(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit -m 'msg'",
				Description:      "commit changes",
				Keywords:         []string{"git", "commit"},
				CommandLower:     "git commit -m 'msg'",
				DescriptionLower: "commit changes",
				KeywordsLower:    []string{"git", "commit"},
			},
			{
				Command:          "tar -czf a.tgz .",
				Description:      "create archive",
				Keywords:         []string{"tar", "archive", "compress"},
				CommandLower:     "tar -czf a.tgz .",
				DescriptionLower: "create archive",
				KeywordsLower:    []string{"tar", "archive", "compress"},
			},
		},
	}
	db.BuildUniversalIndex()
	res := db.SearchUniversal("git commit", SearchOptions{Limit: 5})
	if len(res) == 0 {
		t.Fatalf("expected results")
	}
	if res[0].Command == nil || !strings.Contains(res[0].Command.Command, "git commit") {
		t.Fatalf("expected git commit first, got %v", res[0].Command.Command)
	}
}

func TestSearchUniversalContextBoost(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "find . -name '*.txt'",
				Description:      "find text files",
				Keywords:         []string{"find", "files", "text"},
				CommandLower:     "find . -name '*.txt'",
				DescriptionLower: "find text files",
				KeywordsLower:    []string{"find", "files", "text"},
			},
			{
				Command:          "grep -R pattern .",
				Description:      "search recursively",
				Keywords:         []string{"grep", "search"},
				CommandLower:     "grep -R pattern .",
				DescriptionLower: "search recursively",
				KeywordsLower:    []string{"grep", "search"},
			},
		},
	}
	db.BuildUniversalIndex()
	opts := SearchOptions{Limit: 5, ContextBoosts: map[string]float64{"grep": 3}}
	res := db.SearchUniversal("search files", opts)
	if len(res) == 0 {
		t.Fatalf("expected results")
	}
	// Grep should rank higher due to context boost on 'grep'
	if res[0].Command == nil || !strings.Contains(res[0].Command.Command, "grep") {
		t.Fatalf("expected grep first due to context boost, got %v", res[0].Command.Command)
	}
}

func TestSearchUniversalExplicitPlatformsIncludeCrossByDefault(t *testing.T) {
	db := buildPlatformStatusDB()

	db.BuildUniversalIndex()
	results := db.SearchUniversal("status", SearchOptions{Limit: 10, Platforms: []string{"windows"}})

	if len(results) == 0 {
		t.Fatalf("expected results for explicit windows platform")
	}

	seenWindows := false
	seenCross := false
	for _, result := range results {
		switch result.Command.Command {
		case "Get-Service -Name nginx":
			seenWindows = true
		case "git status":
			seenCross = true
		case "systemctl status nginx":
			t.Fatalf("linux-only command should not be returned for windows platform")
		}
	}

	if !seenWindows {
		t.Fatalf("expected windows command in results")
	}
	if !seenCross {
		t.Fatalf("expected cross-platform command in results by default")
	}
}

func TestSearchUniversalNoCrossPlatformExcludesCrossPlatformResults(t *testing.T) {
	db := buildPlatformStatusDB()

	db.BuildUniversalIndex()
	results := db.SearchUniversal("status", SearchOptions{
		Limit:           10,
		Platforms:       []string{"windows"},
		NoCrossPlatform: true,
	})

	if len(results) == 0 {
		t.Fatalf("expected results for explicit windows platform")
	}

	seenWindows := false
	for _, result := range results {
		switch result.Command.Command {
		case "Get-Service -Name nginx":
			seenWindows = true
		case "git status":
			t.Fatalf("cross-platform command should be excluded when no-cross-platform is true")
		case "systemctl status nginx":
			t.Fatalf("linux-only command should not be returned for windows platform")
		}
	}

	if !seenWindows {
		t.Fatalf("expected windows command in results")
	}
}

func TestSearchUniversalRevertIntentPrioritizesGitRevert(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit",
				Description:      "commit changes to repository",
				Keywords:         []string{"git", "commit"},
				CommandLower:     "git commit",
				DescriptionLower: "commit changes to repository",
				KeywordsLower:    []string{"git", "commit"},
			},
			{
				Command:          "git revert",
				Description:      "create new commit that reverses earlier commit",
				Keywords:         []string{"git", "revert", "undo", "commit"},
				CommandLower:     "git revert",
				DescriptionLower: "create new commit that reverses earlier commit",
				KeywordsLower:    []string{"git", "revert", "undo", "commit"},
			},
			{
				Command:          "git commit-tree",
				Description:      "create commit objects",
				Keywords:         []string{"git", "commit", "tree"},
				CommandLower:     "git commit-tree",
				DescriptionLower: "create commit objects",
				KeywordsLower:    []string{"git", "commit", "tree"},
			},
		},
	}

	db.BuildUniversalIndex()
	db.buildTFIDFSearcher()

	results := db.SearchUniversal("what is the command to revert git commit", SearchOptions{Limit: 5, UseNLP: true})
	if len(results) == 0 {
		t.Fatalf("expected results")
	}

	maxRank := 3
	found := false
	for i := 0; i < len(results) && i < maxRank; i++ {
		if results[i].Command != nil && results[i].Command.Command == "git revert" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected git revert in top %d results, got top results: %q, %q, %q",
			maxRank,
			results[0].Command.Command,
			results[utilsMin(1, len(results)-1)].Command.Command,
			results[utilsMin(2, len(results)-1)].Command.Command,
		)
	}
}

func utilsMin(a, b int) int {
	if a < b {
		return a
	}
	return b
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

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

func TestSearchUniversalBigramTermMatchesCompoundCommand(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git reset --hard",
				Description:      "reset repository state",
				Keywords:         []string{"git", "reset"},
				CommandLower:     "git reset --hard",
				DescriptionLower: "reset repository state",
				KeywordsLower:    []string{"git", "reset"},
			},
			{
				Command:          "reset git index",
				Description:      "reset git index state",
				Keywords:         []string{"reset", "git", "index"},
				CommandLower:     "reset git index",
				DescriptionLower: "reset git index state",
				KeywordsLower:    []string{"reset", "git", "index"},
			},
		},
	}

	db.BuildUniversalIndex()
	scores := db.calculateInitialScores([]string{"git_reset"}, nil, SearchOptions{Limit: 5})

	if len(scores) != 1 {
		t.Fatalf("expected exactly one bigram match, got %d", len(scores))
	}

	if _, ok := scores[0]; !ok {
		t.Fatalf("expected compound command to match bigram term")
	}
}

func TestSearchUniversalCharNGramRecoversTypos(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit -m 'msg'",
				Description:      "commit repository changes",
				Keywords:         []string{"git", "commit"},
				CommandLower:     "git commit -m 'msg'",
				DescriptionLower: "commit repository changes",
				KeywordsLower:    []string{"git", "commit"},
			},
			{
				Command:          "docker build .",
				Description:      "build container image",
				Keywords:         []string{"docker", "build"},
				CommandLower:     "docker build .",
				DescriptionLower: "build container image",
				KeywordsLower:    []string{"docker", "build"},
			},
		},
	}

	db.BuildUniversalIndex()
	results := db.SearchUniversal("gti comit", SearchOptions{
		Limit:          5,
		UseNLP:         false,
		UseFuzzy:       false,
		DisableBigrams: true,
	})

	if len(results) == 0 {
		t.Fatalf("expected typo query to recover candidates via char n-grams")
	}

	if results[0].Command == nil || !strings.Contains(results[0].Command.Command, "git commit") {
		t.Fatalf("expected git commit command first, got %v", results[0].Command.Command)
	}
}

func TestSearchUniversalBM25OverrideMinIDFCanSuppressLowSignalTerms(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "git commit -m 'msg'",
				Description:      "commit repository changes",
				Keywords:         []string{"git", "commit"},
				CommandLower:     "git commit -m 'msg'",
				DescriptionLower: "commit repository changes",
				KeywordsLower:    []string{"git", "commit"},
			},
			{
				Command:          "git status",
				Description:      "show repository status",
				Keywords:         []string{"git", "status"},
				CommandLower:     "git status",
				DescriptionLower: "show repository status",
				KeywordsLower:    []string{"git", "status"},
			},
		},
	}

	db.BuildUniversalIndex()

	highMinIDF := 10.0
	results := db.SearchUniversal("git", SearchOptions{
		Limit:            5,
		UseFuzzy:         false,
		DisableCharNGram: true,
		BM25Overrides: &BM25Overrides{
			MinIDF: &highMinIDF,
		},
	})

	if len(results) != 0 {
		t.Fatalf("expected no results when minIDF override is very high, got %d", len(results))
	}
}

func TestDescriptionProximityBoostPrefersCloserTerms(t *testing.T) {
	closeCmd := &Command{DescriptionLower: "show disk usage by directory quickly"}
	farCmd := &Command{DescriptionLower: "show disk metrics and stats and health and diagnostics by directory"}
	terms := []string{"disk", "directory"}

	closeBoost := calculateDescriptionProximityBoost(closeCmd, terms)
	farBoost := calculateDescriptionProximityBoost(farCmd, terms)

	if closeBoost <= farBoost {
		t.Fatalf("expected close proximity boost > far proximity boost, close=%.4f far=%.4f", closeBoost, farBoost)
	}
	if closeBoost <= 1.0 {
		t.Fatalf("expected proximity boost above neutral for close terms, got %.4f", closeBoost)
	}
}

func TestSearchUniversalDisableProximityChangesScore(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "cmd-near",
				Description:      "show disk usage by directory quickly",
				Keywords:         []string{"disk", "directory"},
				CommandLower:     "cmd-near",
				DescriptionLower: "show disk usage by directory quickly",
				KeywordsLower:    []string{"disk", "directory"},
			},
		},
	}

	db.BuildUniversalIndex()
	query := "disk directory"

	withProximity := db.SearchUniversal(query, SearchOptions{
		Limit:            5,
		UseNLP:           false,
		UseFuzzy:         false,
		DisableBigrams:   true,
		DisableCharNGram: true,
	})
	withoutProximity := db.SearchUniversal(query, SearchOptions{
		Limit:            5,
		UseNLP:           false,
		UseFuzzy:         false,
		DisableBigrams:   true,
		DisableCharNGram: true,
		DisableProximity: true,
	})

	if len(withProximity) == 0 || len(withoutProximity) == 0 {
		t.Fatalf("expected results for both proximity modes")
	}
	if withProximity[0].Command == nil || withProximity[0].Command.Command != "cmd-near" {
		t.Fatalf("expected cmd-near first with proximity enabled, got %v", withProximity[0].Command.Command)
	}
	if withoutProximity[0].Command == nil || withoutProximity[0].Command.Command != "cmd-near" {
		t.Fatalf("expected cmd-near first with proximity disabled, got %v", withoutProximity[0].Command.Command)
	}

	if withProximity[0].Score <= withoutProximity[0].Score {
		t.Fatalf("expected proximity-enabled score > disabled score, with=%.4f without=%.4f",
			withProximity[0].Score, withoutProximity[0].Score)
	}
}

func TestMergeFamilyExpansionCandidatesAddsBaseRelatedCommands(t *testing.T) {
	db := &Database{
		Commands: []Command{
			{
				Command:          "tar -czf archive.tar.gz .",
				Description:      "create compressed tar archive",
				Keywords:         []string{"compress", "archive", "tar", "backup"},
				CommandLower:     "tar -czf archive.tar.gz .",
				DescriptionLower: "create compressed tar archive",
				KeywordsLower:    []string{"compress", "archive", "tar", "backup"},
			},
			{
				Command:          "zip -r backup.zip .",
				Description:      "create zip archive",
				Keywords:         []string{"compress", "archive", "zip", "backup"},
				CommandLower:     "zip -r backup.zip .",
				DescriptionLower: "create zip archive",
				KeywordsLower:    []string{"compress", "archive", "zip", "backup"},
			},
			{
				Command:          "grep -R pattern .",
				Description:      "search files recursively",
				Keywords:         []string{"search", "text", "pattern"},
				CommandLower:     "grep -r pattern .",
				DescriptionLower: "search files recursively",
				KeywordsLower:    []string{"search", "text", "pattern"},
			},
		},
	}

	db.BuildUniversalIndex()
	scores := map[int]float64{}
	terms := []string{"compress", "archive", "backup", "files", "folder", "create", "format", "bundle", "zip", "tar"}

	db.mergeFamilyExpansionCandidates(scores, terms, nil, SearchOptions{
		EnableFamilyExpansion:      true,
		FamilyExpansionMaxBases:    2,
		FamilyExpansionClarityMax:  1.0,
		FamilyExpansionBlendWeight: 0.25,
		Limit:                      5,
	})

	if scores[0] <= 0 && scores[1] <= 0 {
		t.Fatalf("expected at least one archive-family command to get expansion bonus, got scores=%v", scores)
	}
	if scores[2] > 0 {
		t.Fatalf("expected unrelated grep command to remain unboosted, got score=%f", scores[2])
	}
}

func TestMergeFamilyExpansionCandidatesClarityGateBlocksExpansion(t *testing.T) {
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
		},
	}
	db.BuildUniversalIndex()
	scores := map[int]float64{}
	terms := []string{"git", "commit", "changes", "repository", "history", "message", "branch", "remote", "push", "status"}

	db.mergeFamilyExpansionCandidates(scores, terms, nil, SearchOptions{
		EnableFamilyExpansion:      true,
		FamilyExpansionMaxBases:    2,
		FamilyExpansionClarityMax:  -1.0,
		FamilyExpansionBlendWeight: 0.25,
		Limit:                      5,
	})

	if len(scores) != 0 {
		t.Fatalf("expected no expansion due to strict clarity gate, got scores=%v", scores)
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

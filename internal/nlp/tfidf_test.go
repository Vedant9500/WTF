package nlp

import (
	"testing"
)

func TestTFIDFSingleDocumentReturnsResult(t *testing.T) {
	searcher := NewTFIDFSearcher([]Command{
		{Command: "ls", Description: "list files", Keywords: []string{"list", "files"}},
	})

	results := searcher.Search("list files", 5)
	if len(results) == 0 {
		t.Fatal("expected at least one result for single-document corpus")
	}

	if results[0].CommandIndex != 0 {
		t.Fatalf("expected top result CommandIndex 0, got %d", results[0].CommandIndex)
	}
}

func TestTFIDFSearchNegativeLimitDoesNotPanicAndReturnsEmpty(t *testing.T) {
	searcher := NewTFIDFSearcher([]Command{
		{Command: "ls", Description: "list files", Keywords: []string{"list", "files"}},
		{Command: "grep", Description: "search text", Keywords: []string{"search", "text"}},
	})

	results := searcher.Search("list files", -1)
	if len(results) != 0 {
		t.Fatalf("expected no results for negative limit, got %d", len(results))
	}
}

func TestTFIDFStatsEmptyCorpusAverageIsZero(t *testing.T) {
	searcher := NewTFIDFSearcher([]Command{})
	stats := searcher.GetVocabularyStats()

	avgRaw, ok := stats["avg_terms_per_command"]
	if !ok {
		t.Fatal("expected avg_terms_per_command in stats")
	}

	avg, ok := avgRaw.(float64)
	if !ok {
		t.Fatalf("expected avg_terms_per_command to be float64, got %T", avgRaw)
	}

	if avg != 0 {
		t.Fatalf("expected avg_terms_per_command to be 0 for empty corpus, got %v", avg)
	}
}

func TestTFIDFMatchesInflectedQueryTerms(t *testing.T) {
	searcher := NewTFIDFSearcher([]Command{
		{Command: "apt install", Description: "install packages", Keywords: []string{"install", "package"}},
		{Command: "rm", Description: "remove file", Keywords: []string{"remove", "file"}},
	})

	results := searcher.Search("installing packages", 5)
	if len(results) == 0 {
		t.Fatal("expected results for inflected query terms")
	}

	if results[0].CommandIndex != 0 {
		t.Fatalf("expected install command to rank first, got index %d", results[0].CommandIndex)
	}
}

func TestTFIDFBigramHelpsPhraseRanking(t *testing.T) {
	searcher := NewTFIDFSearcher([]Command{
		{Command: "git commit", Description: "record changes", Keywords: []string{"git", "commit"}},
		{Command: "git clone", Description: "clone repository", Keywords: []string{"git", "clone"}},
	})

	results := searcher.Search("git commit", 5)
	if len(results) == 0 {
		t.Fatal("expected phrase query to return results")
	}

	if results[0].CommandIndex != 0 {
		t.Fatalf("expected git commit command to rank first, got index %d", results[0].CommandIndex)
	}
}

package main

import "testing"

func TestFindBestRankUsesStrictPrefixRelevance(t *testing.T) {
	results := []string{"git reset --hard HEAD", "git branch --delete feature"}
	relevant := []string{"git branch"}

	if rank := findBestRank(results, relevant); rank != 2 {
		t.Fatalf("expected rank 2, got %d", rank)
	}
}

func TestFindBestRankMissReturnsZero(t *testing.T) {
	results := []string{"git reset --hard HEAD"}
	relevant := []string{"docker exec"}

	if rank := findBestRank(results, relevant); rank != 0 {
		t.Fatalf("expected miss rank 0, got %d", rank)
	}
}

func TestSearchOptionsForMode(t *testing.T) {
	bm25 := searchOptionsForMode("bm25")
	if !bm25.DisableEmbedding {
		t.Fatalf("expected bm25 mode to disable embeddings")
	}

	sentence := searchOptionsForMode("sentence")
	if !sentence.UseSentenceEmbeddings {
		t.Fatalf("expected sentence mode to opt into sentence-transformer embeddings")
	}
}

func TestNormalizeEvalModeDefaultsUnknownToHybrid(t *testing.T) {
	if mode := normalizeEvalMode("unknown"); mode != "hybrid" {
		t.Fatalf("expected unknown mode to normalize to hybrid, got %q", mode)
	}
}

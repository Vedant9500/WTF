package database

import (
	"testing"

	"github.com/Vedant9500/WTF/internal/embedding"
	"github.com/Vedant9500/WTF/internal/nlp"
)

func TestNewEmbeddingSearcherUsesExhaustiveSearchByDefault(t *testing.T) {
	searcher := newEmbeddingSearcher(&embedding.EnhancedIndex{Dimension: 2}, nlp.NewQueryProcessor())
	if searcher.useANN {
		t.Fatalf("expected exhaustive enhanced embedding search by default")
	}
}

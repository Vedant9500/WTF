package embedding

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"Git Commit -m 'message'", []string{"git", "commit", "message"}},
		{"compress   files", []string{"compress", "files"}},
		{"a b c", []string{}}, // Single letters filtered out
		{"", []string{}},
		{"decompress archive", []string{"decompress", "archive"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		a         []float32
		b         []float32
		expected  float64
		tolerance float64
	}{
		{
			name:      "identical vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{1, 0, 0},
			expected:  1.0,
			tolerance: 0.001,
		},
		{
			name:      "orthogonal vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{0, 1, 0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "opposite vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{-1, 0, 0},
			expected:  -1.0,
			tolerance: 0.001,
		},
		{
			name:      "similar vectors",
			a:         []float32{1, 1, 0},
			b:         []float32{1, 0, 0},
			expected:  0.707, // cos(45°) ≈ 0.707
			tolerance: 0.01,
		},
		{
			name:      "empty vectors",
			a:         []float32{},
			b:         []float32{},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "zero vector",
			a:         []float32{0, 0, 0},
			b:         []float32{1, 1, 1},
			expected:  0.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("CosineSimilarity() = %v, want %v (tolerance %v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestEmbeddingIndex_EmbedQuery(t *testing.T) {
	// Create a mock embedding index with some test vectors
	idx := &EmbeddingIndex{
		Dimension: 3,
		WordVectors: map[string][]float32{
			"compress": {1.0, 0.0, 0.0},
			"files":    {0.0, 1.0, 0.0},
			"archive":  {0.5, 0.5, 0.0},
		},
	}

	tests := []struct {
		query    string
		expected []float32
	}{
		{
			query:    "compress files",
			expected: []float32{0.5, 0.5, 0.0}, // Average of compress and files vectors
		},
		{
			query:    "unknown words only",
			expected: nil, // No matching words
		},
		{
			query:    "compress",
			expected: []float32{1.0, 0.0, 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := idx.EmbedQuery(tt.query)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("EmbedQuery(%q) = %v, want nil", tt.query, result)
				}
				return
			}

			if result == nil {
				t.Errorf("EmbedQuery(%q) = nil, want %v", tt.query, tt.expected)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("EmbedQuery(%q)[%d] = %v, want %v", tt.query, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestEmbeddingIndex_SemanticScores(t *testing.T) {
	idx := &EmbeddingIndex{
		Dimension: 3,
		WordVectors: map[string][]float32{
			"compress": {1.0, 0.0, 0.0},
		},
		CmdEmbeddings: [][]float32{
			{1.0, 0.0, 0.0}, // Identical to compress
			{0.0, 1.0, 0.0}, // Orthogonal
			{0.7, 0.7, 0.0}, // Somewhat similar
		},
	}

	queryEmbed := idx.EmbedQuery("compress")
	scores := idx.SemanticScores(queryEmbed)

	if len(scores) != 3 {
		t.Fatalf("SemanticScores returned %d scores, want 3", len(scores))
	}

	// First command should have highest similarity (1.0)
	if scores[0] < 0.99 {
		t.Errorf("scores[0] = %v, want ~1.0", scores[0])
	}

	// Second command should have similarity near 0
	if scores[1] > 0.01 || scores[1] < -0.01 {
		t.Errorf("scores[1] = %v, want ~0.0", scores[1])
	}

	// Third command should be somewhat similar
	if scores[2] < 0.5 || scores[2] > 0.8 {
		t.Errorf("scores[2] = %v, want ~0.7", scores[2])
	}
}

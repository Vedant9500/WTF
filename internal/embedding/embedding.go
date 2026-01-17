// Package embedding provides word vector embeddings for semantic search.
package embedding

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode"
)

// Index holds word vectors and pre-computed command embeddings.
type Index struct {
	Dimension     int
	WordVectors   map[string][]float32 // word -> 100d vector
	CmdEmbeddings [][]float32          // command index -> 100d vector
}

// LoadWordVectors loads word vectors from binary file.
// Format: [vocab_size:u32] then per word: [word_len:u16][word:bytes][vector:dim*f32]
func LoadWordVectors(filepath string) (*Index, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open word vectors: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read vocab size
	var vocabSize uint32
	if err := binary.Read(reader, binary.LittleEndian, &vocabSize); err != nil {
		return nil, fmt.Errorf("failed to read vocab size: %w", err)
	}

	idx := &Index{
		Dimension:   100, // GloVe 100d
		WordVectors: make(map[string][]float32, vocabSize),
	}

	// Read each word and vector
	for i := uint32(0); i < vocabSize; i++ {
		// Read word length
		var wordLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &wordLen); err != nil {
			return nil, fmt.Errorf("failed to read word length at %d: %w", i, err)
		}

		// Read word
		wordBytes := make([]byte, wordLen)
		if _, err := io.ReadFull(reader, wordBytes); err != nil {
			return nil, fmt.Errorf("failed to read word at %d: %w", i, err)
		}
		word := string(wordBytes)

		// Read vector
		vector := make([]float32, idx.Dimension)
		if err := binary.Read(reader, binary.LittleEndian, vector); err != nil {
			return nil, fmt.Errorf("failed to read vector at %d: %w", i, err)
		}

		idx.WordVectors[word] = vector
	}

	return idx, nil
}

// LoadCommandEmbeddings loads pre-computed command embeddings from binary file.
// Format: [num_commands:u32][dimension:u32] then per command: [embedding:dim*f32]
func (idx *Index) LoadCommandEmbeddings(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open command embeddings: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read header
	var numCommands, dimension uint32
	if err := binary.Read(reader, binary.LittleEndian, &numCommands); err != nil {
		return fmt.Errorf("failed to read num commands: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
		return fmt.Errorf("failed to read dimension: %w", err)
	}

	if int(dimension) != idx.Dimension {
		return fmt.Errorf("dimension mismatch: expected %d, got %d", idx.Dimension, dimension)
	}

	// Read embeddings
	idx.CmdEmbeddings = make([][]float32, numCommands)
	for i := uint32(0); i < numCommands; i++ {
		embedding := make([]float32, dimension)
		if err := binary.Read(reader, binary.LittleEndian, embedding); err != nil {
			return fmt.Errorf("failed to read embedding at %d: %w", i, err)
		}
		idx.CmdEmbeddings[i] = embedding
	}

	return nil
}

// EmbedQuery computes an embedding for a query by averaging word vectors.
func (idx *Index) EmbedQuery(query string) []float32 {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	// Sum vectors for matching tokens
	sum := make([]float32, idx.Dimension)
	count := 0

	for _, token := range tokens {
		if vec, ok := idx.WordVectors[token]; ok {
			for i, v := range vec {
				sum[i] += v
			}
			count++
		}
	}

	if count == 0 {
		return nil // No matching words
	}

	// Average
	for i := range sum {
		sum[i] /= float32(count)
	}

	return sum
}

// CosineSimilarity computes cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SemanticScores computes cosine similarity between query and all commands.
// Returns scores in same order as CmdEmbeddings.
func (idx *Index) SemanticScores(queryEmbedding []float32) []float64 {
	if queryEmbedding == nil || len(idx.CmdEmbeddings) == 0 {
		return nil
	}

	scores := make([]float64, len(idx.CmdEmbeddings))
	for i, cmdEmbed := range idx.CmdEmbeddings {
		scores[i] = CosineSimilarity(queryEmbedding, cmdEmbed)
	}

	return scores
}

// tokenize converts text to lowercase tokens for embedding lookup.
func tokenize(text string) []string {
	text = strings.ToLower(text)

	// Split on non-alphanumeric characters
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Filter short words
	var tokens []string
	for _, w := range words {
		if len(w) >= 2 {
			tokens = append(tokens, w)
		}
	}

	return tokens
}

// VocabSize returns the number of words in the vocabulary.
func (idx *Index) VocabSize() int {
	return len(idx.WordVectors)
}

// NumCommands returns the number of command embeddings.
func (idx *Index) NumCommands() int {
	return len(idx.CmdEmbeddings)
}

// HasWord checks if a word exists in the vocabulary.
func (idx *Index) HasWord(word string) bool {
	_, ok := idx.WordVectors[strings.ToLower(word)]
	return ok
}

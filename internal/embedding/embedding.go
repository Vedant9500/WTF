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

const (
	cmdEmbeddingsMagic   = "WTFE"
	cmdEmbeddingsVersion = uint16(1)
	sifSmoothingA        = 1e-3
)

// Index holds word vectors and pre-computed command embeddings.
type Index struct {
	Dimension         int
	WordVectors       map[string][]float32 // word -> 100d vector
	WordRanks         map[string]uint32    // word -> frequency rank in source vocab (0 = most frequent)
	CmdEmbeddings     [][]float32          // command index -> 100d vector
	CmdEmbeddingHash  string               // optional command snapshot hash from embedding file
	CmdFormatVersion  uint16               // embedding binary format version (legacy = 0)
	CmdSourceCommands uint32               // number of commands in source embedding file
	dominantComponent []float32
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
		WordRanks:   make(map[string]uint32, vocabSize),
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
		idx.WordRanks[word] = i
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

	// Read either the modern metadata header (magic-prefixed) or the legacy header.
	var (
		numCommands uint32
		dimension   uint32
		formatVer   uint16
		hash        string
	)

	prefix := make([]byte, 4)
	if _, err := io.ReadFull(reader, prefix); err != nil {
		return fmt.Errorf("failed to read command embedding header: %w", err)
	}

	if string(prefix) == cmdEmbeddingsMagic {
		formatVer = cmdEmbeddingsVersion

		var version uint16
		if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
			return fmt.Errorf("failed to read embedding format version: %w", err)
		}
		if version != cmdEmbeddingsVersion {
			return fmt.Errorf("unsupported embedding format version: %d", version)
		}

		// Reserved for future compatibility flags.
		var reserved uint16
		if err := binary.Read(reader, binary.LittleEndian, &reserved); err != nil {
			return fmt.Errorf("failed to read embedding reserved header: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &numCommands); err != nil {
			return fmt.Errorf("failed to read num commands: %w", err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
			return fmt.Errorf("failed to read dimension: %w", err)
		}

		var hashLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &hashLen); err != nil {
			return fmt.Errorf("failed to read command hash length: %w", err)
		}
		if hashLen > 0 {
			hashBytes := make([]byte, hashLen)
			if _, err := io.ReadFull(reader, hashBytes); err != nil {
				return fmt.Errorf("failed to read command hash: %w", err)
			}
			hash = string(hashBytes)
		}
	} else {
		// Legacy format: first 4 bytes were num_commands.
		numCommands = binary.LittleEndian.Uint32(prefix)
		if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
			return fmt.Errorf("failed to read dimension: %w", err)
		}
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
		normalizeVector(embedding)
		idx.CmdEmbeddings[i] = embedding
	}
	idx.fitAndRemoveDominantComponent()

	idx.CmdEmbeddingHash = hash
	idx.CmdFormatVersion = formatVer
	idx.CmdSourceCommands = numCommands

	return nil
}

// EmbedQuery computes an embedding for a query by averaging word vectors.
func (idx *Index) EmbedQuery(query string) []float32 {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	// Build a weighted centroid: SIF-like token weighting plus lightweight OOV recovery.
	sum := make([]float64, idx.Dimension)
	var totalWeight float64

	for _, token := range tokens {
		vec, weight, ok := idx.lookupWeightedVector(token)
		if !ok {
			continue
		}
		for i, v := range vec {
			sum[i] += float64(v) * weight
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return nil // No matching words
	}

	out := make([]float32, idx.Dimension)
	for i := range out {
		out[i] = float32(sum[i] / totalWeight)
	}
	idx.removeDominantComponent(out)
	if !normalizeVector(out) {
		return nil
	}

	return out
}

func (idx *Index) lookupWeightedVector(token string) ([]float32, float64, bool) {
	for _, candidate := range tokenVariants(token) {
		if vec, ok := idx.WordVectors[candidate]; ok {
			return vec, idx.tokenWeight(candidate), true
		}
	}
	return nil, 0, false
}

func (idx *Index) tokenWeight(token string) float64 {
	weight := 1.0
	vocab := len(idx.WordVectors)
	if vocab > 0 {
		if rank, ok := idx.WordRanks[token]; ok {
			p := float64(rank+1) / float64(vocab)
			weight *= sifSmoothingA / (sifSmoothingA + p)
		}
	}

	// Retain structured terms (digits/port-like tokens) slightly more strongly.
	for _, r := range token {
		if unicode.IsDigit(r) {
			weight *= 1.20
			break
		}
	}

	if len(token) <= 2 {
		weight *= 0.85
	}

	return weight
}

func tokenVariants(token string) []string {
	variants := []string{token}
	seen := map[string]bool{token: true}
	add := func(v string) {
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		variants = append(variants, v)
	}

	if strings.HasSuffix(token, "ing") && len(token) > 5 {
		add(token[:len(token)-3])
	}
	if strings.HasSuffix(token, "ed") && len(token) > 4 {
		add(token[:len(token)-2])
	}
	if strings.HasSuffix(token, "es") && len(token) > 4 {
		add(token[:len(token)-2])
	}
	if strings.HasSuffix(token, "s") && len(token) > 3 {
		add(token[:len(token)-1])
	}

	return variants
}

func normalizeVector(v []float32) bool {
	var norm float64
	for _, x := range v {
		norm += float64(x * x)
	}
	if norm == 0 {
		return false
	}
	inv := 1.0 / math.Sqrt(norm)
	for i := range v {
		v[i] = float32(float64(v[i]) * inv)
	}
	return true
}

func (idx *Index) fitAndRemoveDominantComponent() {
	if len(idx.CmdEmbeddings) < 16 || idx.Dimension <= 0 {
		idx.dominantComponent = nil
		return
	}

	comp := idx.estimateDominantComponent(24)
	if comp == nil {
		idx.dominantComponent = nil
		return
	}
	idx.dominantComponent = comp

	for i := range idx.CmdEmbeddings {
		idx.removeDominantComponent(idx.CmdEmbeddings[i])
		normalizeVector(idx.CmdEmbeddings[i])
	}
}

func (idx *Index) estimateDominantComponent(iters int) []float32 {
	if len(idx.CmdEmbeddings) == 0 {
		return nil
	}
	v := make([]float64, idx.Dimension)
	for i := range v {
		v[i] = 1.0 / math.Sqrt(float64(idx.Dimension))
	}

	for it := 0; it < iters; it++ {
		next := make([]float64, idx.Dimension)
		for _, row := range idx.CmdEmbeddings {
			var proj float64
			for j := range row {
				proj += float64(row[j]) * v[j]
			}
			for j := range row {
				next[j] += float64(row[j]) * proj
			}
		}
		var norm float64
		for _, x := range next {
			norm += x * x
		}
		if norm == 0 {
			return nil
		}
		inv := 1.0 / math.Sqrt(norm)
		for j := range next {
			v[j] = next[j] * inv
		}
	}

	out := make([]float32, idx.Dimension)
	for i := range out {
		out[i] = float32(v[i])
	}
	if !normalizeVector(out) {
		return nil
	}
	return out
}

func (idx *Index) removeDominantComponent(vec []float32) {
	if idx == nil || len(idx.dominantComponent) == 0 || len(vec) != len(idx.dominantComponent) {
		return
	}
	var proj float64
	for i := range vec {
		proj += float64(vec[i]) * float64(idx.dominantComponent[i])
	}
	for i := range vec {
		vec[i] -= float32(proj * float64(idx.dominantComponent[i]))
	}
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

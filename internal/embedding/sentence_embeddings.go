// Package embedding provides sentence-transformer embedding support.
package embedding

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// SentenceIndex holds sentence-transformer embeddings for commands.
type SentenceIndex struct {
	Dimension     int
	CmdEmbeddings [][]float32 // sentence-transformer embeddings (384d for MiniLM)
	CmdEmbeddingHash string
	FormatVersion uint16
	TotalCommands int
}

// LoadSentenceEmbeddings loads sentence-transformer embeddings from binary file.
func (idx *SentenceIndex) LoadSentenceEmbeddings(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open sentence embeddings: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read and verify header
	magic := make([]byte, 4)
	if _, err := io.ReadFull(reader, magic); err != nil {
		return fmt.Errorf("failed to read magic header: %w", err)
	}

	if string(magic) != "WTFS" {
		return fmt.Errorf("invalid magic: expected WTFS, got %s", string(magic))
	}

	var version uint16
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if version != 3 {
		return fmt.Errorf("unsupported version: expected 3, got %d", version)
	}

	// Read header fields
	var numCommands, dimension uint32
	if err := binary.Read(reader, binary.LittleEndian, &numCommands); err != nil {
		return fmt.Errorf("failed to read num commands: %w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
		return fmt.Errorf("failed to read dimension: %w", err)
	}

	// Read command hash
	var hashLen uint16
	if err := binary.Read(reader, binary.LittleEndian, &hashLen); err != nil {
		return fmt.Errorf("failed to read hash length: %w", err)
	}

	if hashLen > 0 {
		hashBytes := make([]byte, hashLen)
		if _, err := io.ReadFull(reader, hashBytes); err != nil {
			return fmt.Errorf("failed to read command hash: %w", err)
		}
		idx.CmdEmbeddingHash = string(hashBytes)
	}

	idx.Dimension = int(dimension)
	idx.TotalCommands = int(numCommands)
	idx.FormatVersion = version

	// Read embeddings
	idx.CmdEmbeddings = make([][]float32, numCommands)
	for i := uint32(0); i < numCommands; i++ {
		embedding := make([]float32, dimension)
		if err := binary.Read(reader, binary.LittleEndian, embedding); err != nil {
			return fmt.Errorf("failed to read embedding %d: %w", i, err)
		}
		idx.CmdEmbeddings[i] = embedding
	}

	return nil
}

// Search finds the most similar commands to the query embedding.
func (idx *SentenceIndex) Search(queryEmbedding []float32, topK int) []SearchResult {
	if queryEmbedding == nil || len(idx.CmdEmbeddings) == 0 {
		return nil
	}

	candidates := make([]SearchResult, 0, len(idx.CmdEmbeddings))
	for i, cmdEmbed := range idx.CmdEmbeddings {
		sim := CosineSimilarity(queryEmbedding, cmdEmbed)
		if sim > 0.0 {
			candidates = append(candidates, SearchResult{
				CommandIndex: i,
				Similarity:   sim,
				Score:        sim, // Raw cosine similarity
			})
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Return top K
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	return candidates
}

// QueryEmbedIndex holds pre-computed sentence-transformer embeddings for queries.
type QueryEmbedIndex struct {
	Dimension  int
	Queries    []string          // query texts in order
	Embeddings [][]float32       // pre-computed query embeddings
	queryMap   map[string]int    // query text -> index
}

// LoadQueryEmbeddings loads pre-computed query embeddings from binary file.
func (idx *QueryEmbedIndex) LoadQueryEmbeddings(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open query embeddings: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read and verify header
	magic := make([]byte, 4)
	if _, err := io.ReadFull(reader, magic); err != nil {
		return fmt.Errorf("failed to read magic header: %w", err)
	}

	if string(magic) != "WTQE" {
		return fmt.Errorf("invalid magic: expected WTQE, got %s", string(magic))
	}

	var version uint16
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if version != 1 {
		return fmt.Errorf("unsupported version: expected 1, got %d", version)
	}

	var numQueries, dimension uint32
	if err := binary.Read(reader, binary.LittleEndian, &numQueries); err != nil {
		return fmt.Errorf("failed to read num queries: %w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
		return fmt.Errorf("failed to read dimension: %w", err)
	}

	idx.Dimension = int(dimension)
	idx.Queries = make([]string, numQueries)
	idx.Embeddings = make([][]float32, numQueries)
	idx.queryMap = make(map[string]int, numQueries)

	// Read query texts and embeddings
	for i := uint32(0); i < numQueries; i++ {
		var queryLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &queryLen); err != nil {
			return fmt.Errorf("failed to read query length %d: %w", i, err)
		}

		queryBytes := make([]byte, queryLen)
		if _, err := io.ReadFull(reader, queryBytes); err != nil {
			return fmt.Errorf("failed to read query text %d: %w", i, err)
		}
		query := string(queryBytes)
		idx.Queries[i] = query
		idx.queryMap[query] = int(i)

		// Read embedding
		embedding := make([]float32, dimension)
		if err := binary.Read(reader, binary.LittleEndian, embedding); err != nil {
			return fmt.Errorf("failed to read embedding %d: %w", i, err)
		}
		idx.Embeddings[i] = embedding
	}

	return nil
}

// GetQueryEmbedding returns the pre-computed embedding for a query.
func (idx *QueryEmbedIndex) GetQueryEmbedding(query string) []float32 {
	if idx == nil || idx.queryMap == nil {
		return nil
	}

	// Try exact match first
	if i, ok := idx.queryMap[query]; ok {
		return idx.Embeddings[i]
	}

	// Try case-insensitive match
	queryLower := strings.ToLower(query)
	for q, i := range idx.queryMap {
		if strings.ToLower(q) == queryLower {
			return idx.Embeddings[i]
		}
	}

	return nil
}

// HasQueryEmbedding returns true if a pre-computed embedding exists for the query.
func (idx *QueryEmbedIndex) HasQueryEmbedding(query string) bool {
	if idx == nil || idx.queryMap == nil {
		return false
	}
	if _, ok := idx.queryMap[query]; ok {
		return true
	}
	queryLower := strings.ToLower(query)
	for q := range idx.queryMap {
		if strings.ToLower(q) == queryLower {
			return true
		}
	}
	return false
}

// Package embedding provides sentence-transformer embedding support.
package embedding

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
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

package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Vedant9500/WTF/internal/embedding"
)

// FindAssetPath searches for an asset file in common locations.
func FindAssetPath(filename string) string {
	// Check in current directory
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	// Check in assets/ subdirectory
	assetsPath := filepath.Join("assets", filename)
	if _, err := os.Stat(assetsPath); err == nil {
		return assetsPath
	}

	// Check relative to executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		exeAssets := filepath.Join(exeDir, "assets", filename)
		if _, err := os.Stat(exeAssets); err == nil {
			return exeAssets
		}
		// Also check directly next to executable
		exeFile := filepath.Join(exeDir, filename)
		if _, err := os.Stat(exeFile); err == nil {
			return exeFile
		}
	}

	return ""
}

// LoadEmbeddings loads word vectors and command embeddings for semantic search.
// This function gracefully handles missing files - if embeddings aren't available,
// the search will fall back to pure BM25F scoring.
func (db *Database) LoadEmbeddings() error {
	gloveFile := FindAssetPath("glove.bin")
	cmdEmbedFile := FindAssetPath("cmd_embeddings.bin")
	enhancedEmbedFile := FindAssetPath("enhanced_cmd_embeddings.bin")

	// If GloVe file doesn't exist, embeddings are optional
	if gloveFile == "" {
		log.Println("Note: glove.bin not found, semantic search disabled")
		return nil
	}

	// Load word vectors
	idx, err := embedding.LoadWordVectors(gloveFile)
	if err != nil {
		log.Printf("Warning: failed to load word vectors: %v", err)
		return nil
	}

	// Load command embeddings if available
	if cmdEmbedFile != "" {
		if err := idx.LoadCommandEmbeddings(cmdEmbedFile); err != nil {
			log.Printf("Warning: failed to load command embeddings: %v", err)
			// Still usable for query embedding, just no pre-computed cmd vectors
		} else {
			db.validateCommandEmbeddings(idx)
		}
	}

	db.embeddingIndex = idx
	log.Printf("Loaded semantic search: %d words, %d command embeddings",
		idx.VocabSize(), idx.NumCommands())

	// Load enhanced embeddings if available
	if enhancedEmbedFile != "" {
		if err := db.loadEnhancedEmbeddings(enhancedEmbedFile); err != nil {
			log.Printf("Note: enhanced embeddings not loaded: %v", err)
			// Continue with basic embeddings
		} else {
			db.initializeEmbeddingSearcher()
			log.Printf("Loaded enhanced embedding search with ANN indexing")
		}
	}

	return nil
}

// loadEnhancedEmbeddings loads enhanced field-aware embeddings.
func (db *Database) loadEnhancedEmbeddings(filepath string) error {
	enhancedIdx := &embedding.EnhancedIndex{
		Dimension:   100, // Same as GloVe dimension
		WordVectors: db.embeddingIndex.WordVectors,
		WordRanks:   db.embeddingIndex.WordRanks,
	}

	if err := enhancedIdx.LoadEnhancedEmbeddings(filepath); err != nil {
		return err
	}

	db.enhancedEmbeddingIndex = enhancedIdx
	return nil
}

func (db *Database) validateCommandEmbeddings(idx *embedding.Index) {
	if idx == nil || len(idx.CmdEmbeddings) == 0 {
		return
	}

	if len(idx.CmdEmbeddings) > len(db.Commands) {
		log.Printf("Warning: command embeddings count (%d) exceeds loaded commands (%d); semantic boost disabled",
			len(idx.CmdEmbeddings), len(db.Commands))
		idx.CmdEmbeddings = nil
		return
	}

	if idx.CmdEmbeddingHash == "" {
		// Legacy files don't carry command hashes. Keep compatibility but warn for observability.
		log.Printf("Warning: cmd_embeddings.bin has no command hash metadata (legacy format); alignment safety checks are limited")
		return
	}

	expectedHash := db.commandSnapshotHash(len(idx.CmdEmbeddings))
	if expectedHash == "" || expectedHash != idx.CmdEmbeddingHash {
		log.Printf("Warning: command embedding hash mismatch; semantic boost disabled (expected %s, got %s)",
			expectedHash, idx.CmdEmbeddingHash)
		idx.CmdEmbeddings = nil
		return
	}
}

// HasEmbeddings returns true if semantic search embeddings are loaded.
func (db *Database) HasEmbeddings() bool {
	return db.embeddingIndex != nil
}

// EmbedQuery computes an embedding for the given query text.
func (db *Database) EmbedQuery(query string) []float32 {
	if db.embeddingIndex == nil {
		return nil
	}
	return db.embeddingIndex.EmbedQuery(query)
}

// SemanticScores computes cosine similarity between query and all commands.
func (db *Database) SemanticScores(queryEmbedding []float32) []float64 {
	if db.embeddingIndex == nil || queryEmbedding == nil {
		return nil
	}
	return db.embeddingIndex.SemanticScores(queryEmbedding)
}

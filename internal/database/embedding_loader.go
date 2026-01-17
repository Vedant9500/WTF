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
		}
	}

	db.embeddingIndex = idx
	log.Printf("Loaded semantic search: %d words, %d command embeddings",
		idx.VocabSize(), idx.NumCommands())

	return nil
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

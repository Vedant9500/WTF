package database

import (
	"testing"

	"github.com/Vedant9500/WTF/internal/embedding"
)

func TestCommandSnapshotHash_Deterministic(t *testing.T) {
	db := &Database{Commands: []Command{
		{Command: "git commit", Description: "Commit changes", Keywords: []string{"git", "commit"}},
		{Command: "tar", Description: "Create archives", Keywords: []string{"archive", "compress"}},
	}}

	h1 := db.commandSnapshotHash(2)
	h2 := db.commandSnapshotHash(2)
	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}
	if h1 != h2 {
		t.Fatalf("expected stable hash, got %s and %s", h1, h2)
	}
}

func TestValidateCommandEmbeddings_HashMismatchDisablesSemantic(t *testing.T) {
	db := &Database{Commands: []Command{
		{Command: "git commit", Description: "Commit changes", Keywords: []string{"git", "commit"}},
		{Command: "tar", Description: "Create archives", Keywords: []string{"archive", "compress"}},
	}}

	idx := &embedding.Index{
		Dimension:        3,
		CmdEmbeddings:    [][]float32{{1, 0, 0}, {0, 1, 0}},
		CmdEmbeddingHash: "deadbeef",
	}

	db.validateCommandEmbeddings(idx)
	if len(idx.CmdEmbeddings) != 0 {
		t.Fatalf("expected embeddings to be disabled on hash mismatch")
	}
}

func TestValidateCommandEmbeddings_PrefixHashWithPersonalCommands(t *testing.T) {
	main := []Command{
		{Command: "git commit", Description: "Commit changes", Keywords: []string{"git", "commit"}},
		{Command: "tar", Description: "Create archives", Keywords: []string{"archive", "compress"}},
	}
	personal := Command{Command: "my custom", Description: "Custom helper", Keywords: []string{"custom"}}

	db := &Database{Commands: append(append([]Command{}, main...), personal)}
	mainDB := &Database{Commands: main}

	idx := &embedding.Index{
		Dimension:        3,
		CmdEmbeddings:    [][]float32{{1, 0, 0}, {0, 1, 0}},
		CmdEmbeddingHash: mainDB.commandSnapshotHash(2),
	}

	db.validateCommandEmbeddings(idx)
	if len(idx.CmdEmbeddings) != 2 {
		t.Fatalf("expected embeddings to remain enabled, got %d", len(idx.CmdEmbeddings))
	}
}

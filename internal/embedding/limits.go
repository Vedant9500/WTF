package embedding

import "fmt"

const (
	maxEmbeddingVocabEntries = 1_000_000
	maxEmbeddingCommands     = 1_000_000
	maxQueryEmbeddings       = 100_000
	maxEmbeddingDimension    = 4096
	maxEmbeddingTokenBytes   = 256
	maxEmbeddingTextBytes    = 4096
)

func validateCount(name string, count, maxAllowed uint32) error {
	if count > maxAllowed {
		return fmt.Errorf("%s too large: %d exceeds maximum %d", name, count, maxAllowed)
	}
	return nil
}

func validateDimension(dimension uint32, expected int) error {
	if dimension == 0 || dimension > maxEmbeddingDimension {
		return fmt.Errorf("invalid dimension: %d", dimension)
	}
	if expected > 0 && int(dimension) != expected {
		return fmt.Errorf("dimension mismatch: expected %d, got %d", expected, dimension)
	}
	return nil
}

func validateTextLength(name string, length uint16) error {
	if length > maxEmbeddingTextBytes {
		return fmt.Errorf("%s too long: %d exceeds maximum %d", name, length, maxEmbeddingTextBytes)
	}
	return nil
}

func validateTokenLength(name string, length uint16) error {
	if length > maxEmbeddingTokenBytes {
		return fmt.Errorf("%s too long: %d exceeds maximum %d", name, length, maxEmbeddingTokenBytes)
	}
	return nil
}

# Embedding-Based Search Implementation

## Overview

This document describes the new embedding-based search system that supplements (and can replace) the existing BM25F/TF-IDF vocabulary-dependent search with a semantic, vocabulary-independent approach.

## Architecture

### Components

1. **Enhanced Command Embeddings** (`internal/embedding/enhanced_embedding.go`)
   - Field-aware embeddings (command name, description, keywords, tags)
   - Attention-like pooling for better token weighting
   - Subword features for OOV (Out-Of-Vocabulary) recovery
   - ANN (Approximate Nearest Neighbor) indexing for fast retrieval

2. **Embedding Generation Script** (`scripts/generate_enhanced_embeddings.py`)
   - Generates field-aware embeddings from command database
   - Uses GloVe word vectors as base
   - Applies SIF-like weighting with attention pooling
   - Stores metadata for enhanced ranking

3. **Embedding Search Integration** (`internal/database/embedding_search.go`)
   - Intent-aware query embedding
   - Field-aware scoring
   - Hybrid fusion with BM25F results
   - Graceful fallback to BM25F if embeddings unavailable

## How It Works

### Command Embedding Generation

```python
# Generate enhanced embeddings
python scripts/generate_enhanced_embeddings.py
```

This creates `assets/enhanced_cmd_embeddings.bin` with:
- **Pooled embedding**: Weighted average of all field embeddings
- **Field embeddings**: Separate embeddings for command, description, keywords, tags
- **Metadata**: Command name, platform, niche, pipeline flag

### Query Processing

1. **NLP Enhancement**: Query is processed to extract actions, targets, keywords, and intent
2. **Intent-Aware Embedding**: Query embedding is computed with intent-based token weighting
3. **ANN Search**: Fast nearest neighbor retrieval using clustered index
4. **Field-Aware Scoring**: Each field is scored separately and weighted differently
5. **Post-Scoring Boosts**: Metadata-based boosts (exact matches, niche relevance, etc.)

### Search Flow

```
User Query
    ↓
NLP Processing (actions, targets, intent)
    ↓
Query Embedding (intent-aware)
    ↓
ANN Search (fast retrieval)
    ↓
Field-Aware Scoring
    ↓
Post-Scoring Boosts
    ↓
Hybrid Fusion with BM25F (if available)
    ↓
Final Results
```

## Benefits Over BM25F/TF-IDF

### 1. **Vocabulary Independence**
- BM25F: Requires exact word matches or manual synonyms
- Embedding: Understands semantic similarity (e.g., "compress" ≈ "archive")

### 2. **Natural Language Queries**
- BM25F: Struggles with conversational queries
- Embedding: Excels at understanding intent from verbose queries

### 3. **Contextual Understanding**
- BM25F: Treats all terms equally
- Embedding: Weights terms by importance (actions > targets > keywords)

### 4. **OOV Recovery**
- BM25F: Fails on unseen words
- Embedding: Subword features provide graceful degradation

### 5. **Speed**
- BM25F: Scores all documents
- Embedding: ANN index provides O(log n) retrieval

## Configuration

### Search Options

```go
options := SearchOptions{
    UseEmbedding: true,  // Enable embedding search (default: true)
    UseNLP: true,        // Enable NLP processing
    Limit: 5,            // Number of results
}
```

### Embedding Parameters

Constants in `internal/embedding/enhanced_embedding.go`:

```go
FieldWeightCommand = 3.0    // Command name weight
FieldWeightKeyword = 2.0    // Keywords weight
FieldWeightDesc    = 1.0    // Description weight
FieldWeightTag     = 1.2    // Tags weight

SIFSmoothingA     = 1e-3    // SIF smoothing parameter
DigitBoost        = 1.20    // Boost for numeric tokens
ShortTokenPenalty = 0.85    // Penalty for short tokens

ANNNumPartitions   = 100    // ANN cluster count
ANNPointsPerLeaf   = 50     // Points per leaf node
```

## Usage Examples

### Basic Search

```go
db := &Database{}
db.LoadFromFile("commands.yml")

// Embedding search is automatic if enhanced embeddings are loaded
results := db.SearchUniversal("how to compress files", SearchOptions{
    Limit: 5,
    UseNLP: true,
})
```

### Force BM25F Only

```go
results := db.SearchUniversal("tar command", SearchOptions{
    Limit: 5,
    UseEmbedding: false, // Disable embedding search
})
```

### Platform-Specific Search

```go
results := db.SearchUniversal("list processes", SearchOptions{
    Limit: 5,
    Platforms: []string{"linux"},
})
```

## Performance

### Embedding Generation

- **Time**: ~2-3 seconds for 10,000 commands
- **Memory**: ~3-5 MB for embedding file
- **Vocabulary**: 100k GloVe words (100d vectors)

### Search Speed

- **ANN Search**: O(log n) average case
- **Exhaustive Search**: O(n) but vectorized
- **Typical Latency**: < 10ms for 10k commands

### Memory Usage

- **Word Vectors**: ~40 MB (100k words × 100d)
- **Command Embeddings**: ~5 MB (10k commands × 100d × 4 bytes)
- **ANN Index**: ~1 MB overhead

## Testing

### Generate Embeddings

```bash
# Prerequisites: GloVe vectors
python scripts/prepare_glove.py

# Generate enhanced embeddings
python scripts/generate_enhanced_embeddings.py
```

### Run Search Tests

```bash
# Build the project
go build ./cmd/wtf

# Test search
./wtf search "how to compress a directory into a tar file"

# Evaluation (if available)
go run cmd/eval/main.go
```

## Troubleshooting

### Embeddings Not Loading

Check logs for messages like:
- `Note: enhanced embeddings not loaded: ...`
- `Warning: failed to load command embeddings: ...`

Ensure `assets/enhanced_cmd_embeddings.bin` exists.

### Poor Search Results

1. **Check embedding quality**: Verify with `scripts/generate_enhanced_embeddings.py`
2. **Adjust field weights**: Modify `FieldWeight*` constants
3. **Tune ANN index**: Adjust `ANNNumPartitions` for speed/accuracy tradeoff
4. **Enable hybrid fusion**: Ensure BM25F index is also built

### Performance Issues

1. **Disable ANN**: Set `useANN: false` in `embeddingSearcher`
2. **Reduce topK**: Lower `topK` in `embeddingSearcher` (default: 50)
3. **Min score threshold**: Increase `minScore` to filter weak matches (default: 5.0)

## Future Improvements

1. **Dynamic Embedding Fine-Tuning**: Adapt embeddings based on user feedback
2. **Contextual Embeddings**: Use transformer-based models for better context
3. **Multi-Modal Search**: Combine embeddings with other signals (usage frequency, recency)
4. **Incremental Updates**: Update embeddings without full regeneration
5. **Quantization**: Reduce embedding size with product quantization

## References

- GloVe: Global Vectors for Word Representation (Stanford NLP)
- SIF (Smooth Inverse Frequency) weighting for sentence embeddings
- ANN (Approximate Nearest Neighbor) search algorithms
- BM25F: BM25 for weighted fields

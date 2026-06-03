# Embedding-Based Search

## Overview

WTF uses embeddings as an optional semantic layer on top of the normal BM25F/NLP search pipeline. The default runtime path is conservative: enhanced GloVe command embeddings produce semantic candidates, BM25F produces lexical candidates, and the two lists are fused with reciprocal-rank fusion.

Sentence-transformer command embeddings are supported only when a query embedding is available. Today the Go CLI does not run a transformer model at query time; `assets/query_embeddings.bin` is an offline/evaluation artifact. Because of that, sentence-transformer search is explicit opt-in and should not be treated as the default live CLI behavior.

## Runtime Search Flow

```text
User query
  -> argument parsing / SearchOptions normalization
  -> optional sentence-transformer search only when UseSentenceEmbeddings is true
  -> enhanced GloVe embedding candidate retrieval when enhanced assets are loaded
  -> BM25F lexical search with NLP, fuzzy, platform, and pipeline filters
  -> reciprocal-rank fusion of filtered embedding and BM25F results
  -> terminal rendering / preview / execution flow
```

The enhanced embedding path is exhaustive by default. For the current local corpus size, scoring every command vector is cheap and avoids the recall loss caused by the older single-cluster ANN shortcut. The ANN index can still exist in memory, but the default `embeddingSearcher` does not use it.

## Assets

`assets/glove.bin`
: Word vectors used to embed live queries.

`assets/cmd_embeddings.bin`
: Legacy pooled command vectors used for semantic boosting. Newer files carry a command snapshot hash.

`assets/enhanced_cmd_embeddings.bin`
: Field-aware command vectors used by the enhanced hybrid search path. Format v3 includes a command snapshot hash. Legacy v2 files remain loadable, but alignment checks are limited.

`assets/sentence_cmd_embeddings.bin`
: Sentence-transformer command vectors. These are not enough for arbitrary live queries unless a matching query embedding source exists.

`assets/query_embeddings.bin`
: Pre-computed sentence-transformer embeddings for known evaluation queries. This is for evaluation/experimentation, not general CLI search.

## Generating Enhanced Embeddings

```bash
python scripts/prepare_glove.py
python scripts/generate_enhanced_embeddings.py
```

The generator writes:

- pooled command embeddings
- separate command, description, keyword, and tag field embeddings
- command metadata used for ranking diagnostics
- a command snapshot hash used by the Go loader to detect stale or misordered assets

The hash signature matches the runtime loader: normalized command, description, and keywords for each command in order.

## Search Options

```go
results := db.SearchUniversal("how to compress files", database.SearchOptions{
    Limit:  5,
    UseNLP: true,
})
```

Embedding search is enabled by default when assets are available. To force lexical BM25F/NLP search:

```go
results := db.SearchUniversal("tar command", database.SearchOptions{
    Limit:            5,
    DisableEmbedding: true,
})
```

To explicitly evaluate pre-computed sentence-transformer query embeddings:

```go
results := db.SearchUniversal("git undo", database.SearchOptions{
    Limit:                 5,
    UseNLP:                true,
    UseSentenceEmbeddings: true,
})
```

## Evaluation

The focused embedding evaluator can compare the live hybrid path, lexical-only search, and pre-computed sentence-transformer mode:

```bash
go run ./cmd/eval_embedding -mode hybrid
go run ./cmd/eval_embedding -mode bm25
go run ./cmd/eval_embedding -mode sentence
```

`-mode sentence` only works for queries present in `assets/query_embeddings.bin`. Unknown queries fall back through the normal search pipeline.

The broader evaluator remains:

```bash
go run ./cmd/eval -set all -limit 5
```

Prefer the stricter evaluator metrics when judging real search quality. Loose first-token matching can hide failures for command families such as `git reset` versus `git revert`.

## Performance Notes

- Enhanced GloVe search is O(n) over command vectors by default. With thousands or low tens of thousands of commands this is usually faster than the user-visible CLI overhead and improves recall.
- BM25F index construction is lazy and cached per database snapshot.
- Loading word vectors dominates memory use. A 100k x 100d float32 GloVe asset is roughly 40 MB before map overhead.
- Sentence-transformer search needs either a runtime embedding backend or pre-computed query vectors. Without that, arbitrary live query quality cannot match Python offline evaluation.

## Troubleshooting

If embeddings do not load, check logs for:

- `glove.bin not found`
- `enhanced embeddings not loaded`
- `command embedding hash mismatch`
- `sentence-transformer embedding hash mismatch`

Hash mismatch means the embedding asset was generated from a different command ordering or content snapshot. Regenerate the asset from the current `assets/commands.yml`.

If results are underwhelming:

- compare `go run ./cmd/eval_embedding -mode hybrid` with `-mode bm25`
- verify `enhanced_cmd_embeddings.bin` is v3 and matches the current command snapshot
- do not use `query_embeddings.bin` as evidence for arbitrary live query performance
- inspect misses by slice with `go run ./cmd/eval -set all -limit 5`

# WTF Retrieval Improvement Plan (Offline, Lightweight, Non-LLM)

## 1. Goal
Build a scalable command retrieval system that handles natural language requests well without:
- LLM APIs
- heavyweight online inference
- per-command hint maps
- query-specific manual tuning

Target corpus: ~6.6k commands and growing.

## 2. Constraints
- Offline-first runtime.
- Low memory and CPU overhead.
- Deterministic behavior.
- Scalable to many commands without manual curation.
- Keep current CLI UX and platform filtering behavior.

## 3. Key Findings So Far
- Manual hints and weak-query patches can improve benchmark numbers short-term, but do not scale.
- Removing handcrafted hinting exposed a real semantic gap on long conversational queries.
- Current stack (BM25F + lightweight NLP + word vectors) is not sufficient alone for robust intent-level retrieval.
- Benchmark expansion reduced bias and gave a more realistic quality signal.
- Documents are very short (~2–3 tokens cmd, ~8–15 tokens desc, ~3–5 tokens keywords), which limits the impact of techniques designed for long-document corpora.
- GloVe embeddings and command embeddings are already integrated and contributing semantic signal — this is an active component, not optional.

## 4. Design Principles
1. Prefer data-driven, corpus-wide methods over command-specific rules.
2. Keep retrieval interpretable and debuggable.
3. Add components behind measurable gates and feature flags.
4. Measure slice-wise quality, not only aggregate metrics.
5. Keep each stage modular so we can rollback isolated changes.
6. Validate every change against a properly-sized evaluation set before merging.

## 5. Current Baseline Components
- BM25F-style lexical retrieval with field weighting (k1=1.2, cmd:3.5, desc:1.0, keys:2.0, tags:1.2).
- NLP normalization, action/target extraction, term expansion via `nlp.ProcessQuery()`.
- Fuzzy fallback for zero-result recovery.
- GloVe-based semantic scoring (active, blended at 0.35/0.55 weight in `SearchUniversal`).
- TF-IDF cosine reranking for top candidates.
- Learned family prior index (`learnedFamilyIndex`) for corpus-driven command-base associations.
- Cascading NLP boost with action/target/keyword/context/intent signals.
- Long-query normalization and term importance weighting.
- Evaluation harness with Top1, Hit@K, MRR, NDCG@K across 36 short + 27 long queries.

## 6. Phase Plan

### [COMPLETED] Phase 0: Evaluation Infrastructure (Pre-requisite)
Objective: Build a robust evaluation foundation before making retrieval changes.

Work items:
- [x] 1. Expand eval set from 63 to 150–200 queries:
   - Ensure ≥10 queries per slice for statistical significance.
   - Add "easy" queries (exact command names) to detect trivial regressions.
   - Add explicit negative relevance judgments for common false positives.
   - Include more typo variants (currently only 2 queries).
- [x] 2. Split eval into dev set (80%) and held-out test set (20%).
- [x] 3. Add `make bench-eval` target that runs both eval sets and outputs a markdown comparison table with per-slice breakdowns.
- [x] 4. Measure current `hints.go` contribution by running benchmarks with hints enabled vs disabled.
- [x] 5. Add feature flag infrastructure so each phase can be toggled via config without code changes.

Success criteria:
- [x] Eval set has ≥150 queries with proper slice coverage.
- [x] Dev/test split established; test set is never used during development.
- [x] Benchmark automation produces reproducible comparison reports.

---

### Phase 1: Lexical Retrieval Upgrades (High ROI, Low Risk)
Objective: Improve long-query and paraphrase handling using classic IR only.

Work items:
- [x] 1. **Bigram/phrase indexing** for `command` and `keywords` fields:
   - Multi-word command names like `git reset`, `docker build`, `pip install` are effectively compound tokens — bigrams will match them naturally.
   - Skip `description` field initially (too short for meaningful phrase matching beyond what multi-term BM25F already provides). Extend only if measured improvement justifies it.
- [x] 2. **BM25F parameter sweep** (replaces BM25+ evaluation):
   - BM25+ solves over-penalization of long documents — irrelevant for this corpus where docs are uniformly short (length normalization term ≈ 1.0).
   - Instead, systematically tune the parameters that actually matter for short structured docs:
     - `k1` (term frequency saturation)
     - Per-field `b` values (length normalization per field)
     - Field weight ratios (cmd vs desc vs keys vs tags)
     - `minIDF` threshold (currently 0.0 — raising this filters low-info expansion terms)
   - `TopTermsCap` for long queries (candidate 8 identified in sweep; default kept at 10 pending split-wise validation)
   - Use grid search or Bayesian optimization over the dev eval set.
- [x] 3. **Char n-gram feature channel** for typo robustness:
   - Current fuzzy search is all-or-nothing fallback. Char n-grams provide partial-match credit within the primary BM25F scoring path.
   - Add as a separate scoring channel blended with lexical score, not a replacement for fuzzy fallback.
- [x] 4. **Proximity scoring** (low priority, optional):
   - Research shows proximity helps most on long documents. With fields of 2–15 tokens, co-occurring query terms are inherently close together.
   - Only implement for `description` field, and only if bigrams + param sweep don't close the gap sufficiently.
   - Implemented as a description-only boost behind `DisableProximity` feature flag.
   - Ablation (dev, no-hints): Top-1 moved on 8 queries, RR improved on 5 and worsened on 2.
   - Long-query-only dev NDCG@3 improved from 0.1976 -> 0.2010, so channel is kept enabled by default.

Success criteria:
- Long-query NDCG@3 improves versus current no-hints baseline.
- No slice regression > 10% without explicit justification.
- Parameter sweep documents optimal BM25F settings with ablation results.

---

### Phase 2: Corpus-Native Query Expansion + PRF
Objective: Add scalable query expansion without handcrafted hints.

Work items:
1. **Corpus-native expansion via `learnedFamilyIndex`** (implement first):
   - The existing `tokenToBases` map already captures corpus-learned token→command-base associations.
   - Use this to expand queries with related command base names — this is essentially PRF native to the corpus, requiring minimal new infrastructure.
   - Example: query "compress folder" → expansion terms from bases associated with "compress" tokens (tar, gzip, zip).
   - This replaces the role of `hints.go` in a data-driven, scalable way.
2. **RM3-style expansion** (implement second, compare against corpus-native):
   - Adapt for short-document corpus:
     - Use `feedbackDocCount=3` (not the typical 10–20 — short docs produce sparse, noisy term distributions).
     - Weight feedback terms by source field (command-field terms are far more informative than description terms).
     - Field-weighted clarity measure instead of standard KL-divergence (which assumes long unigram documents).
   - Interpolation: `P_expanded(w) = (1 - λ)P(w|q) + λP(w|feedback)`.
3. **Query-clarity gating**:
   - Apply expansion only for low-clarity/ambiguous queries to avoid topic drift.
   - Adapted clarity scoring that accounts for structured short-document format.
4. **Rocchio-style fallback** for comparison against RM3.
5. **Tunable controls**:
   - feedbackDocCount (default: 3)
   - feedbackTermCount (default: 5)
   - interpolation coefficient λ (default: 0.3)
   - clarity threshold for gating

Success criteria:
- Long-query Hit@3 and NDCG@3 improve.
- Short-query precision remains stable (±2%).
- Corpus-native expansion matches or exceeds RM3 on at least half of slices (determines whether full RM3 is worth the complexity).

---

### Phase 3: Lightweight Learned Reranker (Non-Neural)
Objective: Replace brittle manual boosts with learned global weights. This is expected to be the highest-ROI phase.

The current pipeline already produces ~9 feature signals combined via hardcoded multipliers in `SearchUniversal()`. A learned model replaces all manual tuning constants.

Work items:
1. **Build feature vector per candidate** from existing signals:
   - BM25F score (`calculateInitialScores`)
   - TF-IDF cosine similarity (`calculateTFIDFCandidateScores`)
   - GloVe semantic similarity (`calculateSemanticCandidateScores`)
   - Query-term coverage (`calculateQueryCoverageBoost`)
   - Learned family prior (`calculateLearnedFamilyBoost`)
   - Cascading NLP boost score (`cascadingBoost`)
   - Intent match score (`calculateIntentBoost`)
   - Action+Target co-occurrence (binary)
   - Long-query intent feature (`calculateLongQueryIntentFeatureBoost`)
   - Field coverage (how many fields matched: cmd, desc, keys, tags)
   - PRF/expansion score (from Phase 2)
   - Char n-gram/fuzzy score (from Phase 1)
   - Platform compatibility (binary)
2. **Train pairwise linear ranker offline**:
   - LambdaRank-style pairwise training using relevant vs. non-relevant result pairs per query.
   - Requires expanding eval set to 200+ queries (started in Phase 0).
   - Logistic regression as simpler alternative if pairwise approach overfits.
3. **Serialize weights** as a small JSON file (`assets/reranker_weights.json`):
   - Load as `[]float64` in Go runtime.
   - Simple dot product: `score = Σ(weight_i × feature_i)`.
4. **Versioned model metadata** and easy rollback:
   - Track model version, training date, training set hash, eval metrics at time of training.
5. **Embed feature vector in `SearchResult`** for debugging and interpretability.

Success criteria:
- Aggregate and slice metrics improve over Phase 2.
- No command-specific rule tables needed.
- All hardcoded multipliers in `SearchUniversal()` can be removed.

---

### Phase 4: Semantic Signal Quality Upgrade
Objective: Improve the existing semantic signal from general-domain GloVe to corpus-aware embeddings.

Current state: GloVe vectors (40MB `glove.bin`) and precomputed command embeddings (`cmd_embeddings.bin`) are already integrated and blended in `SearchUniversal()`. The issue isn't absence of semantic search — it's that general-domain vectors don't capture CLI-specific relationships well (e.g., "tar" ≈ "archive" ≈ "compress").

Options (evaluate in order):
1. **Corpus-trained fastText embeddings** over `commands.yml` text:
   - Captures domain-specific word relationships.
   - Subword model handles typos natively (e.g., "comprss" → similar to "compress").
   - Smaller model size (corpus is only ~6.6k docs).
2. **LSA (TF-IDF + truncated SVD)** over the BM25F field representation:
   - Lightweight latent semantic matching for paraphrase handling.
   - No external model training needed — computed from corpus at build time.
3. **Random Indexing** as a low-memory alternative to LSA.

Integration:
- Use as one feature in the Phase 3 reranker, replacing current GloVe signal.
- Keep component swappable via config flag.
- A/B test against current GloVe baseline to confirm improvement.

Success criteria:
- Paraphrase and conversational intent slices improve over GloVe baseline.
- Runtime latency stays within 10% of current semantic scoring.
- Model size ≤ current GloVe size (40MB).

---

### Phase 5: Cleanup and Deprecation
Objective: Remove legacy systems that the learned reranker has replaced.

Work items:
1. **Deprecate `hints.go`** (294 lines of manual action→command mappings):
   - Phase 0 measures its contribution; Phase 2 corpus-native expansion replaces its function.
   - Remove after confirming reranker + expansion matches or exceeds hint-aided results.
2. **Remove hardcoded query-specific boosts** in `search_universal.go`:
   - `calculateTargetedQueryIntentBoost()` and its sub-functions (`applyProcessPortBoost`, `applyTarGzBoost`, etc.) are per-query-pattern hacks.
   - These should be fully replaced by the learned reranker features.
3. **Remove hardcoded multipliers** (`* 0.35`, `* 0.45`, `* 0.55`, etc.) from `SearchUniversal()`:
   - Replaced by learned weights from Phase 3.
4. **Clean up unused constants** in `constants.go` after above removals.

Success criteria:
- All eval metrics remain stable or improve after removals.
- No command-specific or query-specific hardcoded rules remain in the search path.

## 7. Evaluation Strategy

### Datasets
- **Dev set** (80% of queries):
  - Used for development, parameter tuning, and ablation testing.
  - assets/eval_queries.yaml (short queries)
  - assets/eval_queries_long.yaml (long queries)
- **Held-out test set** (20% of queries):
  - Used only for final phase validation. Never tuned against.
- Continue slice labels and expand challenging slices over time.
- Periodically refresh both sets to prevent evaluation overfitting.

### Metrics
Track at minimum:
- Top1 (exact first result accuracy)
- Hit@3 (relevant result in top 3)
- MRR (Mean Reciprocal Rank)
- NDCG@3 (graded relevance ranking quality)
- Per-slice metrics and worst-query list
- Regression count (how many queries got worse per change)

### Gates
- Every phase must run both short and long benchmarks on the dev set.
- Reject changes that improve one slice while heavily regressing others (>10% drop).
- Maintain a changelog of metric deltas per phase.
- Final phase validation on held-out test set only.

### Automation
- `make bench-eval` runs full eval and outputs comparison report.
- CI integration to flag regressions on PRs.
- Feature flags for A/B comparison of each phase.

## 8. Implementation Sequence (Recommended)
- [x] 1. **Phase 0** — evaluation infrastructure and expanded eval set.
- [ ] 2. **Phase 1** — bigram indexing (cmd+keys) + BM25F param sweep + char n-grams.
- [ ] 3. **Phase 2** — corpus-native expansion via learnedFamilyIndex, then RM3 if needed.
- [ ] 4. **Phase 3** — learned reranker over feature vector (highest expected ROI).
- [ ] 5. **Phase 4** — corpus-trained fastText or LSA replacing GloVe.
- [ ] 6. **Phase 5** — deprecate hints.go, remove hardcoded boosts, clean up constants.

This ordering maximizes gains while keeping complexity and dependencies low. Phase 0 is a prerequisite — making retrieval changes without a robust eval set risks unmeasured regressions.

## 9. Risks and Mitigations
- **Risk:** Over-expansion hurts precision.
  - Mitigation: query-clarity gating, strict interpolation limits, corpus-native expansion as conservative first step.
- **Risk:** Feature creep in reranker.
  - Mitigation: start with small feature set and ablation tests; L1 regularization to auto-prune weak features.
- **Risk:** Evaluation overfitting.
  - Mitigation: held-out test set, periodic benchmark refresh, never tune against test split.
- **Risk:** Short-doc sparsity makes RM3 feedback noisy.
  - Mitigation: low feedbackDocCount (3), field-weighted feedback terms, compare against simpler corpus-native expansion before committing.
- **Risk:** Index build time increases with bigrams/n-grams.
  - Mitigation: add index serialization (save/load from binary cache) to keep startup fast.
- **Risk:** Removing hints.go causes regressions on specific query patterns.
  - Mitigation: measure hint contribution in Phase 0; only remove after learned reranker covers those patterns.

## 10. Out of Scope
- Remote LLM APIs.
- Heavy neural rerankers requiring GPU/large RAM.
- Per-command hardcoded maps as a primary strategy.
- Proximity scoring beyond description field (low ROI for short docs).

## 11. Immediate Next Steps
- [x] 1. Expand eval set to 150+ queries with proper slice coverage and dev/test split.
- [x] 2. Add `make bench-eval` automation target.
- [x] 3. Measure hints.go contribution (benchmark with/without).
- [x] 4. Implement Phase 1.1 bigram/phrase indexing for command and keywords fields.
- [x] 5. Run BM25F parameter sweep and document optimal settings.
- [ ] 6. Run full benchmark suite and document deltas before starting Phase 2.

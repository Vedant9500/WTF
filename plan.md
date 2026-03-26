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

## 4. Design Principles
1. Prefer data-driven, corpus-wide methods over command-specific rules.
2. Keep retrieval interpretable and debuggable.
3. Add components behind measurable gates.
4. Measure slice-wise quality, not only aggregate metrics.
5. Keep each stage modular so we can rollback isolated changes.

## 5. Current Baseline Components
- BM25F-style lexical retrieval with field weighting.
- NLP normalization, action/target extraction, term expansion.
- Fuzzy fallback.
- Optional embedding-based semantic signals.
- Evaluation harness with Top1, Hit@K, MRR, NDCG@K.

## 6. Phase Plan

### Phase 1: Lexical Retrieval Upgrades (High ROI, Low Risk)
Objective: Improve long-query and paraphrase handling using classic IR only.

Work items:
1. Add phrase/bigram indexing for command, keywords, and description fields.
2. Add proximity feature scoring:
   - ordered proximity bonus
   - unordered window proximity bonus
3. Evaluate BM25 normalization variants:
   - BM25F tuning
   - BM25+ style lower-bounded tf normalization
4. Improve typo robustness with controlled char n-gram feature channel.

Success criteria:
- Long-query NDCG@3 improves versus current no-hints baseline.
- No slice regression > 10% without explicit justification.

---

### Phase 2: Pseudo-Relevance Feedback (PRF)
Objective: Add scalable query expansion without handcrafted hints.

Work items:
1. Implement RM3-style expansion:
   - initial retrieval top-k docs
   - extract feedback term distribution
   - interpolate original query + feedback model
2. Add Rocchio-style fallback option for comparison.
3. Add query-clarity gating:
   - apply PRF only for low-clarity/ambiguous queries
4. Add controls:
   - feedbackDocCount
   - feedbackTermCount
   - interpolation coefficient

Success criteria:
- Long-query Hit@3 and NDCG@3 improve.
- Short-query precision remains stable.

---

### Phase 3: Lightweight Learned Reranker (Non-Neural)
Objective: Replace brittle manual boosts with learned global weights.

Work items:
1. Build a feature vector per candidate:
   - lexical score(s)
   - field coverage
   - proximity features
   - query-term coverage
   - PRF score
   - typo/fuzzy signals
   - platform compatibility signals
2. Train simple linear model offline:
   - logistic regression or pairwise linear ranker
3. Serialize weights and load locally in Go runtime.
4. Add versioned model metadata and easy rollback.

Success criteria:
- Aggregate and slice metrics improve over Phase 2.
- No command-specific rule tables needed.

---

### Phase 4: Optional Lightweight Distributional Semantics (Non-LLM)
Objective: Add corpus-trained semantic signal without transformer inference.

Options:
1. LSA (TF-IDF + truncated SVD).
2. Random Indexing.
3. Local subword embeddings (fastText-style training over corpus text).

Integration:
- Use as one feature in reranking, not as sole retriever.
- Keep component optional via config flag.

Success criteria:
- Improves paraphrase and conversational intent slices.
- Keeps runtime overhead within budget.

## 7. Evaluation Strategy

### Datasets
- Keep both:
  - assets/eval_queries.yaml
  - assets/eval_queries_long.yaml
- Continue slice labels and expand challenging slices over time.

### Metrics
Track at minimum:
- Top1
- Hit@3
- MRR
- NDCG@3
- Per-slice metrics and worst-query list

### Gates
- Every phase must run both short and long benchmarks.
- Reject changes that improve one slice while heavily regressing others.
- Maintain a changelog of metric deltas per phase.

## 8. Implementation Sequence (Recommended)
1. Phase 1 lexical improvements.
2. Phase 2 RM3 + clarity gating.
3. Phase 3 learned reranker.
4. Phase 4 optional distributional semantics.

This ordering maximizes gains while keeping complexity and dependencies low.

## 9. Risks and Mitigations
- Risk: Over-expansion hurts precision.
  - Mitigation: query-clarity gating and strict interpolation limits.
- Risk: Feature creep in reranker.
  - Mitigation: start with small feature set and ablation tests.
- Risk: Evaluation overfitting.
  - Mitigation: keep holdout queries and periodically refresh benchmark sets.

## 10. Out of Scope
- Remote LLM APIs.
- Heavy neural rerankers requiring GPU/large RAM.
- Per-command hardcoded maps as a primary strategy.

## 11. Immediate Next Steps
1. Implement Phase 1.1 bigram/phrase field indexing.
2. Implement Phase 1.2 proximity scoring in candidate ranking.
3. Run full benchmark suite and document deltas.
4. Start RM3 prototype on top of new lexical baseline.

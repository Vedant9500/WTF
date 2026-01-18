package database

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/Vedant9500/WTF/internal/constants"
	"github.com/Vedant9500/WTF/internal/nlp"
	"github.com/Vedant9500/WTF/internal/utils"
)

// universalIndex implements BM25F scoring with field weights.
type universalIndex struct {
	// term -> postings
	postings map[string][]posting

	// document lengths per field
	docLens []docLens

	// average document lengths per field
	avgLen docLensF

	// document frequency per term
	df map[string]int

	// total documents
	N int

	// scoring parameters
	params bm25fParams
}

type posting struct {
	docID int
	// term frequency per field
	tf fieldTF
}

type fieldTF struct {
	cmd  int
	desc int
	keys int
	tags int
}

type docLens struct {
	cmd  int
	desc int
	keys int
	tags int
}

type docLensF struct {
	cmd  float64
	desc float64
	keys float64
	tags float64
}

type bm25fParams struct {
	k1     float64
	b      docLensF // b per field
	w      docLensF // field weights
	minIDF float64  // minimum idf to count
}

// tokenizer for index/query
func normalizeAndTokenize(s string) []string {
	if s == "" {
		return nil
	}
	// Normalize similarly to NLP pipeline, then lowercase
	s = nlp.NormalizeText(s)
	lower := strings.ToLower(s)
	words := strings.FieldsFunc(lower, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })

	out := make([]string, 0, len(words))
	sw := stopWords
	for _, w := range words {
		if len(w) < 2 { // drop very short tokens
			continue
		}
		if sw[w] {
			continue
		}
		out = append(out, w)
	}
	return out
}

var stopWords = nlp.StopWords()

func defaultParams() bm25fParams {
	return bm25fParams{
		k1:     1.2,
		b:      docLensF{cmd: 0.75, desc: 0.75, keys: 0.7, tags: 0.7},
		w:      docLensF{cmd: 3.5, desc: 1.0, keys: 2.0, tags: 1.2}, // Increased cmd weight for better command matching
		minIDF: 0.0,
	}
}

// selectTopTerms keeps the most informative terms by IDF to avoid noise from long queries.
// Preserves original query terms to ensure important specific terms aren't lost.
func (db *Database) selectTopTerms(terms []string, maxTerms int) []string {
	if maxTerms <= 0 || len(terms) <= maxTerms || db.uIndex == nil {
		return terms
	}

	// Conservative approach: preserve the first few terms which are likely from the original query
	preserveCount := utils.Min(4, len(terms)) // Preserve at least first 4 terms

	list := db.scoreTerms(terms, preserveCount)

	if len(list) <= maxTerms {
		return flattenTermList(list)
	}

	return db.filterAndSortTerms(list, maxTerms)
}

type termWithScore struct {
	term       string
	idf        float64
	isOriginal bool
}

func (db *Database) scoreTerms(terms []string, preserveCount int) []termWithScore {
	idx := db.uIndex
	seen := map[string]bool{}
	list := make([]termWithScore, 0, len(terms))

	for i, t := range terms {
		if seen[t] {
			continue
		}
		seen[t] = true
		df, ok := idx.df[t]
		if !ok || df == 0 {
			// If term not in index, still preserve it if it's from original query
			if i < preserveCount {
				list = append(list, termWithScore{term: t, idf: 1.0, isOriginal: true})
			}
			continue
		}
		isOriginal := i < preserveCount
		list = append(list, termWithScore{term: t, idf: bm25IDF(idx.N, df), isOriginal: isOriginal})
	}
	return list
}

func flattenTermList(list []termWithScore) []string {
	out := make([]string, 0, len(list))
	for _, it := range list {
		out = append(out, it.term)
	}
	return out
}

func (db *Database) filterAndSortTerms(list []termWithScore, maxTerms int) []string {
	// Separate original and enhanced terms
	var originalList, enhancedList []termWithScore
	for _, item := range list {
		if item.isOriginal {
			originalList = append(originalList, item)
		} else {
			enhancedList = append(enhancedList, item)
		}
	}

	// Always keep all original terms, then add best enhanced terms
	out := make([]string, 0, maxTerms)

	// Add all original terms first
	for _, item := range originalList {
		out = append(out, item.term)
	}

	// Add enhanced terms by IDF score to fill remaining slots
	remaining := maxTerms - len(out)
	if remaining > 0 && len(enhancedList) > 0 {
		sort.Slice(enhancedList, func(i, j int) bool {
			return enhancedList[i].idf > enhancedList[j].idf
		})
		for i := 0; i < utils.Min(remaining, len(enhancedList)); i++ {
			out = append(out, enhancedList[i].term)
		}
	}

	return out
}

// BuildUniversalIndex constructs the inverted index. Call after loading/merging commands.
func (db *Database) BuildUniversalIndex() {
	idx := &universalIndex{
		postings: make(map[string][]posting),
		df:       make(map[string]int),
		N:        len(db.Commands),
		params:   defaultParams(),
	}

	if idx.N == 0 {
		db.uIndex = idx
		return
	}

	idx.docLens = make([]docLens, idx.N)

	// First pass: tokenize per field, accumulate TFs and lengths per doc
	perDocTFs := make([]map[string]fieldTF, idx.N)
	for i := range db.Commands {
		lens, tf := indexCommand(&db.Commands[i])
		idx.docLens[i] = lens
		perDocTFs[i] = tf

		// update df once per term per doc
		for term := range tf {
			idx.df[term]++
		}
	}

	// compute avg lengths
	var sum docLens
	for _, l := range idx.docLens {
		sum.cmd += l.cmd
		sum.desc += l.desc
		sum.keys += l.keys
		sum.tags += l.tags
	}
	n := float64(idx.N)
	idx.avgLen = docLensF{cmd: float64(sum.cmd) / n, desc: float64(sum.desc) / n, keys: float64(sum.keys) / n, tags: float64(sum.tags) / n}

	// build postings
	for docID, tf := range perDocTFs {
		for term, ftf := range tf {
			idx.postings[term] = append(idx.postings[term], posting{docID: docID, tf: ftf})
		}
	}

	db.uIndex = idx
}

// buildTFIDFSearcher constructs a TF-IDF searcher and index map for reranking.
func (db *Database) buildTFIDFSearcher() {
	if len(db.Commands) == 0 {
		db.tfidf = nil
		db.cmdIndex = nil
		return
	}
	cmds := make([]nlp.Command, len(db.Commands))
	for i, c := range db.Commands {
		cmds[i] = nlp.Command{
			Command:     c.Command,
			Description: c.Description,
			Keywords:    c.Keywords,
		}
	}
	db.tfidf = nlp.NewTFIDFSearcher(cmds)
	db.cmdIndex = make(map[*Command]int, len(db.Commands))
	for i := range db.Commands {
		db.cmdIndex[&db.Commands[i]] = i
	}
}

// SearchUniversal performs BM25F search over the index with optional platform/pipeline filters.
func (db *Database) SearchUniversal(query string, options SearchOptions) []SearchResult {
	if db.uIndex == nil || db.uIndex.N != len(db.Commands) {
		// (Re)build lazily if needed
		db.BuildUniversalIndex()
	}

	if options.Limit <= 0 {
		options.Limit = 10
	}

	terms := normalizeAndTokenize(query)
	var pq *nlp.ProcessedQuery

	// NLP enhancement
	if options.UseNLP {
		pq, terms = db.enhanceQueryWithNLP(query, terms)
	}

	// If no terms after processing, try fuzzy search as fallback
	if len(terms) == 0 {
		if options.UseFuzzy {
			return db.performFuzzySearch(query, options)
		}
		return nil
	}

	// Reduce noise for long queries by keeping top-IDF terms
	termsCap := options.TopTermsCap
	if termsCap <= 0 {
		termsCap = 10
	}
	terms = db.selectTopTerms(terms, termsCap)

	// Calculate initial scores using BM25F
	scores := db.calculateInitialScores(terms, pq, options)

	// If no BM25F results, try fuzzy search as fallback for typos
	if len(scores) == 0 {
		if options.UseFuzzy {
			return db.performFuzzySearch(query, options)
		}
		return nil
	}

	// Convert to results and apply pipeline boosts
	results := db.collectResults(scores, pq, options)

	// Sort preliminarily
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })

	// Apply all post-scoring boosts (NLP reranking, cascading, semantic)
	results = db.applyPostScoringBoosts(results, pq, query, options)

	// Limit
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}
	return results
}

func (db *Database) calculateInitialScores(terms []string, pq *nlp.ProcessedQuery, options SearchOptions) map[int]float64 {
	idx := db.uIndex
	scores := make(map[int]float64, len(db.Commands)/4)
	currentPlatform := getCurrentPlatform()

	// Prepare per-term boosts (context + NLP action/target emphasis)
	termBoost := make(map[string]float64)
	for k, v := range options.ContextBoosts {
		termBoost[k] = v
	}
	if pq != nil {
		for _, a := range pq.Actions {
			if termBoost[a] < 2.0 {
				termBoost[a] = 2.0
			}
		}
		for _, t := range pq.Targets {
			if termBoost[t] < 1.6 {
				termBoost[t] = 1.6
			}
		}
	}

	// accumulate scores
	for _, term := range terms {
		postings, ok := idx.postings[term]
		if !ok {
			continue
		}
		idf := bm25IDF(idx.N, idx.df[term])
		if idf < idx.params.minIDF {
			continue
		}
		boost := 1.0
		if b, ok := termBoost[term]; ok && b > 0 {
			boost = b
		}
		db.processPostingsForTerm(postings, idx, idf, boost, scores, currentPlatform, options)
	}
	return scores
}

func (db *Database) processPostingsForTerm(
	postings []posting,
	idx *universalIndex,
	idf, boost float64,
	scores map[int]float64,
	currentPlatform string,
	options SearchOptions,
) {
	for _, p := range postings {
		doc := &db.Commands[p.docID]

		// Platform filtering (skip if AllPlatforms is enabled)
		if !options.AllPlatforms && len(doc.Platform) > 0 {
			if !isPlatformCompatible(doc.Platform, currentPlatform) && !isCrossPlatformTool(doc.Command) {
				continue
			}
		}

		// Pipeline filtering
		if options.PipelineOnly && !isPipelineCommand(doc) {
			continue
		}

		s := scores[p.docID]
		s += (idf * boost) * idx.termBM25F(p.docID, p.tf)
		scores[p.docID] = s
	}
}

func (db *Database) enhanceQueryWithNLP(query string, terms []string) (pq *nlp.ProcessedQuery, enhancedTerms []string) {
	processor := nlp.NewQueryProcessor()
	pq = processor.ProcessQuery(query)
	enh := pq.GetEnhancedKeywords()

	// Add relevant enhanced terms that aren't already present
	if len(enh) > 0 {
		for _, enhTerm := range enh {
			found := false
			for _, origTerm := range terms {
				if origTerm == enhTerm {
					found = true
					break
				}
			}
			if !found && len(terms) < 8 {
				terms = append(terms, enhTerm)
			}
		}
	}
	return pq, terms
}

func indexCommand(cmd *Command) (uniqueLens docLens, termFreqs map[string]fieldTF) {
	// prefer cached lowercase fields if available
	cmdText := cmd.Command
	if cmd.CommandLower != "" {
		cmdText = cmd.CommandLower
	}
	descText := cmd.Description
	if cmd.DescriptionLower != "" {
		descText = cmd.DescriptionLower
	}

	// tokens per field
	cmdTokens := normalizeAndTokenize(cmdText)
	descTokens := normalizeAndTokenize(descText)
	keysTokens := make([]string, 0)
	if len(cmd.KeywordsLower) > 0 {
		keysTokens = normalizeAndTokenize(strings.Join(cmd.KeywordsLower, " "))
	} else if len(cmd.Keywords) > 0 {
		keysTokens = normalizeAndTokenize(strings.Join(cmd.Keywords, " "))
	}
	tagsTokens := make([]string, 0)
	if len(cmd.TagsLower) > 0 {
		tagsTokens = normalizeAndTokenize(strings.Join(cmd.TagsLower, " "))
	} else if len(cmd.Tags) > 0 {
		tagsTokens = normalizeAndTokenize(strings.Join(cmd.Tags, " "))
	}

	// record doc lengths
	uniqueLens = docLens{cmd: len(cmdTokens), desc: len(descTokens), keys: len(keysTokens), tags: len(tagsTokens)}

	// term frequencies
	termFreqs = make(map[string]fieldTF)
	inc := func(tok string, f string) {
		entry := termFreqs[tok]
		switch f {
		case "cmd":
			entry.cmd++
		case "desc":
			entry.desc++
		case "keys":
			entry.keys++
		case "tags":
			entry.tags++
		}
		termFreqs[tok] = entry
	}

	for _, t := range cmdTokens {
		inc(t, "cmd")
	}
	for _, t := range descTokens {
		inc(t, "desc")
	}
	for _, t := range keysTokens {
		inc(t, "keys")
	}
	for _, t := range tagsTokens {
		inc(t, "tags")
	}
	return uniqueLens, termFreqs
}

func (db *Database) collectResults(scores map[int]float64, pq *nlp.ProcessedQuery, options SearchOptions) []SearchResult {
	results := make([]SearchResult, 0, utils.Min(len(scores), options.Limit*3))
	for docID, score := range scores {
		cmd := &db.Commands[docID]

		// Apply intent-based boost if NLP is active
		if pq != nil {
			score *= calculateIntentBoost(cmd, pq)
			// Co-occurrence boost
			docText := cmd.CommandLower + " " + cmd.DescriptionLower
			if containsAnyLocal(docText, pq.Actions) && containsAnyLocal(docText, pq.Targets) {
				score *= 1.2
			}
		}
		if isPipelineCommand(cmd) && options.PipelineBoost > 0 {
			score *= options.PipelineBoost
		}
		results = append(results, SearchResult{Command: cmd, Score: score})
	}
	return results
}

func (db *Database) rerankWithNLP(results []SearchResult, query string, options SearchOptions) []SearchResult {
	// Take a larger top slice for reranking to ensure good candidates aren't missed
	// Minimum 10 to ensure NLP hints can boost commands that rank lower in pure BM25F
	topK := results
	candidateLimit := options.Limit * 5
	if candidateLimit < 10 {
		candidateLimit = 10
	}
	if len(topK) > candidateLimit {
		topK = topK[:candidateLimit]
	}
	// Run TF-IDF search to get semantic similarity
	tfidfRes := db.tfidf.Search(query, len(topK))
	simByIdx := make(map[int]float64, len(tfidfRes))
	for _, r := range tfidfRes {
		simByIdx[r.CommandIndex] = r.Similarity
	}
	// Blend similarity into scores (small weight to avoid overdominance)
	for i := range topK {
		idx := db.cmdIndex[topK[i].Command]
		if sim, ok := simByIdx[idx]; ok {
			alpha := 0.35
			topK[i].Score += sim * alpha * 100.0
		}
	}
	// Resort after blending
	sort.Slice(topK, func(i, j int) bool { return topK[i].Score > topK[j].Score })
	return topK
}

func (idx *universalIndex) termBM25F(docID int, tf fieldTF) float64 {
	// per-field BM25 sum
	var score float64
	// command
	if tf.cmd > 0 {
		score += idx.fieldBM25(float64(tf.cmd), float64(idx.docLens[docID].cmd), idx.avgLen.cmd, idx.params.w.cmd, idx.params.b.cmd)
	}
	// description
	if tf.desc > 0 {
		score += idx.fieldBM25(float64(tf.desc), float64(idx.docLens[docID].desc), idx.avgLen.desc, idx.params.w.desc, idx.params.b.desc)
	}
	// keywords
	if tf.keys > 0 {
		score += idx.fieldBM25(float64(tf.keys), float64(idx.docLens[docID].keys), idx.avgLen.keys, idx.params.w.keys, idx.params.b.keys)
	}
	// tags
	if tf.tags > 0 {
		score += idx.fieldBM25(float64(tf.tags), float64(idx.docLens[docID].tags), idx.avgLen.tags, idx.params.w.tags, idx.params.b.tags)
	}
	return score
}

func (idx *universalIndex) fieldBM25(tf, dl, avgdl, w, b float64) float64 {
	if avgdl <= 0 {
		avgdl = 1
	}
	k1 := idx.params.k1
	norm := (1 - b) + b*(dl/avgdl)
	tfw := w * tf
	return (tfw * (k1 + 1)) / (tfw + k1*norm)
}

func bm25IDF(n, df int) float64 {
	// Okapi BM25 idf with 0.5 adjustments
	return math.Log((float64(n)-float64(df)+0.5)/(float64(df)+0.5) + 1)
}

func isPlatformCompatible(platforms []string, current string) bool {
	for _, p := range platforms {
		if strings.EqualFold(p, "cross-platform") || strings.EqualFold(p, current) {
			return true
		}
		if checkPlatformVariant(p, current) {
			return true
		}
	}
	return false
}

func checkPlatformVariant(p, current string) bool {
	pLower := strings.ToLower(p)
	switch current {
	case constants.PlatformWindows:
		if pLower == "cmd" || pLower == "powershell" || pLower == "windows-cmd" ||
			pLower == "windows-powershell" || strings.HasPrefix(pLower, constants.PlatformWindows) {
			return true
		}
	case constants.PlatformMacOS:
		if pLower == "darwin" || strings.HasPrefix(pLower, "macos") {
			return true
		}
	case constants.PlatformLinux:
		if pLower == "unix" || pLower == "bash" || pLower == "zsh" ||
			strings.HasPrefix(pLower, "linux") {
			return true
		}
	}
	return false
}

func containsAnyLocal(s string, words []string) bool {
	if len(words) == 0 || s == "" {
		return false
	}
	for _, w := range words {
		if w == "" {
			continue
		}
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

// applySemanticBoost blends embedding similarity into BM25F scores.
// This uses pre-computed command embeddings and GloVe word vectors to add
// semantic understanding to purely lexical search results.
func (db *Database) applySemanticBoost(results []SearchResult, query string) []SearchResult {
	// Compute query embedding
	queryEmbed := db.EmbedQuery(query)
	if queryEmbed == nil {
		return results // No valid embedding for query
	}

	// Get semantic similarity for all commands
	semanticScores := db.SemanticScores(queryEmbed)
	if semanticScores == nil {
		return results
	}

	// Build lookup from result's Command to its index in db.Commands
	// We need this to find the right semantic score for each result
	cmdToIdx := db.cmdIndex
	if cmdToIdx == nil {
		// Rebuild if not available (shouldn't happen if BuildUniversalIndex was called)
		cmdToIdx = make(map[*Command]int, len(db.Commands))
		for i := range db.Commands {
			cmdToIdx[&db.Commands[i]] = i
		}
	}

	// Apply semantic boost to each result
	for i := range results {
		idx, ok := cmdToIdx[results[i].Command]
		if !ok || idx >= len(semanticScores) {
			continue
		}

		similarity := semanticScores[idx]

		// Only apply boost if similarity exceeds minimum threshold
		if similarity >= constants.SemanticMinScore {
			results[i].Score *= (1.0 + constants.SemanticAlpha*similarity)
		}
	}

	// Re-sort after applying semantic boost
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

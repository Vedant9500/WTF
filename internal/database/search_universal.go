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
	// character n-gram -> doc IDs (command+keywords channel)
	charPostings map[string][]int
	// unique character n-gram count per doc (command+keywords channel)
	docCharGramCount []int

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

type learnedFamilyIndex struct {
	tokenToBases map[string]map[string]int
	baseToTokens map[string]map[string]int
	baseDocFreq  map[string]int
	cmdBaseByDoc []string
	totalDocs    int
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

func (p bm25fParams) withOverrides(overrides *BM25Overrides) bm25fParams {
	if overrides == nil {
		return p
	}
	if overrides.K1 != nil && *overrides.K1 > 0 {
		p.k1 = *overrides.K1
	}
	if overrides.MinIDF != nil && *overrides.MinIDF >= 0 {
		p.minIDF = *overrides.MinIDF
	}
	if overrides.B != nil {
		p.b = docLensF{
			cmd:  clamp01(overrides.B.Cmd),
			desc: clamp01(overrides.B.Desc),
			keys: clamp01(overrides.B.Keys),
			tags: clamp01(overrides.B.Tags),
		}
	}
	if overrides.W != nil {
		if overrides.W.Cmd > 0 {
			p.w.cmd = overrides.W.Cmd
		}
		if overrides.W.Desc > 0 {
			p.w.desc = overrides.W.Desc
		}
		if overrides.W.Keys > 0 {
			p.w.keys = overrides.W.Keys
		}
		if overrides.W.Tags > 0 {
			p.w.tags = overrides.W.Tags
		}
	}
	return p
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func appendAdjacentBigrams(tokens []string) []string {
	if len(tokens) < 2 {
		return tokens
	}
	out := make([]string, 0, len(tokens)*2)
	out = append(out, tokens...)
	for i := 0; i+1 < len(tokens); i++ {
		left := tokens[i]
		right := tokens[i+1]
		if left == "" || right == "" {
			continue
		}
		out = append(out, left+"_"+right)
	}
	return out
}

func uniqueCharNGrams(text string, n int) map[string]struct{} {
	if n <= 0 {
		return nil
	}
	compact := strings.ReplaceAll(strings.ToLower(text), " ", "")
	runes := []rune(compact)
	if len(runes) < n {
		return nil
	}
	out := make(map[string]struct{}, len(runes)-n+1)
	for i := 0; i+n <= len(runes); i++ {
		gram := string(runes[i : i+n])
		out[gram] = struct{}{}
	}
	return out
}

// selectTopTerms keeps the most informative terms by IDF to avoid noise from long queries.
// Preserves original query terms to ensure important specific terms aren't lost.
func (db *Database) selectTopTerms(terms []string, maxTerms int) []string {
	if maxTerms <= 0 || len(terms) <= maxTerms || db.uIndex == nil {
		return terms
	}

	longQuery := len(terms) >= constants.LongQueryTermThreshold

	// Conservative for short queries, stricter for verbose queries.
	preserveCount := utils.Min(4, len(terms))
	if longQuery {
		preserveCount = utils.Min(constants.LongQueryPreserveOriginalTerms, len(terms))
	}

	list := db.scoreTerms(terms, preserveCount, longQuery)

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

func (db *Database) scoreTerms(terms []string, preserveCount int, longQuery bool) []termWithScore {
	idx := db.uIndex
	seen := map[string]bool{}
	list := make([]termWithScore, 0, len(terms))
	preserved := 0

	for i, t := range terms {
		if seen[t] {
			continue
		}
		seen[t] = true
		df, ok := idx.df[t]
		if !ok || df == 0 {
			// Unknown terms are often noise in long conversational prompts.
			if i < preserveCount && (!longQuery || looksLikeStructuredEntity(t)) {
				list = append(list, termWithScore{term: t, idf: 1.0, isOriginal: true})
				if longQuery {
					preserved++
				}
			}
			continue
		}
		idf := bm25IDF(idx.N, df)
		if longQuery {
			idf *= longQueryTermImportance(t)
		}

		isOriginal := false
		if longQuery {
			if preserved < preserveCount && isLongQueryPreserveCandidate(t) {
				isOriginal = true
				preserved++
			}
		} else {
			isOriginal = i < preserveCount
		}

		list = append(list, termWithScore{term: t, idf: idf, isOriginal: isOriginal})
	}
	return list
}

func longQueryTermImportance(term string) float64 {
	importance := 1.0

	if isGenericQueryVerb(term) {
		importance *= 0.35
	}
	if isLowSignalLongQueryTerm(term) {
		importance *= 0.45
	}
	if looksLikeStructuredEntity(term) {
		importance *= 1.35
	}
	if isLongQueryAnchorLexeme(term) {
		importance *= 1.20
	}

	return importance
}

func isLongQueryPreserveCandidate(term string) bool {
	if looksLikeStructuredEntity(term) || isLongQueryAnchorLexeme(term) {
		return true
	}
	if isGenericQueryVerb(term) || isLowSignalLongQueryTerm(term) {
		return false
	}
	return len(term) >= 3
}

func isLowSignalLongQueryTerm(term string) bool {
	switch term {
	case "me", "current", "project", "every", "each", "custom", "name", "under", "sorted", "should", "which", "matching", "lines":
		return true
	default:
		return false
	}
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
		postings:         make(map[string][]posting),
		charPostings:     make(map[string][]int),
		df:               make(map[string]int),
		N:                len(db.Commands),
		params:           defaultParams(),
		docCharGramCount: make([]int, len(db.Commands)),
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

		charField := db.Commands[i].CommandLower + " " + strings.Join(db.Commands[i].KeywordsLower, " ")
		grams := uniqueCharNGrams(charField, 3)
		idx.docCharGramCount[i] = len(grams)
		for gram := range grams {
			idx.charPostings[gram] = append(idx.charPostings[gram], i)
		}

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
	db.buildLearnedFamilyIndex()
}

func (db *Database) buildLearnedFamilyIndex() {
	idx := &learnedFamilyIndex{
		tokenToBases: make(map[string]map[string]int),
		baseToTokens: make(map[string]map[string]int),
		baseDocFreq:  make(map[string]int),
		cmdBaseByDoc: make([]string, len(db.Commands)),
		totalDocs:    len(db.Commands),
	}

	for i := range db.Commands {
		cmd := &db.Commands[i]
		base := getCommandBase(strings.ToLower(cmd.Command))
		idx.cmdBaseByDoc[i] = base
		idx.baseDocFreq[base]++

		tokens := make(map[string]bool)
		expansionTokens := make(map[string]bool)

		for _, t := range normalizeAndTokenize(cmd.CommandLower) {
			tokens[t] = true
			expansionTokens[t] = true
		}

		for _, t := range normalizeAndTokenize(cmd.DescriptionLower) {
			tokens[t] = true
		}
		for _, t := range normalizeAndTokenize(strings.Join(cmd.KeywordsLower, " ")) {
			tokens[t] = true
			expansionTokens[t] = true
		}
		for _, t := range normalizeAndTokenize(strings.Join(cmd.TagsLower, " ")) {
			tokens[t] = true
		}

		for tok := range tokens {
			if idx.tokenToBases[tok] == nil {
				idx.tokenToBases[tok] = make(map[string]int)
			}
			idx.tokenToBases[tok][base]++
		}

		if idx.baseToTokens[base] == nil {
			idx.baseToTokens[base] = make(map[string]int)
		}
		for tok := range expansionTokens {
			idx.baseToTokens[base][tok]++
		}
	}

	db.familyPriorIndex = idx
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

	options = normalizeSearchOptions(options)

	// Try embedding-based search first if enabled and available
	if options.UseEmbedding && db.enhancedEmbeddingIndex != nil && db.embeddingSearcher != nil {
		if results := db.tryEmbeddingSearch(query, options); len(results) > 0 {
			return results
		}
	}

	// Fall back to BM25F search
	return db.searchUniversalBM25F(query, options)
}

// tryEmbeddingSearch attempts embedding-based search and returns results if successful.
func (db *Database) tryEmbeddingSearch(query string, options SearchOptions) []SearchResult {
	if db.embeddingSearcher == nil {
		return nil
	}

	// Perform embedding search
	embeddingResults := db.embeddingSearcher.Search(query, options.Limit*2)
	if len(embeddingResults) == 0 {
		return nil
	}

	// Convert to standard search results
	results := make([]SearchResult, 0, len(embeddingResults))
	for _, er := range embeddingResults {
		if er.CommandIndex < len(db.Commands) {
			cmd := db.Commands[er.CommandIndex]
			
			// Apply platform/pipeline filters if needed
			if !db.matchesSearchOptions(&cmd, options) {
				continue
			}
			
			results = append(results, SearchResult{
				Command: &cmd,
				Score:   er.Score,
			})
		}
	}

	// If we also have BM25F, fuse the results
	if db.uIndex != nil && len(results) > 0 {
		bm25Results := db.searchUniversalBM25F(query, options)
		if len(bm25Results) > 0 {
			return db.hybridFuseResults(embeddingResults, bm25Results, options.Limit)
		}
	}

	// Sort and limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

	return results
}

// matchesSearchOptions checks if a command matches the search options (platform/pipeline filters).
func (db *Database) matchesSearchOptions(cmd *Command, options SearchOptions) bool {
	// Check pipeline filter
	if options.PipelineOnly && !cmd.Pipeline {
		return false
	}

	// Check platform filter
	if !options.AllPlatforms && len(options.Platforms) > 0 {
		matches := false
		for _, platform := range options.Platforms {
			for _, cmdPlatform := range cmd.Platform {
				if strings.EqualFold(platform, cmdPlatform) {
					matches = true
					break
				}
			}
			if matches {
				break
			}
		}
		if !matches {
			return false
		}
	}

	return true
}

// searchUniversalBM25F performs the original BM25F search.
func (db *Database) searchUniversalBM25F(query string, options SearchOptions) []SearchResult {
	if db.uIndex == nil || db.uIndex.N != len(db.Commands) {
		// (Re)build lazily if needed
		db.BuildUniversalIndex()
	}

	options = normalizeSearchOptions(options)

	terms := normalizeAndTokenize(query)
	var pq *nlp.ProcessedQuery

	// NLP enhancement
	if options.UseNLP {
		pq, terms = db.enhanceQueryWithNLP(query, terms)
	}

	// Dual-path query handling: preserve short-query behavior and normalize
	// verbose long queries into compact intent-bearing terms before BM25F.
	terms = normalizeLongQueryTermsForScoring(terms, pq)
	if !options.DisableBigrams {
		terms = appendAdjacentBigrams(terms)
	}

	// If no terms after processing, try fuzzy search as fallback
	if len(terms) == 0 {
		return db.handleNoTermsFallback(query, options)
	}

	// Reduce noise for long queries by keeping top-IDF terms
	termsCap := resolveTopTermsCap(options)
	terms = db.selectTopTerms(terms, termsCap)

	// Calculate initial scores using BM25F
	scores := db.calculateInitialScores(terms, pq, options)
	db.mergeFamilyExpansionCandidates(scores, terms, pq, options)
	db.mergeHybridCandidates(scores, query, options)

	// If no BM25F results, try fuzzy search as fallback for typos
	if len(scores) == 0 {
		return db.handleNoTermsFallback(query, options)
	}

	// Convert to results and apply pipeline boosts
	results := db.collectResults(scores, query, terms, pq, options)

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

func normalizeSearchOptions(options SearchOptions) SearchOptions {
	if options.Limit <= 0 {
		options.Limit = 10
	}
	if options.FamilyExpansionClarityMax <= 0 {
		options.FamilyExpansionClarityMax = constants.FamilyExpansionClarityMaxDefault
	}
	if options.FamilyExpansionBlendWeight <= 0 {
		options.FamilyExpansionBlendWeight = constants.FamilyExpansionBlendWeightDefault
	}
	return options
}

func familyQueryClarity(baseScores map[string]float64) float64 {
	if len(baseScores) == 0 {
		return 1.0
	}

	total := 0.0
	for _, score := range baseScores {
		if score > 0 {
			total += score
		}
	}
	if total <= 0 {
		return 1.0
	}
	if len(baseScores) == 1 {
		return 1.0
	}

	entropy := 0.0
	for _, score := range baseScores {
		if score <= 0 {
			continue
		}
		p := score / total
		entropy -= p * math.Log(p)
	}

	maxEntropy := math.Log(float64(len(baseScores)))
	if maxEntropy <= 0 {
		return 1.0
	}
	normalizedEntropy := entropy / maxEntropy
	if normalizedEntropy < 0 {
		normalizedEntropy = 0
	}
	if normalizedEntropy > 1 {
		normalizedEntropy = 1
	}

	return 1.0 - normalizedEntropy
}

func topFamilyBases(baseScores map[string]float64, maxBases int) []string {
	type familyScore struct {
		base  string
		score float64
	}
	list := make([]familyScore, 0, len(baseScores))
	for base, score := range baseScores {
		if score <= 0 {
			continue
		}
		list = append(list, familyScore{base: base, score: score})
	}
	if len(list) == 0 {
		return nil
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].score == list[j].score {
			return list[i].base < list[j].base
		}
		return list[i].score > list[j].score
	})

	if maxBases > 0 && len(list) > maxBases {
		list = list[:maxBases]
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		out = append(out, item.base)
	}
	return out
}

func (db *Database) handleNoTermsFallback(query string, options SearchOptions) []SearchResult {
	if options.UseFuzzy {
		return db.performFuzzySearch(query, options)
	}
	return nil
}

func resolveTopTermsCap(options SearchOptions) int {
	if options.TopTermsCap > 0 {
		return options.TopTermsCap
	}
	return 10
}

func (db *Database) mergeFamilyExpansionCandidates(scores map[int]float64, terms []string, pq *nlp.ProcessedQuery, options SearchOptions) {
	allowed := db.resolveFamilyExpansionAllowedBases(terms, pq, options)
	if len(allowed) == 0 {
		return
	}

	blend := resolveFamilyExpansionBlendWeight(options)
	currentPlatform := getCurrentPlatform()
	for docID := range db.Commands {
		baseScore, include := db.familyExpansionBaseScore(docID, allowed)
		if !include {
			continue
		}

		cmd := &db.Commands[docID]
		if !shouldIncludeFamilyExpansionCommand(cmd, options, currentPlatform) {
			continue
		}

		bonus := familyExpansionBonus(baseScore, blend)
		scores[docID] += bonus
	}
}

func (db *Database) resolveFamilyExpansionAllowedBases(terms []string, pq *nlp.ProcessedQuery, options SearchOptions) map[string]float64 {
	if !options.EnableFamilyExpansion || db.familyPriorIndex == nil || len(terms) == 0 {
		return nil
	}
	if len(terms) < constants.LongQueryNormalizationThreshold {
		return nil
	}

	baseScores := db.estimateQueryFamilyScores(terms, pq)
	if len(baseScores) == 0 {
		return nil
	}
	if familyQueryClarity(baseScores) > options.FamilyExpansionClarityMax {
		return nil
	}

	maxBases := options.FamilyExpansionMaxBases
	if maxBases <= 0 {
		maxBases = constants.FamilyExpansionTopBasesDefault
	}

	topBases := topFamilyBases(baseScores, maxBases)
	if len(topBases) == 0 {
		return nil
	}

	allowed := make(map[string]float64, len(topBases))
	for _, base := range topBases {
		allowed[base] = baseScores[base]
	}
	return allowed
}

func resolveFamilyExpansionBlendWeight(options SearchOptions) float64 {
	blend := options.FamilyExpansionBlendWeight
	if blend <= 0 {
		return constants.FamilyExpansionBlendWeightDefault
	}
	return blend
}

func (db *Database) familyExpansionBaseScore(docID int, allowed map[string]float64) (float64, bool) {
	base := db.familyPriorIndex.cmdBaseByDoc[docID]
	baseScore, ok := allowed[base]
	if !ok || baseScore <= 0 {
		return 0, false
	}
	return baseScore, true
}

func shouldIncludeFamilyExpansionCommand(cmd *Command, options SearchOptions, currentPlatform string) bool {
	if !matchesPlatformOptions(cmd, options, currentPlatform) {
		return false
	}
	if options.PipelineOnly && !isPipelineCommand(cmd) {
		return false
	}
	return true
}

func familyExpansionBonus(baseScore, blend float64) float64 {
	bonus := constants.FamilyExpansionMaxContribution * baseScore * blend
	if bonus > constants.FamilyExpansionMaxContribution {
		return constants.FamilyExpansionMaxContribution
	}
	return bonus
}

func (db *Database) mergeHybridCandidates(scores map[int]float64, query string, options SearchOptions) {
	if !options.DisableCharNGram {
		charScores := db.calculateCharNGramCandidateScores(query, options)
		for docID, ngramScore := range charScores {
			if cur, ok := scores[docID]; ok {
				scores[docID] = cur + ngramScore*0.35
				continue
			}
			scores[docID] = ngramScore * 0.80
		}
	}

	// Merge scalable semantic candidates (no command-specific hinting) to improve
	// lexical miss recovery on natural language phrasing.
	if options.UseNLP && db.HasEmbeddings() {
		semanticScores := db.calculateSemanticCandidateScores(query, options)
		for docID, semScore := range semanticScores {
			if cur, ok := scores[docID]; ok {
				scores[docID] = cur + semScore*0.35
				continue
			}
			scores[docID] = semScore * 0.55
		}
	}

	if options.UseNLP && db.tfidf != nil {
		tfidfScores := db.calculateTFIDFCandidateScores(query, options)
		for docID, tfidfScore := range tfidfScores {
			if cur, ok := scores[docID]; ok {
				scores[docID] = cur + tfidfScore*0.45
				continue
			}
			scores[docID] = tfidfScore * 0.65
		}
	}
}

func (db *Database) calculateCharNGramCandidateScores(query string, options SearchOptions) map[int]float64 {
	idx := db.uIndex
	if idx == nil || len(idx.charPostings) == 0 {
		return nil
	}

	queryNGrams := uniqueCharNGrams(query, 3)
	if len(queryNGrams) == 0 {
		return nil
	}

	overlapByDoc := overlapCountsByDoc(idx, queryNGrams)
	if len(overlapByDoc) == 0 {
		return nil
	}

	currentPlatform := getCurrentPlatform()
	maxCandidates := options.Limit * constants.ResultsBufferMultiplier * 2
	if maxCandidates < 12 {
		maxCandidates = 12
	}

	type cand struct {
		docID int
		score float64
	}
	list := make([]cand, 0, len(overlapByDoc))
	qNorm := math.Sqrt(float64(len(queryNGrams)))

	for docID, overlap := range overlapByDoc {
		sim, ok := db.charNGramSimilarityForDoc(docID, overlap, qNorm, idx, options, currentPlatform)
		if !ok {
			continue
		}
		list = append(list, cand{docID: docID, score: sim})
	}

	if len(list) == 0 {
		return nil
	}

	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })
	if len(list) > maxCandidates {
		list = list[:maxCandidates]
	}

	out := make(map[int]float64, len(list))
	for _, c := range list {
		out[c.docID] = 55.0 * c.score
	}

	return out
}

func overlapCountsByDoc(idx *universalIndex, queryNGrams map[string]struct{}) map[int]int {
	overlapByDoc := make(map[int]int)
	for gram := range queryNGrams {
		for _, docID := range idx.charPostings[gram] {
			overlapByDoc[docID]++
		}
	}
	return overlapByDoc
}

func (db *Database) charNGramSimilarityForDoc(
	docID, overlap int,
	qNorm float64,
	idx *universalIndex,
	options SearchOptions,
	currentPlatform string,
) (float64, bool) {
	if docID < 0 || docID >= len(db.Commands) {
		return 0, false
	}

	doc := &db.Commands[docID]
	if !matchesPlatformOptions(doc, options, currentPlatform) {
		return 0, false
	}
	if options.PipelineOnly && !isPipelineCommand(doc) {
		return 0, false
	}

	dNorm := math.Sqrt(float64(idx.docCharGramCount[docID]))
	if dNorm == 0 || qNorm == 0 {
		return 0, false
	}

	sim := float64(overlap) / (qNorm * dNorm)
	if sim < 0.12 {
		return 0, false
	}

	return sim, true
}

func (db *Database) calculateSemanticCandidateScores(query string, options SearchOptions) map[int]float64 {
	queryEmbed := db.EmbedQuery(query)
	if queryEmbed == nil {
		return nil
	}

	allSemantic := db.SemanticScores(queryEmbed)
	if len(allSemantic) == 0 {
		return nil
	}

	currentPlatform := getCurrentPlatform()
	maxCandidates := options.Limit * constants.ResultsBufferMultiplier
	if maxCandidates < 10 {
		maxCandidates = 10
	}

	type cand struct {
		docID int
		score float64
	}
	list := make([]cand, 0, maxCandidates*2)

	for i, sim := range allSemantic {
		if i >= len(db.Commands) {
			break
		}
		if sim < constants.SemanticMinScore {
			continue
		}
		cmd := &db.Commands[i]
		if !matchesPlatformOptions(cmd, options, currentPlatform) {
			continue
		}
		if options.PipelineOnly && !isPipelineCommand(cmd) {
			continue
		}

		list = append(list, cand{docID: i, score: sim})
	}

	if len(list) == 0 {
		return nil
	}

	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })
	if len(list) > maxCandidates {
		list = list[:maxCandidates]
	}

	out := make(map[int]float64, len(list))
	for _, c := range list {
		// Scale semantic similarity into BM25-compatible range.
		out[c.docID] = 45.0 * c.score
	}

	return out
}

func (db *Database) calculateTFIDFCandidateScores(query string, options SearchOptions) map[int]float64 {
	if db.tfidf == nil {
		return nil
	}

	currentPlatform := getCurrentPlatform()
	maxCandidates := options.Limit * constants.ResultsBufferMultiplier * 2
	if maxCandidates < 15 {
		maxCandidates = 15
	}

	tfidfRes := db.tfidf.Search(query, maxCandidates)
	if len(tfidfRes) == 0 {
		return nil
	}

	out := make(map[int]float64, len(tfidfRes))
	for _, r := range tfidfRes {
		if r.CommandIndex < 0 || r.CommandIndex >= len(db.Commands) {
			continue
		}
		cmd := &db.Commands[r.CommandIndex]
		if !matchesPlatformOptions(cmd, options, currentPlatform) {
			continue
		}
		if options.PipelineOnly && !isPipelineCommand(cmd) {
			continue
		}
		out[r.CommandIndex] = 100.0 * r.Similarity
	}

	return out
}

func (db *Database) calculateInitialScores(terms []string, pq *nlp.ProcessedQuery, options SearchOptions) map[int]float64 {
	idx := db.uIndex
	params := idx.params.withOverrides(options.BM25Overrides)
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
		if idf < params.minIDF {
			continue
		}
		boost := 1.0
		if b, ok := termBoost[term]; ok && b > 0 {
			boost = b
		}
		db.processPostingsForTerm(postings, idx, params, idf, boost, scores, currentPlatform, options)
	}
	return scores
}

func (db *Database) processPostingsForTerm(
	postings []posting,
	idx *universalIndex,
	params bm25fParams,
	idf, boost float64,
	scores map[int]float64,
	currentPlatform string,
	options SearchOptions,
) {
	for _, p := range postings {
		doc := &db.Commands[p.docID]

		// Platform filtering (skip only if command does not match requested platform behavior)
		if !matchesPlatformOptions(doc, options, currentPlatform) {
			continue
		}

		// Pipeline filtering
		if options.PipelineOnly && !isPipelineCommand(doc) {
			continue
		}

		s := scores[p.docID]
		s += (idf * boost) * idx.termBM25F(p.docID, p.tf, params)
		scores[p.docID] = s
	}
}

func matchesPlatformOptions(cmd *Command, options SearchOptions, currentPlatform string) bool {
	if options.AllPlatforms {
		return true
	}

	selectedPlatforms := options.Platforms
	if len(selectedPlatforms) == 0 {
		selectedPlatforms = []string{currentPlatform}
	}

	for _, platform := range selectedPlatforms {
		if options.NoCrossPlatform {
			if isPlatformCompatibleWithoutCross(cmd.Platform, platform) {
				return true
			}
			continue
		}

		if isPlatformCompatible(cmd.Platform, platform) {
			return true
		}
	}

	if options.NoCrossPlatform {
		return false
	}

	// Fallback to cross-platform commands unless explicitly disabled.
	if isPlatformCompatible(cmd.Platform, "cross-platform") || isCrossPlatformTool(cmd.Command) {
		return true
	}

	return false
}

func isPlatformCompatibleWithoutCross(platforms []string, requested string) bool {
	if len(platforms) == 0 {
		return true
	}

	for _, p := range platforms {
		if strings.EqualFold(p, "cross-platform") {
			continue
		}
		if strings.EqualFold(p, requested) || checkPlatformVariant(p, requested) {
			return true
		}
	}
	return false
}

func (db *Database) enhanceQueryWithNLP(query string, terms []string) (pq *nlp.ProcessedQuery, enhancedTerms []string) {
	processor := nlp.NewQueryProcessor()
	pq = processor.ProcessQuery(query)
	longQuery := len(terms) >= constants.LongQueryTermThreshold

	enh := pq.GetEnhancedKeywords()

	// Add relevant enhanced terms that aren't already present
	if len(enh) > 0 {
		for _, enhTerm := range enh {
			if longQuery && !isLongQueryExpansionTerm(enhTerm) {
				continue
			}
			found := false
			for _, origTerm := range terms {
				if origTerm == enhTerm {
					found = true
					break
				}
			}
			limit := 8
			if longQuery {
				limit = 10
			}
			if !found && len(terms) < limit {
				terms = append(terms, enhTerm)
			}
		}
	}
	return pq, terms
}

func normalizeLongQueryTermsForScoring(terms []string, pq *nlp.ProcessedQuery) []string {
	if len(terms) < constants.LongQueryNormalizationThreshold {
		return terms
	}

	if pq != nil {
		if core := buildLongQueryCoreTerms(terms, pq); len(core) >= 3 {
			return core
		}
	}

	seen := make(map[string]bool)
	filtered := make([]string, 0, len(terms))
	for _, t := range terms {
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		if isGenericQueryVerb(t) || isLowSignalLongQueryTerm(t) {
			continue
		}
		filtered = append(filtered, t)
	}

	if len(filtered) >= 3 {
		return filtered
	}

	return terms
}

func buildLongQueryCoreTerms(originalTerms []string, pq *nlp.ProcessedQuery) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, 10)
	add := func(term string) {
		if term == "" || seen[term] || len(out) >= 10 {
			return
		}
		seen[term] = true
		out = append(out, term)
	}

	// 1. Anchor terms from original query (highest signal)
	for _, t := range originalTerms {
		if looksLikeStructuredEntity(t) || isLongQueryAnchorLexeme(t) {
			add(t)
		}
	}

	// 2. Intent-bearing terms from NLP analysis
	if pq != nil {
		addNLPCoreTerms(pq, add)
	}

	// 3. Fallback expansion terms from original query
	for _, t := range originalTerms {
		if isLongQueryExpansionTerm(t) {
			add(t)
		}
	}

	return out
}

func addNLPCoreTerms(pq *nlp.ProcessedQuery, add func(string)) {
	// Structured/Anchor keywords
	for _, t := range pq.Keywords {
		if looksLikeStructuredEntity(t) || isLongQueryAnchorLexeme(t) {
			add(t)
		}
	}
	// Action verbs (filtered)
	for _, t := range pq.Actions {
		if !isGenericQueryVerb(t) && isLongQueryExpansionTerm(t) {
			add(t)
		}
	}
	// Target nouns (filtered)
	for _, t := range pq.Targets {
		if !isLowSignalTargetTerm(t) && isLongQueryExpansionTerm(t) {
			add(t)
		}
	}
	// All other keywords
	for _, t := range pq.Keywords {
		if isLongQueryExpansionTerm(t) {
			add(t)
		}
	}
}

func isLowSignalTargetTerm(term string) bool {
	switch term {
	case "directory", "directories", "folder", "folders", "path", "location", "content", "contents":
		return true
	default:
		return false
	}
}

func isLongQueryExpansionTerm(term string) bool {
	if term == "" {
		return false
	}
	if looksLikeStructuredEntity(term) || isLongQueryAnchorLexeme(term) {
		return true
	}
	if isGenericQueryVerb(term) || isLowSignalLongQueryTerm(term) {
		return false
	}
	return len(term) >= 4
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
	cmdTokens = appendAdjacentBigrams(cmdTokens)
	descTokens := normalizeAndTokenize(descText)
	keysTokens := make([]string, 0)
	if len(cmd.KeywordsLower) > 0 {
		keysTokens = normalizeAndTokenize(strings.Join(cmd.KeywordsLower, " "))
	} else if len(cmd.Keywords) > 0 {
		keysTokens = normalizeAndTokenize(strings.Join(cmd.Keywords, " "))
	}
	keysTokens = appendAdjacentBigrams(keysTokens)
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

func (db *Database) collectResults(
	scores map[int]float64,
	_ string,
	terms []string,
	pq *nlp.ProcessedQuery,
	options SearchOptions,
) []SearchResult {
	results := make([]SearchResult, 0, utils.Min(len(scores), options.Limit*3))
	for docID, score := range scores {
		cmd := &db.Commands[docID]

		if !options.DisableProximity {
			score *= calculateDescriptionProximityBoost(cmd, terms)
		}

		// Apply IDF-weighted query coverage so commands matching more query intent terms
		// are favored over commands that only match one dominant token.
		score *= db.calculateQueryCoverageBoost(cmd, terms, pq)

		// Apply data-driven family prior learned from command corpus and weighted
		// by query entities/targets instead of fixed command maps.
		score *= db.calculateLearnedFamilyBoost(docID, terms, pq)

		// Additional universal feature weighting for verbose queries where
		// users express concrete intent like download or disk usage.
		score *= db.calculateLongQueryIntentFeatureBoost(cmd, terms, pq)

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

func calculateDescriptionProximityBoost(cmd *Command, terms []string) float64 {
	descTerms, queryTerms, ok := proximityInputs(cmd, terms)
	if !ok {
		return 1.0
	}

	matchedCount, minDist := minProximityDistance(descTerms, queryTerms)
	if matchedCount < 2 || minDist == math.MaxInt {
		return 1.0
	}

	coverageDenom := utils.Min(4, len(queryTerms))
	if coverageDenom <= 0 {
		return 1.0
	}
	coverage := float64(matchedCount) / float64(coverageDenom)
	if coverage > 1.0 {
		coverage = 1.0
	}

	// Close co-occurrence in description signals tighter intent fit for verbose queries.
	proximity := 1.0 / (1.0 + float64(minDist))
	boost := 1.0 + 0.20*coverage*proximity
	if boost > 1.10 {
		return 1.10
	}
	return boost
}

func proximityInputs(cmd *Command, terms []string) (descTerms, queryTerms []string, ok bool) {
	if cmd == nil {
		return nil, nil, false
	}
	descText := cmd.DescriptionLower
	if descText == "" {
		descText = strings.ToLower(cmd.Description)
	}
	descTerms = normalizeAndTokenize(descText)
	queryTerms = uniqueProximityTerms(terms)
	if len(descTerms) < 4 || len(queryTerms) < 2 {
		return nil, nil, false
	}
	return descTerms, queryTerms, true
}

func minProximityDistance(descTerms, queryTerms []string) (matchedCount, minDist int) {
	querySet := make(map[string]struct{}, len(queryTerms))
	for _, t := range queryTerms {
		querySet[t] = struct{}{}
	}

	matchedTerms := make(map[string]struct{}, len(queryTerms))
	lastPos := make(map[string]int, len(queryTerms))
	minDist = math.MaxInt

	for pos, tok := range descTerms {
		if _, ok := querySet[tok]; !ok {
			continue
		}
		matchedTerms[tok] = struct{}{}
		for other, otherPos := range lastPos {
			if other == tok {
				continue
			}
			dist := pos - otherPos
			if dist < 0 {
				dist = -dist
			}
			if dist < minDist {
				minDist = dist
			}
		}
		lastPos[tok] = pos
	}

	return len(matchedTerms), minDist
}

func uniqueProximityTerms(terms []string) []string {
	seen := make(map[string]struct{}, len(terms))
	out := make([]string, 0, len(terms))
	for _, t := range terms {
		if t == "" || strings.Contains(t, "_") {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func (db *Database) calculateLongQueryIntentFeatureBoost(cmd *Command, terms []string, pq *nlp.ProcessedQuery) float64 {
	if len(terms) < constants.LongQueryNormalizationThreshold {
		return 1.0
	}

	intents := detectLongQueryIntents(terms, pq)
	if !intents.any() {
		return 1.0
	}

	text := buildCommandSearchText(cmd)
	positive, negative := calculateIntentMatchScores(text, intents)

	raw := 1.0 + constants.LongQueryIntentBoostAlpha*(float64(positive)-constants.LongQueryIntentNegativeWeight*float64(negative))
	if raw < constants.LongQueryIntentMinBoost {
		return constants.LongQueryIntentMinBoost
	}
	if raw > constants.LongQueryIntentMaxBoost {
		return constants.LongQueryIntentMaxBoost
	}
	return raw
}

type longQueryIntents struct {
	download      bool
	disk          bool
	recursiveText bool
	processPort   bool
}

func (i longQueryIntents) any() bool {
	return i.download || i.disk || i.recursiveText || i.processPort
}

func detectLongQueryIntents(terms []string, pq *nlp.ProcessedQuery) longQueryIntents {
	downloadCues := []string{"download", "fetch", "url", "http", "https", "file", "save", "wget", "curl"}
	diskCues := []string{"disk", "usage", "folder", "directory", "size", "sorted", "space", "du", "df"}
	recursiveTextCues := []string{
		"recursive", "recursively", "text", "files", "file", "timeout",
		"error", "grep", constants.Search, "find", "replace", "json",
	}
	processPortCues := []string{"process", "port", "listening", "listen", "socket", "windows", "8080", "netstat", "lsof", "ss"}

	return longQueryIntents{
		download:      countLongQueryIntentCues(terms, pq, downloadCues) >= 2,
		disk:          countLongQueryIntentCues(terms, pq, diskCues) >= 2,
		recursiveText: countLongQueryIntentCues(terms, pq, recursiveTextCues) >= 2,
		processPort:   countLongQueryIntentCues(terms, pq, processPortCues) >= 2,
	}
}

func buildCommandSearchText(cmd *Command) string {
	cmdText := cmd.CommandLower
	if cmdText == "" {
		cmdText = strings.ToLower(cmd.Command)
	}
	descText := cmd.DescriptionLower
	if descText == "" {
		descText = strings.ToLower(cmd.Description)
	}
	keysText := strings.Join(cmd.KeywordsLower, " ")
	if keysText == "" && len(cmd.Keywords) > 0 {
		keysText = strings.ToLower(strings.Join(cmd.Keywords, " "))
	}
	tagsText := strings.Join(cmd.TagsLower, " ")
	if tagsText == "" && len(cmd.Tags) > 0 {
		tagsText = strings.ToLower(strings.Join(cmd.Tags, " "))
	}
	return cmdText + " " + descText + " " + keysText + " " + tagsText
}

func calculateIntentMatchScores(text string, intents longQueryIntents) (positive, negative int) {
	if intents.download {
		positive += countWordMatches(text, []string{"download", "fetch", "url", "http", "https", "wget", "curl", "save", "output"})
		negative += countWordMatches(text, []string{"tree", "set-location", "cd", "realpath", "path", "directory"})
	}

	if intents.disk {
		positive += countWordMatches(text, []string{"disk", "usage", "du", "df", "size", "space", "folder", "directory", "sort", "sorted"})
		negative += countWordMatches(text, []string{"tree", "set-location", "cd", "realpath", "path", "root"})
	}

	if intents.recursiveText {
		positive += countWordMatches(text, []string{
			"grep", "ripgrep", "rg", "find", "recursive", "recursively",
			"pattern", "regex", constants.Search, "replace", "sed", "awk", "perl",
		})
		negative += countWordMatches(text, []string{"conda", "npm", "gh", "pdf", "tree", "repo", "project"})
	}

	if intents.processPort {
		positive += countWordMatches(text, []string{"netstat", "ss", "lsof", "port", "socket", "listening", "process", "pid", "taskkill"})
		negative += countWordMatches(text, []string{"find", "tree", "package", "manager", "macports"})
	}

	return positive, negative
}

func countLongQueryIntentCues(terms []string, pq *nlp.ProcessedQuery, cues []string) int {
	set := make(map[string]bool, len(cues))
	for _, cue := range cues {
		set[cue] = true
	}

	seen := make(map[string]bool)
	count := 0

	for _, t := range terms {
		if set[t] && !seen[t] {
			seen[t] = true
			count++
		}
	}

	if pq == nil {
		return count
	}

	for _, t := range pq.Actions {
		if set[t] && !seen[t] {
			seen[t] = true
			count++
		}
	}
	for _, t := range pq.Targets {
		if set[t] && !seen[t] {
			seen[t] = true
			count++
		}
	}
	for _, t := range pq.Keywords {
		if set[t] && !seen[t] {
			seen[t] = true
			count++
		}
	}

	return count
}

func countWordMatches(text string, terms []string) int {
	count := 0
	for _, t := range terms {
		if containsWord(text, t) {
			count++
		}
	}
	return count
}

func (db *Database) calculateLearnedFamilyBoost(docID int, terms []string, pq *nlp.ProcessedQuery) float64 {
	if db.familyPriorIndex == nil || docID < 0 || docID >= len(db.familyPriorIndex.cmdBaseByDoc) {
		return 1.0
	}

	longQuery := len(terms) >= constants.LongQueryTermThreshold
	baseScores := db.estimateQueryFamilyScores(terms, pq)
	if len(baseScores) == 0 {
		return 1.0
	}
	if longQuery {
		topScore, margin := familyConfidence(baseScores)
		if topScore < constants.LongQueryLearnedFamilyMinConfidence || margin < constants.LongQueryLearnedFamilyMinMargin {
			return 1.0
		}
	}

	base := db.familyPriorIndex.cmdBaseByDoc[docID]
	score, ok := baseScores[base]
	if !ok || score <= 0 {
		return 1.0
	}

	alpha := constants.LearnedFamilyPriorAlpha
	if longQuery {
		alpha *= constants.LongQueryLearnedFamilyAlphaScale
	}

	return 1.0 + alpha*score
}

func familyConfidence(scores map[string]float64) (top, margin float64) {
	if len(scores) == 0 {
		return 0.0, 0.0
	}

	top = 0.0
	second := 0.0
	for _, s := range scores {
		if s > top {
			second = top
			top = s
			continue
		}
		if s > second {
			second = s
		}
	}

	margin = top - second
	return top, margin
}

func (db *Database) estimateQueryFamilyScores(terms []string, pq *nlp.ProcessedQuery) map[string]float64 {
	if db.familyPriorIndex == nil || len(terms) == 0 {
		return nil
	}
	longQuery := len(terms) >= constants.LongQueryTermThreshold

	scores := make(map[string]float64)
	seen := make(map[string]bool)
	for _, term := range terms {
		if seen[term] {
			continue
		}
		seen[term] = true

		bases, ok := db.familyPriorIndex.tokenToBases[term]
		if !ok {
			continue
		}

		termWeight := db.queryTermWeight(term) * queryEntityWeight(term, pq, longQuery)

		total := 0
		for _, c := range bases {
			total += c
		}
		if total == 0 {
			continue
		}

		for base, c := range bases {
			scores[base] += termWeight * (float64(c) / float64(total))
		}
	}

	return normalizeTopFamilyScores(scores)
}

func normalizeTopFamilyScores(scores map[string]float64) map[string]float64 {
	if len(scores) == 0 {
		return scores
	}

	type kv struct {
		base  string
		score float64
	}
	list := make([]kv, 0, len(scores))
	for b, s := range scores {
		list = append(list, kv{base: b, score: s})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })
	if len(list) > constants.LearnedFamilyTopBases {
		list = list[:constants.LearnedFamilyTopBases]
	}

	maxScore := list[0].score
	if maxScore <= 0 {
		return map[string]float64{}
	}

	out := make(map[string]float64, len(list))
	for _, it := range list {
		out[it.base] = it.score / maxScore
	}
	return out
}

func queryEntityWeight(term string, pq *nlp.ProcessedQuery, longQuery bool) float64 {
	weight := 1.0

	if isGenericQueryVerb(term) {
		if longQuery {
			weight *= constants.LongQueryGenericVerbWeight
		} else {
			weight *= 0.7
		}
	}

	if looksLikeStructuredEntity(term) {
		if longQuery {
			weight *= constants.LongQueryStructuredEntityWeight
		} else {
			weight *= 1.25
		}
	}

	if pq != nil {
		if containsToken(pq.Targets, term) {
			weight *= 1.35
		}
		if containsToken(pq.Actions, term) {
			weight *= 0.85
		}
	}

	return weight
}

func containsToken(tokens []string, term string) bool {
	for _, t := range tokens {
		if t == term {
			return true
		}
	}
	return false
}

func isGenericQueryVerb(term string) bool {
	switch term {
	case "find", constants.Search, "show", "check", "list", "get":
		return true
	default:
		return false
	}
}

func looksLikeStructuredEntity(term string) bool {
	if len(term) == 4 {
		allDigits := true
		for _, r := range term {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return true
		}
	}

	switch term {
	case "kb", "mb", "gb", "tb", "tar", "gz", "zip", "port", "process":
		return true
	default:
		return false
	}
}

func (db *Database) calculateQueryCoverageBoost(cmd *Command, terms []string, pq *nlp.ProcessedQuery) float64 {
	if len(terms) == 0 {
		return 1.0
	}

	metrics := db.calculateCoverageMetrics(cmd, terms, pq)
	if metrics.totalWeight == 0 {
		return 1.0
	}

	coverage := metrics.matchedWeight / metrics.totalWeight
	boost := 0.55 + coverage

	// Apply query-size and intent-based adjustments
	if len(terms) >= 3 {
		if !metrics.hasMaxTermMatch {
			boost *= 0.45
		}
		if coverage < 0.7 {
			boost *= 0.7
		}
	}

	// Apply long-query evidence penalties
	if len(terms) >= constants.LongQueryTermThreshold {
		boost = applyLongQueryEvidencePenalties(boost, metrics)
	}

	return boost
}

type coverageMetrics struct {
	totalWeight           float64
	matchedWeight         float64
	hasMaxTermMatch       bool
	matchedTermCount      int
	strongFieldMatchCount int
	hasAnchorTerm         bool
	hasAnchorMatch        bool
}

func (db *Database) calculateCoverageMetrics(cmd *Command, terms []string, pq *nlp.ProcessedQuery) coverageMetrics {
	cmdText, keysText, descText := getCommandFieldText(cmd)
	var m coverageMetrics
	var maxTermWeight float64

	for _, term := range terms {
		w := db.queryTermWeight(term)
		matchStrength := queryTermFieldMatchStrength(term, cmdText, keysText, descText)
		matched := matchStrength > 0
		isAnchor := isLongQueryAnchorTerm(term, pq)

		if isAnchor {
			m.hasAnchorTerm = true
		}
		if w > maxTermWeight {
			maxTermWeight = w
			m.hasMaxTermMatch = matched
		}
		m.totalWeight += w
		if matched {
			m.matchedTermCount++
			if matchStrength >= 0.9 {
				m.strongFieldMatchCount++
			}
			if isAnchor {
				m.hasAnchorMatch = true
			}
			m.matchedWeight += w * matchStrength
		}
		if matched && w == maxTermWeight {
			m.hasMaxTermMatch = true
		}
	}
	return m
}

func getCommandFieldText(cmd *Command) (cmdText, keysText, descText string) {
	cmdText = cmd.CommandLower
	if cmdText == "" {
		cmdText = strings.ToLower(cmd.Command)
	}
	descText = cmd.DescriptionLower
	if descText == "" {
		descText = strings.ToLower(cmd.Description)
	}
	keysText = strings.Join(cmd.KeywordsLower, " ")
	if keysText == "" && len(cmd.Keywords) > 0 {
		keysText = strings.ToLower(strings.Join(cmd.Keywords, " "))
	}
	return cmdText, keysText, descText
}

func applyLongQueryEvidencePenalties(boost float64, m coverageMetrics) float64 {
	if m.matchedTermCount < constants.LongQueryMinMatchedTerms {
		boost *= constants.LongQueryLowEvidencePenalty
	}
	if m.strongFieldMatchCount < constants.LongQueryMinStrongFieldMatches {
		boost *= constants.LongQueryWeakFieldPenalty
	}
	if m.hasAnchorTerm && !m.hasAnchorMatch {
		boost *= constants.LongQueryNoAnchorPenalty
	}
	return boost
}

func isLongQueryAnchorTerm(term string, pq *nlp.ProcessedQuery) bool {
	if looksLikeStructuredEntity(term) {
		return true
	}
	if pq != nil && containsToken(pq.Targets, term) {
		return true
	}

	return isLongQueryAnchorLexeme(term)
}

func isLongQueryAnchorLexeme(term string) bool {
	switch term {
	case "download", "upload", "url", "http", "https", "port", "process",
		"disk", "usage", "replace", constants.Archive, "compress", "extract",
		"revert", "undo", "windows", "linux", "macos", "grep", constants.Search:
		return true
	default:
		return false
	}
}

func queryTermFieldMatchStrength(term, cmdText, keysText, descText string) float64 {
	if containsWord(cmdText, term) {
		return 1.0
	}
	if containsWord(keysText, term) {
		return 0.9
	}
	if containsWord(descText, term) {
		return 0.65
	}
	return 0.0
}

func (db *Database) queryTermWeight(term string) float64 {
	if db.uIndex == nil || db.uIndex.N == 0 {
		return 1.0
	}
	df, ok := db.uIndex.df[term]
	if !ok || df <= 0 {
		return 1.0
	}
	return 0.5 + bm25IDF(db.uIndex.N, df)
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

func (idx *universalIndex) termBM25F(docID int, tf fieldTF, params bm25fParams) float64 {
	// per-field BM25 sum
	var score float64
	// command
	if tf.cmd > 0 {
		score += idx.fieldBM25(float64(tf.cmd), float64(idx.docLens[docID].cmd), idx.avgLen.cmd, params.w.cmd, params.b.cmd, params.k1)
	}
	// description
	if tf.desc > 0 {
		score += idx.fieldBM25(float64(tf.desc), float64(idx.docLens[docID].desc), idx.avgLen.desc, params.w.desc, params.b.desc, params.k1)
	}
	// keywords
	if tf.keys > 0 {
		score += idx.fieldBM25(float64(tf.keys), float64(idx.docLens[docID].keys), idx.avgLen.keys, params.w.keys, params.b.keys, params.k1)
	}
	// tags
	if tf.tags > 0 {
		score += idx.fieldBM25(float64(tf.tags), float64(idx.docLens[docID].tags), idx.avgLen.tags, params.w.tags, params.b.tags, params.k1)
	}
	return score
}

func (idx *universalIndex) fieldBM25(tf, dl, avgdl, w, b, k1 float64) float64 {
	if avgdl <= 0 {
		avgdl = 1
	}
	norm := (1 - b) + b*(dl/avgdl)
	tfw := w * tf
	return (tfw * (k1 + 1)) / (tfw + k1*norm)
}

func bm25IDF(n, df int) float64 {
	// Okapi BM25 idf with 0.5 adjustments
	return math.Log((float64(n)-float64(df)+0.5)/(float64(df)+0.5) + 1)
}

func isPlatformCompatible(platforms []string, current string) bool {
	if len(platforms) == 0 {
		return true
	}

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

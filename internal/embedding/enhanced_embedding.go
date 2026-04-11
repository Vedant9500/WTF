// Package embedding provides enhanced embedding search with contextual pooling and ANN indexing.
package embedding

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

const (
	// Enhanced embedding format version
	EnhancedEmbeddingVersion = uint16(2)

	// Field weights for contextual pooling
	FieldWeightCommand = 3.0
	FieldWeightKeyword = 2.0
	FieldWeightDesc    = 1.0
	FieldWeightTag     = 1.2

	// SIF smoothing parameters
	SIFSmoothingA     = 1e-3
	DigitBoost        = 1.20
	ShortTokenPenalty = 0.85

	// Subword feature parameters
	MinNGramLength = 3
	MaxNGramLength = 6

	// Attention-like pooling temperature
	PoolingTemperature = 0.5

	// ANN index parameters
	ANNNumPartitions   = 100
	ANNPointsPerLeaf   = 50
)

// EnhancedIndex provides embedding-based search with ANN indexing.
type EnhancedIndex struct {
	Dimension     int
	WordVectors   map[string][]float32
	WordRanks     map[string]uint32
	WordFreqs     map[string]float64 // IDF-like frequency weighting

	// Command embeddings with field-aware structure
	CmdEmbeddings    [][]float32 // pooled command embeddings
	CmdFieldEmbeds   []FieldEmbeddings
	CmdMetadata      []CmdMeta // lightweight metadata for debugging/ranking

	// ANN index for fast retrieval
	ANNIndex *ANNIndex

	// Dominant component for centering
	dominantComponent []float32

	// Vocabulary statistics
	TotalCommands int
	VocabSize     int
}

// FieldEmbeddings stores separate embeddings for each field
type FieldEmbeddings struct {
	Command []float32
	Desc    []float32
	Keyword []float32
	Tag     []float32
}

// CmdMeta stores lightweight metadata for each command
type CmdMeta struct {
	Command   string
	Platform  []string
	Niche     string
	IsPipeline bool
}

// ANNIndex provides approximate nearest neighbor search
type ANNIndex struct {
	Dimension  int
	Partitions [][]int       // cluster ID -> command indices
	Centroids  [][]float32   // cluster centroids
	Points     [][]float32   // all command embeddings (reference)
	PointsPerLeaf int
}

// SearchResult represents an embedding search result
type SearchResult struct {
	CommandIndex int
	Score        float64
	Similarity   float64
}

// LoadEnhancedEmbeddings loads enhanced command embeddings with field-aware structure.
func (idx *EnhancedIndex) LoadEnhancedEmbeddings(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open enhanced embeddings: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read and verify header
	magic := make([]byte, 4)
	if _, err := io.ReadFull(reader, magic); err != nil {
		return fmt.Errorf("failed to read magic header: %w", err)
	}

	if string(magic) != "WTFS" { // WTF Search enhanced
		return fmt.Errorf("invalid magic: expected WTFS, got %s", string(magic))
	}

	var version uint16
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if version != EnhancedEmbeddingVersion {
		return fmt.Errorf("unsupported version: expected %d, got %d", EnhancedEmbeddingVersion, version)
	}

	// Read header fields
	var numCommands, dimension uint32
	if err := binary.Read(reader, binary.LittleEndian, &numCommands); err != nil {
		return fmt.Errorf("failed to read num commands: %w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &dimension); err != nil {
		return fmt.Errorf("failed to read dimension: %w", err)
	}

	if int(dimension) != idx.Dimension {
		return fmt.Errorf("dimension mismatch: expected %d, got %d", idx.Dimension, dimension)
	}

	// Read embeddings
	idx.CmdEmbeddings = make([][]float32, numCommands)
	idx.CmdFieldEmbeds = make([]FieldEmbeddings, numCommands)
	idx.CmdMetadata = make([]CmdMeta, numCommands)

	for i := uint32(0); i < numCommands; i++ {
		// Read metadata
		meta, err := readCmdMeta(reader)
		if err != nil {
			return fmt.Errorf("failed to read metadata for command %d: %w", i, err)
		}
		idx.CmdMetadata[i] = meta

		// Read pooled embedding
		pooled := make([]float32, dimension)
		if err := binary.Read(reader, binary.LittleEndian, pooled); err != nil {
			return fmt.Errorf("failed to read pooled embedding %d: %w", i, err)
		}
		normalizeVector(pooled)
		idx.CmdEmbeddings[i] = pooled

		// Read field embeddings (optional, for enhanced ranking)
		fieldEmbeds, err := readFieldEmbeddings(reader, int(dimension))
		if err != nil {
			// Field embeddings are optional
			fieldEmbeds = FieldEmbeddings{}
		}
		idx.CmdFieldEmbeds[i] = fieldEmbeds
	}

	idx.TotalCommands = int(numCommands)

	// Remove dominant component and build ANN index
	idx.enhancedFitAndRemoveDominantComponent()
	idx.BuildANNIndex()

	return nil
}

// enhancedFitAndRemoveDominantComponent fits and removes the dominant component for enhanced index.
func (idx *EnhancedIndex) enhancedFitAndRemoveDominantComponent() {
	if idx.TotalCommands < 16 || idx.Dimension <= 0 {
		idx.dominantComponent = nil
		return
	}

	comp := idx.enhancedEstimateDominantComponent(24)
	if comp == nil {
		idx.dominantComponent = nil
		return
	}
	idx.dominantComponent = comp

	for i := range idx.CmdEmbeddings {
		idx.enhancedRemoveDominantComponent(idx.CmdEmbeddings[i])
		normalizeVector(idx.CmdEmbeddings[i])
	}
}

// enhancedEstimateDominantComponent estimates the dominant component using power iteration.
func (idx *EnhancedIndex) enhancedEstimateDominantComponent(iters int) []float32 {
	if len(idx.CmdEmbeddings) == 0 {
		return nil
	}
	v := make([]float64, idx.Dimension)
	for i := range v {
		v[i] = 1.0 / float64(idx.Dimension)
	}

	for it := 0; it < iters; it++ {
		next := make([]float64, idx.Dimension)
		for _, row := range idx.CmdEmbeddings {
			var proj float64
			for j := range row {
				proj += float64(row[j]) * v[j]
			}
			for j := range row {
				next[j] += float64(row[j]) * proj
			}
		}
		var norm float64
		for _, x := range next {
			norm += x * x
		}
		if norm == 0 {
			return nil
		}
		inv := 1.0 / float64(norm)
		for j := range next {
			v[j] = next[j] * inv
		}
	}

	out := make([]float32, idx.Dimension)
	for i := range out {
		out[i] = float32(v[i])
	}
	if !normalizeVector(out) {
		return nil
	}
	return out
}

// enhancedRemoveDominantComponent removes the dominant component from a vector.
func (idx *EnhancedIndex) enhancedRemoveDominantComponent(vec []float32) {
	if idx == nil || len(idx.dominantComponent) == 0 || len(vec) != len(idx.dominantComponent) {
		return
	}
	var proj float64
	for i := range vec {
		proj += float64(vec[i]) * float64(idx.dominantComponent[i])
	}
	for i := range vec {
		vec[i] -= float32(proj * float64(idx.dominantComponent[i]))
	}
}

func readCmdMeta(reader *bufio.Reader) (CmdMeta, error) {
	var meta CmdMeta

	// Read command name
	var cmdLen uint16
	if err := binary.Read(reader, binary.LittleEndian, &cmdLen); err != nil {
		return meta, err
	}
	cmdBytes := make([]byte, cmdLen)
	if _, err := io.ReadFull(reader, cmdBytes); err != nil {
		return meta, err
	}
	meta.Command = string(cmdBytes)

	// Read platform
	var numPlatforms uint8
	if err := binary.Read(reader, binary.LittleEndian, &numPlatforms); err != nil {
		return meta, err
	}
	meta.Platform = make([]string, numPlatforms)
	for i := uint8(0); i < numPlatforms; i++ {
		var platLen uint8
		if err := binary.Read(reader, binary.LittleEndian, &platLen); err != nil {
			return meta, err
		}
		platBytes := make([]byte, platLen)
		if _, err := io.ReadFull(reader, platBytes); err != nil {
			return meta, err
		}
		meta.Platform[i] = string(platBytes)
	}

	// Read niche
	var nicheLen uint16
	if err := binary.Read(reader, binary.LittleEndian, &nicheLen); err != nil {
		return meta, err
	}
	nicheBytes := make([]byte, nicheLen)
	if _, err := io.ReadFull(reader, nicheBytes); err != nil {
		return meta, err
	}
	meta.Niche = string(nicheBytes)

	// Read pipeline flag
	if err := binary.Read(reader, binary.LittleEndian, &meta.IsPipeline); err != nil {
		return meta, err
	}

	return meta, nil
}

func readFieldEmbeddings(reader *bufio.Reader, dim int) (FieldEmbeddings, error) {
	var embeds FieldEmbeddings

	// Each field embedding is prefixed with a presence byte
	var present uint8
	if err := binary.Read(reader, binary.LittleEndian, &present); err != nil {
		return embeds, err
	}

	if present == 0 {
		return embeds, nil
	}

	// Read command field embedding
	embeds.Command = make([]float32, dim)
	if err := binary.Read(reader, binary.LittleEndian, embeds.Command); err != nil {
		return embeds, err
	}

	// Read desc field embedding
	embeds.Desc = make([]float32, dim)
	if err := binary.Read(reader, binary.LittleEndian, embeds.Desc); err != nil {
		return embeds, err
	}

	// Read keyword field embedding
	embeds.Keyword = make([]float32, dim)
	if err := binary.Read(reader, binary.LittleEndian, embeds.Keyword); err != nil {
		return embeds, err
	}

	// Read tag field embedding
	embeds.Tag = make([]float32, dim)
	if err := binary.Read(reader, binary.LittleEndian, embeds.Tag); err != nil {
		return embeds, err
	}

	return embeds, nil
}

// EmbedQueryWithIntent computes an intent-aware query embedding.
func (idx *EnhancedIndex) EmbedQueryWithIntent(query string, intentWeights map[string]float64) []float32 {
	tokens := tokenizeEnhanced(query)
	if len(tokens) == 0 {
		return nil
	}

	// Build weighted centroid with intent enhancement
	sum := make([]float64, idx.Dimension)
	var totalWeight float64

	// Track which tokens are actions vs targets vs keywords
	tokenTypes := classifyTokens(tokens, idx)

	for _, token := range tokens {
		vec, weight, ok := idx.lookupWeightedVector(token)
		if !ok {
			continue
		}

		// Apply intent-based boosting
		if tokenType, exists := tokenTypes[token]; exists {
			switch tokenType {
			case "action":
				weight *= 1.3 // Boost action tokens
			case "target":
				weight *= 1.15 // Boost target tokens
			}
		}

		// Apply explicit intent weights if provided
		if intentWeight, exists := intentWeights[token]; exists {
			weight *= (1.0 + intentWeight*0.5)
		}

		for i, v := range vec {
			sum[i] += float64(v) * weight
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return nil
	}

	out := make([]float32, idx.Dimension)
	for i := range out {
		out[i] = float32(sum[i] / totalWeight)
	}
	idx.enhancedRemoveDominantComponent(out)
	if !normalizeVector(out) {
		return nil
	}

	return out
}

type TokenType string

const (
	TokenAction  TokenType = "action"
	TokenTarget  TokenType = "target"
	TokenKeyword TokenType = "keyword"
)

// classifyTokens identifies token types based on vocabulary patterns
func classifyTokens(tokens []string, idx *EnhancedIndex) map[string]TokenType {
	// Simple heuristic: check if token appears in common action/target vocab
	actions := map[string]bool{
		"find": true, "search": true, "locate": true, "list": true,
		"show": true, "view": true, "display": true, "see": true,
		"create": true, "make": true, "build": true, "generate": true,
		"delete": true, "remove": true, "destroy": true, "clean": true,
		"copy": true, "move": true, "rename": true,
		"run": true, "execute": true, "start": true, "launch": true,
		"install": true, "setup": true, "configure": true,
		"compress": true, "archive": true, "zip": true, "pack": true,
		"extract": true, "unzip": true, "unpack": true, "decompress": true,
		"download": true, "fetch": true, "upload": true,
		"kill": true, "stop": true, "terminate": true,
	}

	targets := map[string]bool{
		"file": true, "files": true, "folder": true, "directory": true,
		"process": true, "processes": true, "service": true, "services": true,
		"network": true, "ip": true, "port": true, "server": true,
		"permission": true, "permissions": true, "user": true, "group": true,
		"archive": true, "zip": true, "tar": true,
	}

	classifications := make(map[string]TokenType)
	for _, token := range tokens {
		if actions[token] {
			classifications[token] = TokenAction
		} else if targets[token] {
			classifications[token] = TokenTarget
		} else {
			classifications[token] = TokenKeyword
		}
	}

	return classifications
}

// EnhancedSearch performs embedding-based search with field-aware matching.
func (idx *EnhancedIndex) EnhancedSearch(queryEmbedding []float32, topK int, useANN bool) []SearchResult {
	if queryEmbedding == nil || len(idx.CmdEmbeddings) == 0 {
		return nil
	}

	var candidates []SearchResult

	if useANN && idx.ANNIndex != nil && idx.TotalCommands > 100 {
		// Use ANN for fast retrieval
		candidates = idx.ANNNNearestNeighbors(queryEmbedding, topK*3)
	} else {
		// Exhaustive search
		candidates = make([]SearchResult, 0, len(idx.CmdEmbeddings))
		for i, cmdEmbed := range idx.CmdEmbeddings {
			sim := CosineSimilarity(queryEmbedding, cmdEmbed)
			// Use raw similarity as score (typically -1 to 1, but usually 0 to 1 for normalized vectors)
			if sim > 0.0 {
				candidates = append(candidates, SearchResult{
					CommandIndex: i,
					Similarity:   sim,
					Score:        sim, // Keep raw cosine similarity, don't inflate
				})
			}
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Return top K
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	return candidates
}

// ANNSearch builds the approximate nearest neighbor index.
func (idx *EnhancedIndex) BuildANNIndex() {
	if idx.TotalCommands < 100 {
		// Too small for ANN
		idx.ANNIndex = nil
		return
	}

	numPartitions := ANNNumPartitions
	if numPartitions > idx.TotalCommands/10 {
		numPartitions = idx.TotalCommands / 10
	}
	if numPartitions < 2 {
		numPartitions = 2
	}

	// Simple k-means style clustering for partitions
	centroids := idx.computeCentroids(numPartitions)
	partitions := make([][]int, numPartitions)

	// Assign each command to nearest centroid
	for i, cmdEmbed := range idx.CmdEmbeddings {
		bestCluster := 0
		bestSim := -1.0

		for j, centroid := range centroids {
			sim := CosineSimilarity(cmdEmbed, centroid)
			if sim > bestSim {
				bestSim = sim
				bestCluster = j
			}
		}

		partitions[bestCluster] = append(partitions[bestCluster], i)
	}

	idx.ANNIndex = &ANNIndex{
		Dimension:     idx.Dimension,
		Partitions:    partitions,
		Centroids:     centroids,
		Points:        idx.CmdEmbeddings,
		PointsPerLeaf: ANNPointsPerLeaf,
	}
}

func (idx *EnhancedIndex) computeCentroids(k int) [][]float32 {
	centroids := make([][]float32, k)

	// Initialize centroids by sampling
	step := len(idx.CmdEmbeddings) / k
	for i := 0; i < k; i++ {
		sampleIdx := i * step
		if sampleIdx >= len(idx.CmdEmbeddings) {
			sampleIdx = len(idx.CmdEmbeddings) - 1
		}
		centroids[i] = make([]float32, idx.Dimension)
		copy(centroids[i], idx.CmdEmbeddings[sampleIdx])
	}

	// Refine centroids with a few iterations
	for iter := 0; iter < 5; iter++ {
		assignments := make([][]int, k)
		sums := make([][]float64, k)
		counts := make([]int, k)

		for i := range sums {
			sums[i] = make([]float64, idx.Dimension)
		}

		// Assign points to nearest centroid
		for i, cmdEmbed := range idx.CmdEmbeddings {
			bestCluster := 0
			bestSim := -1.0

			for j, centroid := range centroids {
				sim := CosineSimilarity(cmdEmbed, centroid)
				if sim > bestSim {
					bestSim = sim
					bestCluster = j
				}
			}

			assignments[bestCluster] = append(assignments[bestCluster], i)
			counts[bestCluster]++

			for d := 0; d < idx.Dimension; d++ {
				sums[bestCluster][d] += float64(cmdEmbed[d])
			}
		}

		// Update centroids
		for j := 0; j < k; j++ {
			if counts[j] == 0 {
				continue
			}
			for d := 0; d < idx.Dimension; d++ {
				centroids[j][d] = float32(sums[j][d] / float64(counts[j]))
			}
			normalizeVector(centroids[j])
		}
	}

	return centroids
}

// ANNNNearestNeighbors finds nearest neighbors using ANN index.
func (idx *EnhancedIndex) ANNNNearestNeighbors(queryEmbedding []float32, topK int) []SearchResult {
	if idx.ANNIndex == nil {
		return nil
	}

	// Find nearest centroid
	bestCluster := 0
	bestSim := -1.0

	for j, centroid := range idx.ANNIndex.Centroids {
		sim := CosineSimilarity(queryEmbedding, centroid)
		if sim > bestSim {
			bestSim = sim
			bestCluster = j
		}
	}

	// Search within the cluster
	candidates := idx.ANNIndex.Partitions[bestCluster]
	results := make([]SearchResult, 0, len(candidates))

	for _, cmdIdx := range candidates {
		cmdEmbed := idx.ANNIndex.Points[cmdIdx]
		sim := CosineSimilarity(queryEmbedding, cmdEmbed)
		if sim > 0.0 {
			results = append(results, SearchResult{
				CommandIndex: cmdIdx,
				Similarity:   sim,
				Score:        sim * 100.0,
			})
		}
	}

	// Sort and return top results
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// ComputeFieldAwareScore computes a score that weighs different fields separately.
func (idx *EnhancedIndex) ComputeFieldAwareScore(queryEmbedding []float32, cmdIdx int) float64 {
	if queryEmbedding == nil || cmdIdx >= len(idx.CmdFieldEmbeds) {
		return 0.0
	}

	fieldEmbeds := idx.CmdFieldEmbeds[cmdIdx]
	var totalScore float64

	// Score each field separately with different weights
	if len(fieldEmbeds.Command) > 0 {
		cmdSim := CosineSimilarity(queryEmbedding, fieldEmbeds.Command)
		totalScore += cmdSim * FieldWeightCommand
	}

	if len(fieldEmbeds.Desc) > 0 {
		descSim := CosineSimilarity(queryEmbedding, fieldEmbeds.Desc)
		totalScore += descSim * FieldWeightDesc
	}

	if len(fieldEmbeds.Keyword) > 0 {
		keywordSim := CosineSimilarity(queryEmbedding, fieldEmbeds.Keyword)
		totalScore += keywordSim * FieldWeightKeyword
	}

	if len(fieldEmbeds.Tag) > 0 {
		tagSim := CosineSimilarity(queryEmbedding, fieldEmbeds.Tag)
		totalScore += tagSim * FieldWeightTag
	}

	// Normalize by total weight
	totalWeight := FieldWeightCommand + FieldWeightDesc + FieldWeightKeyword + FieldWeightTag
	if totalWeight > 0 {
		totalScore /= totalWeight
	}

	return totalScore // Return raw similarity score
}

// Helper: lookup with weighting
func (idx *EnhancedIndex) lookupWeightedVector(token string) ([]float32, float64, bool) {
	for _, candidate := range tokenVariants(token) {
		if vec, ok := idx.WordVectors[candidate]; ok {
			return vec, idx.tokenWeight(candidate), true
		}
	}
	return nil, 0, false
}

func (idx *EnhancedIndex) tokenWeight(token string) float64 {
	weight := 1.0
	vocab := len(idx.WordVectors)
	if vocab > 0 {
		if rank, ok := idx.WordRanks[token]; ok {
			p := float64(rank+1) / float64(vocab)
			weight *= SIFSmoothingA / (SIFSmoothingA + p)
		}
	}

	// Retain structured terms (digits/port-like tokens) slightly more strongly.
	for _, r := range token {
		if unicode.IsDigit(r) {
			weight *= DigitBoost
			break
		}
	}

	if len(token) <= 2 {
		weight *= ShortTokenPenalty
	}

	return weight
}

// tokenizeEnhanced converts text to lowercase tokens for embedding lookup.
func tokenizeEnhanced(text string) []string {
	text = strings.ToLower(text)

	// Split on non-alphanumeric characters
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Filter short tokens
	var tokens []string
	for _, w := range words {
		if len(w) >= 2 {
			tokens = append(tokens, w)
		}
	}

	return tokens
}

// generateSubwordFeatures generates character n-gram features for OOV recovery
func generateSubwordFeatures(token string) []string {
	var features []string
	for n := MinNGramLength; n <= MaxNGramLength && n <= len(token); n++ {
		for i := 0; i <= len(token)-n; i++ {
			features = append(features, token[i:i+n])
		}
	}
	return features
}

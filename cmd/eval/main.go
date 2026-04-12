// Package main implements the WTF evaluation harness for measuring search quality.
//
// Usage:
//
//	go run ./cmd/eval [flags]
//
// Flags:
//
//	-set       Which set to evaluate: "dev", "test", or "all" (default: "dev")
//	-db        Path to commands database (default: "assets/commands.yml")
//	-short     Path to short eval queries (default: "assets/eval_queries.yaml")
//	-long      Path to long eval queries (default: "assets/eval_queries_long.yaml")
//	-limit     Max results per query (default: 5)
//	-no-hints  Disable hints.go command hints for comparison
//	-json      Output results as JSON instead of markdown
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Vedant9500/WTF/internal/database"
)

// evalQuery represents a single evaluation query with expected results.
type evalQuery struct {
	Query    string   `yaml:"query"`
	Slice    string   `yaml:"slice"`
	Set      string   `yaml:"set"`
	Relevant []string `yaml:"relevant"`
}

// evalFile represents the YAML structure of an eval queries file.
type evalFile struct {
	Queries []evalQuery `yaml:"queries"`
}

// queryResult holds the evaluation result for a single query.
type queryResult struct {
	Query        string   `json:"query"`
	Slice        string   `json:"slice"`
	Set          string   `json:"set"`
	TopResult    string   `json:"top_result"`
	Top3Results  []string `json:"top3_results"`
	IsTop1       bool     `json:"is_top1"`
	IsHit3       bool     `json:"is_hit3"`
	RR           float64  `json:"reciprocal_rank"`
	NDCG3        float64  `json:"ndcg3"`
	ResultsCount int      `json:"results_count"`
}

// sliceMetrics holds aggregated metrics for a query slice.
type sliceMetrics struct {
	Slice      string  `json:"slice"`
	QueryCount int     `json:"query_count"`
	Top1       float64 `json:"top1"`
	Hit3       float64 `json:"hit3"`
	MRR        float64 `json:"mrr"`
	NDCG3      float64 `json:"ndcg3"`
	Top1Count  int     `json:"-"`
	Hit3Count  int     `json:"-"`
	RRSum      float64 `json:"-"`
	NDCG3Sum   float64 `json:"-"`
}

// evalReport holds the full evaluation report.
type evalReport struct {
	Config       evalConfig     `json:"config"`
	Aggregate    sliceMetrics   `json:"aggregate"`
	Slices       []sliceMetrics `json:"slices"`
	WorstQueries []queryResult  `json:"worst_queries"`
	AllResults   []queryResult  `json:"all_results"`
}

type evalConfig struct {
	Set                        string  `json:"set"`
	Limit                      int     `json:"limit"`
	NoHints                    bool    `json:"no_hints"`
	FamilyExpansionProfile     string  `json:"family_expansion_profile"`
	FamilyExpansionEnabled     bool    `json:"family_expansion_enabled"`
	FamilyExpansionMaxBases    int     `json:"family_expansion_max_bases"`
	FamilyExpansionMaxTerms    int     `json:"family_expansion_max_terms"`
	FamilyExpansionClarityMax  float64 `json:"family_expansion_clarity_max"`
	FamilyExpansionBlendWeight float64 `json:"family_expansion_blend_weight"`
	NoBigrams                  bool    `json:"no_bigrams"`
	NoCharNGram                bool    `json:"no_char_ngram"`
	NoProximity                bool    `json:"no_proximity"`
	TopTermsCap                int     `json:"top_terms_cap"`
	BM25K1                     float64 `json:"bm25_k1"`
	BM25BCmd                   float64 `json:"bm25_b_cmd"`
	BM25BDesc                  float64 `json:"bm25_b_desc"`
	BM25BKeys                  float64 `json:"bm25_b_keys"`
	BM25BTags                  float64 `json:"bm25_b_tags"`
	BM25WCmd                   float64 `json:"bm25_w_cmd"`
	BM25WDesc                  float64 `json:"bm25_w_desc"`
	BM25WKeys                  float64 `json:"bm25_w_keys"`
	BM25WTags                  float64 `json:"bm25_w_tags"`
	BM25MinIDF                 float64 `json:"bm25_min_idf"`
	ShortQueries               int     `json:"short_queries"`
	LongQueries                int     `json:"long_queries"`
	TotalQueries               int     `json:"total_queries"`
}

type bm25EvalConfig struct {
	K1                         float64
	B                          database.BM25FieldValues
	W                          database.BM25FieldValues
	MinIDF                     float64
	TopTermsCap                int
	EnableFamilyExpansion      bool
	FamilyExpansionMaxBases    int
	FamilyExpansionMaxTerms    int
	FamilyExpansionClarityMax  float64
	FamilyExpansionBlendWeight float64
	DisableBigrams             bool
	DisableCharNGram           bool
	DisableProximity           bool
}

type cliOptions struct {
	setFlag                string
	dbPath                 string
	shortPath              string
	longPath               string
	limit                  int
	noHints                bool
	asJSON                 bool
	familyExpansionProfile string
	bm25Cfg                bm25EvalConfig
}

func main() {
	opts := parseCLIOptions()

	// Load database
	db, err := loadDatabase(opts.dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading database: %v\n", err)
		os.Exit(1)
	}
	db.BuildUniversalIndex()

	// Load eval queries
	shortQueries, err := loadEvalQueries(opts.shortPath, opts.setFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading short queries: %v\n", err)
		os.Exit(1)
	}
	longQueries, err := loadEvalQueries(opts.longPath, opts.setFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading long queries: %v\n", err)
		os.Exit(1)
	}

	allQueries := make([]evalQuery, 0, len(shortQueries)+len(longQueries))
	allQueries = append(allQueries, shortQueries...)
	allQueries = append(allQueries, longQueries...)

	if len(allQueries) == 0 {
		fmt.Fprintf(os.Stderr, "No queries found for set %q\n", opts.setFlag)
		os.Exit(1)
	}

	// Run evaluation
	report := runEvaluation(db, allQueries, opts.limit, opts.noHints, opts.bm25Cfg)
	report.Config = evalConfig{
		Set:                        opts.setFlag,
		Limit:                      opts.limit,
		NoHints:                    opts.noHints,
		FamilyExpansionProfile:     opts.familyExpansionProfile,
		FamilyExpansionEnabled:     opts.bm25Cfg.EnableFamilyExpansion,
		FamilyExpansionMaxBases:    opts.bm25Cfg.FamilyExpansionMaxBases,
		FamilyExpansionMaxTerms:    opts.bm25Cfg.FamilyExpansionMaxTerms,
		FamilyExpansionClarityMax:  opts.bm25Cfg.FamilyExpansionClarityMax,
		FamilyExpansionBlendWeight: opts.bm25Cfg.FamilyExpansionBlendWeight,
		NoBigrams:                  opts.bm25Cfg.DisableBigrams,
		NoCharNGram:                opts.bm25Cfg.DisableCharNGram,
		NoProximity:                opts.bm25Cfg.DisableProximity,
		TopTermsCap:                opts.bm25Cfg.TopTermsCap,
		BM25K1:                     opts.bm25Cfg.K1,
		BM25BCmd:                   opts.bm25Cfg.B.Cmd,
		BM25BDesc:                  opts.bm25Cfg.B.Desc,
		BM25BKeys:                  opts.bm25Cfg.B.Keys,
		BM25BTags:                  opts.bm25Cfg.B.Tags,
		BM25WCmd:                   opts.bm25Cfg.W.Cmd,
		BM25WDesc:                  opts.bm25Cfg.W.Desc,
		BM25WKeys:                  opts.bm25Cfg.W.Keys,
		BM25WTags:                  opts.bm25Cfg.W.Tags,
		BM25MinIDF:                 opts.bm25Cfg.MinIDF,
		ShortQueries:               len(shortQueries),
		LongQueries:                len(longQueries),
		TotalQueries:               len(allQueries),
	}

	if opts.asJSON {
		outputJSON(report)
	} else {
		outputMarkdown(report)
	}
}

func parseCLIOptions() cliOptions {
	setFlag := flag.String("set", "dev", "Which set to evaluate: dev, test, or all")
	dbPath := flag.String("db", "assets/commands.yml", "Path to commands database")
	shortPath := flag.String("short", "assets/eval_queries.yaml", "Path to short eval queries")
	longPath := flag.String("long", "assets/eval_queries_long.yaml", "Path to long eval queries")
	limit := flag.Int("limit", 5, "Max results per query")
	noHints := flag.Bool("no-hints", false, "Disable command hints for comparison")
	asJSON := flag.Bool("json", false, "Output results as JSON")

	bm25K1 := flag.Float64("bm25-k1", 1.2, "BM25F k1 override")
	bm25BCmd := flag.Float64("bm25-b-cmd", 0.75, "BM25F b (command field)")
	bm25BDesc := flag.Float64("bm25-b-desc", 0.75, "BM25F b (description field)")
	bm25BKeys := flag.Float64("bm25-b-keys", 0.7, "BM25F b (keywords field)")
	bm25BTags := flag.Float64("bm25-b-tags", 0.7, "BM25F b (tags field)")
	bm25WCmd := flag.Float64("bm25-w-cmd", 3.5, "BM25F weight (command field)")
	bm25WDesc := flag.Float64("bm25-w-desc", 1.0, "BM25F weight (description field)")
	bm25WKeys := flag.Float64("bm25-w-keys", 2.0, "BM25F weight (keywords field)")
	bm25WTags := flag.Float64("bm25-w-tags", 1.2, "BM25F weight (tags field)")
	bm25MinIDF := flag.Float64("bm25-min-idf", 0.0, "BM25F minimum IDF threshold")
	topTermsCap := flag.Int("top-terms-cap", 10, "Top-IDF terms cap used for long query scoring")
	familyExpansionProfile := flag.String(
		"family-expansion-profile",
		"custom",
		"Family expansion profile: off, safe, experimental, custom",
	)
	enableFamilyExpansion := flag.Bool("family-expansion", false, "Enable Phase 2 corpus-native family expansion")
	familyExpansionMaxBases := flag.Int("family-expansion-max-bases", 3, "Max learned command bases considered for expansion")
	familyExpansionMaxTerms := flag.Int("family-expansion-max-terms", 4, "Max expansion terms appended")
	familyExpansionClarityMax := flag.Float64(
		"family-expansion-clarity-max",
		0.55,
		"Expand only when family clarity/confidence is <= threshold",
	)
	familyExpansionBlendWeight := flag.Float64("family-expansion-blend-weight", 0.25, "Additive blend weight for expansion channel")
	disableBigrams := flag.Bool("disable-bigrams", false, "Disable command/keyword bigram channel")
	disableCharNGram := flag.Bool("disable-char-ngram", false, "Disable character n-gram channel")
	disableProximity := flag.Bool("disable-proximity", false, "Disable description proximity boost")
	flag.Parse()

	bm25Cfg := bm25EvalConfig{
		K1: *bm25K1,
		B: database.BM25FieldValues{
			Cmd:  *bm25BCmd,
			Desc: *bm25BDesc,
			Keys: *bm25BKeys,
			Tags: *bm25BTags,
		},
		W: database.BM25FieldValues{
			Cmd:  *bm25WCmd,
			Desc: *bm25WDesc,
			Keys: *bm25WKeys,
			Tags: *bm25WTags,
		},
		MinIDF:                     *bm25MinIDF,
		TopTermsCap:                *topTermsCap,
		EnableFamilyExpansion:      *enableFamilyExpansion,
		FamilyExpansionMaxBases:    *familyExpansionMaxBases,
		FamilyExpansionMaxTerms:    *familyExpansionMaxTerms,
		FamilyExpansionClarityMax:  *familyExpansionClarityMax,
		FamilyExpansionBlendWeight: *familyExpansionBlendWeight,
		DisableBigrams:             *disableBigrams,
		DisableCharNGram:           *disableCharNGram,
		DisableProximity:           *disableProximity,
	}
	familyProfile := strings.ToLower(strings.TrimSpace(*familyExpansionProfile))
	applyFamilyExpansionProfile(&bm25Cfg, familyProfile)

	return cliOptions{
		setFlag:                *setFlag,
		dbPath:                 *dbPath,
		shortPath:              *shortPath,
		longPath:               *longPath,
		limit:                  *limit,
		noHints:                *noHints,
		asJSON:                 *asJSON,
		familyExpansionProfile: familyProfile,
		bm25Cfg:                bm25Cfg,
	}
}

func applyFamilyExpansionProfile(cfg *bm25EvalConfig, profile string) {
	if cfg == nil {
		return
	}

	switch profile {
	case "off":
		cfg.EnableFamilyExpansion = false
	case "safe":
		cfg.EnableFamilyExpansion = true
		cfg.FamilyExpansionClarityMax = 0.55
		cfg.FamilyExpansionMaxBases = 2
		cfg.FamilyExpansionMaxTerms = 2
		cfg.FamilyExpansionBlendWeight = 0.30
	case "experimental":
		cfg.EnableFamilyExpansion = true
		cfg.FamilyExpansionClarityMax = 0.45
		cfg.FamilyExpansionMaxBases = 3
		cfg.FamilyExpansionMaxTerms = 4
		cfg.FamilyExpansionBlendWeight = 0.30
	case "custom", "":
		// Keep explicit CLI knobs unchanged.
	default:
		// Unknown values fall back to custom behavior.
	}
}

func loadDatabase(path string) (*database.Database, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var commands []database.Command
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	// Populate lowercased cache fields
	for i := range commands {
		commands[i].CommandLower = strings.ToLower(commands[i].Command)
		commands[i].DescriptionLower = strings.ToLower(commands[i].Description)
		commands[i].KeywordsLower = make([]string, len(commands[i].Keywords))
		for j, kw := range commands[i].Keywords {
			commands[i].KeywordsLower[j] = strings.ToLower(kw)
		}
		commands[i].TagsLower = make([]string, len(commands[i].Tags))
		for j, tag := range commands[i].Tags {
			commands[i].TagsLower[j] = strings.ToLower(tag)
		}
	}

	db := &database.Database{Commands: commands}

	// Load embeddings for semantic search
	if err := db.LoadEmbeddings(); err != nil {
		return nil, fmt.Errorf("loading embeddings: %w", err)
	}

	return db, nil
}

func loadEvalQueries(path, setFilter string) ([]evalQuery, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var ef evalFile
	if err := yaml.Unmarshal(data, &ef); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if setFilter == "all" {
		return ef.Queries, nil
	}

	var filtered []evalQuery
	for _, q := range ef.Queries {
		if q.Set == setFilter {
			filtered = append(filtered, q)
		}
	}
	return filtered, nil
}

func runEvaluation(db *database.Database, queries []evalQuery, limit int, noHints bool, cfg bm25EvalConfig) evalReport {
	results := make([]queryResult, 0, len(queries))

	for _, q := range queries {
		result := evaluateQuery(db, q, limit, noHints, cfg)
		results = append(results, result)
	}

	// Compute aggregate and per-slice metrics
	sliceMap := make(map[string]*sliceMetrics)
	var agg sliceMetrics
	agg.Slice = "AGGREGATE"

	for _, r := range results {
		agg.QueryCount++
		if r.IsTop1 {
			agg.Top1Count++
		}
		if r.IsHit3 {
			agg.Hit3Count++
		}
		agg.RRSum += r.RR
		agg.NDCG3Sum += r.NDCG3

		sm, ok := sliceMap[r.Slice]
		if !ok {
			sm = &sliceMetrics{Slice: r.Slice}
			sliceMap[r.Slice] = sm
		}
		sm.QueryCount++
		if r.IsTop1 {
			sm.Top1Count++
		}
		if r.IsHit3 {
			sm.Hit3Count++
		}
		sm.RRSum += r.RR
		sm.NDCG3Sum += r.NDCG3
	}

	// Finalize aggregate
	if agg.QueryCount > 0 {
		n := float64(agg.QueryCount)
		agg.Top1 = float64(agg.Top1Count) / n
		agg.Hit3 = float64(agg.Hit3Count) / n
		agg.MRR = agg.RRSum / n
		agg.NDCG3 = agg.NDCG3Sum / n
	}

	// Finalize slices
	slices := make([]sliceMetrics, 0, len(sliceMap))
	for _, sm := range sliceMap {
		n := float64(sm.QueryCount)
		sm.Top1 = float64(sm.Top1Count) / n
		sm.Hit3 = float64(sm.Hit3Count) / n
		sm.MRR = sm.RRSum / n
		sm.NDCG3 = sm.NDCG3Sum / n
		slices = append(slices, *sm)
	}
	sort.Slice(slices, func(i, j int) bool {
		if slices[i].NDCG3 != slices[j].NDCG3 {
			return slices[i].NDCG3 < slices[j].NDCG3 // worst first
		}
		return slices[i].Slice < slices[j].Slice
	})

	// Worst queries (lowest RR)
	worst := make([]queryResult, len(results))
	copy(worst, results)
	sort.Slice(worst, func(i, j int) bool {
		return worst[i].RR < worst[j].RR
	})
	maxWorst := 15
	if len(worst) < maxWorst {
		maxWorst = len(worst)
	}
	worst = worst[:maxWorst]

	return evalReport{
		Aggregate:    agg,
		Slices:       slices,
		WorstQueries: worst,
		AllResults:   results,
	}
}

func evaluateQuery(db *database.Database, q evalQuery, limit int, noHints bool, cfg bm25EvalConfig) queryResult {
	k1 := cfg.K1
	minIDF := cfg.MinIDF
	b := cfg.B
	w := cfg.W

	options := database.SearchOptions{
		Limit:                      limit,
		UseNLP:                     !noHints,
		UseFuzzy:                   true,
		TopTermsCap:                cfg.TopTermsCap,
		EnableFamilyExpansion:      cfg.EnableFamilyExpansion,
		FamilyExpansionMaxBases:    cfg.FamilyExpansionMaxBases,
		FamilyExpansionMaxTerms:    cfg.FamilyExpansionMaxTerms,
		FamilyExpansionClarityMax:  cfg.FamilyExpansionClarityMax,
		FamilyExpansionBlendWeight: cfg.FamilyExpansionBlendWeight,
		DisableBigrams:             cfg.DisableBigrams,
		DisableCharNGram:           cfg.DisableCharNGram,
		DisableProximity:           cfg.DisableProximity,
		BM25Overrides: &database.BM25Overrides{
			K1:     &k1,
			B:      &b,
			W:      &w,
			MinIDF: &minIDF,
		},
	}

	results := db.SearchUniversal(q.Query, options)

	qr := queryResult{
		Query:        q.Query,
		Slice:        q.Slice,
		Set:          q.Set,
		ResultsCount: len(results),
	}

	if len(results) > 0 {
		qr.TopResult = results[0].Command.Command
	}

	top3 := make([]string, 0, 3)
	for i := 0; i < len(results) && i < 3; i++ {
		top3 = append(top3, results[i].Command.Command)
	}
	qr.Top3Results = top3

	// Check relevance: a result is relevant if its command starts with any relevant prefix
	for i, r := range results {
		if !isRelevant(r.Command.Command, q.Relevant) {
			continue
		}
		rank := i + 1
		if rank == 1 {
			qr.IsTop1 = true
		}
		if rank <= 3 {
			qr.IsHit3 = true
		}
		if qr.RR == 0 {
			qr.RR = 1.0 / float64(rank)
		}
		break // Only first relevant result matters for RR
	}

	// NDCG@3
	qr.NDCG3 = computeNDCG(results, q.Relevant, 3)

	return qr
}

func isRelevant(command string, relevantPrefixes []string) bool {
	cmdLower := strings.ToLower(command)
	// Also extract base command (first word or first two words for subcommands)
	cmdParts := strings.Fields(cmdLower)

	for _, prefix := range relevantPrefixes {
		prefixLower := strings.ToLower(prefix)

		// Check if command starts with the relevant prefix
		if strings.HasPrefix(cmdLower, prefixLower) {
			return true
		}

		// Check base command match (e.g., "git" matches "git commit -m ...")
		prefixParts := strings.Fields(prefixLower)
		if len(prefixParts) > 0 && len(cmdParts) >= len(prefixParts) {
			match := true
			for k := 0; k < len(prefixParts); k++ {
				if cmdParts[k] != prefixParts[k] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

func computeNDCG(results []database.SearchResult, relevant []string, k int) float64 {
	if len(results) == 0 || len(relevant) == 0 {
		return 0
	}

	// DCG@k
	dcg := 0.0
	for i := 0; i < k && i < len(results); i++ {
		if isRelevant(results[i].Command.Command, relevant) {
			dcg += 1.0 / math.Log2(float64(i+2)) // log2(rank+1), rank is 1-indexed
		}
	}

	// IDCG@k (ideal: all top-k are relevant)
	numRelevant := len(relevant)
	if numRelevant > k {
		numRelevant = k
	}
	idcg := 0.0
	for i := 0; i < numRelevant; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

func outputMarkdown(report evalReport) {
	fmt.Println("# WTF Evaluation Report")
	fmt.Println()
	fmt.Printf("**Set:** %s | **Queries:** %d (short: %d, long: %d) | **Limit:** %d | **Hints:** %v\n\n",
		report.Config.Set, report.Config.TotalQueries,
		report.Config.ShortQueries, report.Config.LongQueries,
		report.Config.Limit, !report.Config.NoHints)
	fmt.Printf(
		"**BM25:** k1=%.3f minIDF=%.3f topTerms=%d | b(cmd/desc/keys/tags)=%.3f/%.3f/%.3f/%.3f | w(cmd/desc/keys/tags)=%.3f/%.3f/%.3f/%.3f\n\n",
		report.Config.BM25K1, report.Config.BM25MinIDF, report.Config.TopTermsCap,
		report.Config.BM25BCmd, report.Config.BM25BDesc, report.Config.BM25BKeys, report.Config.BM25BTags,
		report.Config.BM25WCmd, report.Config.BM25WDesc, report.Config.BM25WKeys, report.Config.BM25WTags)
	fmt.Printf("**Channels:** bigrams=%v char-ngram=%v proximity=%v\n\n",
		!report.Config.NoBigrams, !report.Config.NoCharNGram, !report.Config.NoProximity)

	// Aggregate
	fmt.Println("## Aggregate Metrics")
	fmt.Println()
	fmt.Println("| Metric | Value |")
	fmt.Println("|--------|-------|")
	fmt.Printf("| Top1   | %.1f%% (%d/%d) |\n", report.Aggregate.Top1*100, report.Aggregate.Top1Count, report.Aggregate.QueryCount)
	fmt.Printf("| Hit@3  | %.1f%% (%d/%d) |\n", report.Aggregate.Hit3*100, report.Aggregate.Hit3Count, report.Aggregate.QueryCount)
	fmt.Printf("| MRR    | %.4f |\n", report.Aggregate.MRR)
	fmt.Printf("| NDCG@3 | %.4f |\n", report.Aggregate.NDCG3)
	fmt.Println()

	// Per-slice
	fmt.Println("## Per-Slice Metrics (sorted by NDCG@3, worst first)")
	fmt.Println()
	fmt.Println("| Slice | N | Top1 | Hit@3 | MRR | NDCG@3 |")
	fmt.Println("|-------|---|------|-------|-----|--------|")
	for _, s := range report.Slices {
		fmt.Printf("| %s | %d | %.0f%% | %.0f%% | %.3f | %.3f |\n",
			s.Slice, s.QueryCount, s.Top1*100, s.Hit3*100, s.MRR, s.NDCG3)
	}
	fmt.Println()

	// Worst queries
	fmt.Println("## Worst Queries (lowest reciprocal rank)")
	fmt.Println()
	fmt.Println("| Query | Slice | RR | Top Result |")
	fmt.Println("|-------|-------|----|------------|")
	for _, w := range report.WorstQueries {
		topResult := w.TopResult
		if len(topResult) > 40 {
			topResult = topResult[:40] + "..."
		}
		query := w.Query
		if len(query) > 60 {
			query = query[:60] + "..."
		}
		fmt.Printf("| %s | %s | %.2f | %s |\n", query, w.Slice, w.RR, topResult)
	}
}

func outputJSON(report evalReport) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

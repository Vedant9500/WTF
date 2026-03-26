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
	Slice       string  `json:"slice"`
	QueryCount  int     `json:"query_count"`
	Top1        float64 `json:"top1"`
	Hit3        float64 `json:"hit3"`
	MRR         float64 `json:"mrr"`
	NDCG3       float64 `json:"ndcg3"`
	Top1Count   int     `json:"-"`
	Hit3Count   int     `json:"-"`
	RRSum       float64 `json:"-"`
	NDCG3Sum    float64 `json:"-"`
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
	Set          string `json:"set"`
	Limit        int    `json:"limit"`
	NoHints      bool   `json:"no_hints"`
	ShortQueries int    `json:"short_queries"`
	LongQueries  int    `json:"long_queries"`
	TotalQueries int    `json:"total_queries"`
}

func main() {
	setFlag := flag.String("set", "dev", "Which set to evaluate: dev, test, or all")
	dbPath := flag.String("db", "assets/commands.yml", "Path to commands database")
	shortPath := flag.String("short", "assets/eval_queries.yaml", "Path to short eval queries")
	longPath := flag.String("long", "assets/eval_queries_long.yaml", "Path to long eval queries")
	limit := flag.Int("limit", 5, "Max results per query")
	noHints := flag.Bool("no-hints", false, "Disable command hints for comparison")
	asJSON := flag.Bool("json", false, "Output results as JSON")
	flag.Parse()

	// Load database
	db, err := loadDatabase(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading database: %v\n", err)
		os.Exit(1)
	}
	db.BuildUniversalIndex()

	// Load eval queries
	shortQueries, err := loadEvalQueries(*shortPath, *setFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading short queries: %v\n", err)
		os.Exit(1)
	}
	longQueries, err := loadEvalQueries(*longPath, *setFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading long queries: %v\n", err)
		os.Exit(1)
	}

	allQueries := append(shortQueries, longQueries...)

	if len(allQueries) == 0 {
		fmt.Fprintf(os.Stderr, "No queries found for set %q\n", *setFlag)
		os.Exit(1)
	}

	// Run evaluation
	report := runEvaluation(db, allQueries, *limit, *noHints)
	report.Config = evalConfig{
		Set:          *setFlag,
		Limit:        *limit,
		NoHints:      *noHints,
		ShortQueries: len(shortQueries),
		LongQueries:  len(longQueries),
		TotalQueries: len(allQueries),
	}

	if *asJSON {
		outputJSON(report)
	} else {
		outputMarkdown(report)
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

func runEvaluation(db *database.Database, queries []evalQuery, limit int, noHints bool) evalReport {
	results := make([]queryResult, 0, len(queries))

	for _, q := range queries {
		result := evaluateQuery(db, q, limit, noHints)
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

func evaluateQuery(db *database.Database, q evalQuery, limit int, noHints bool) queryResult {
	options := database.SearchOptions{
		Limit:    limit,
		UseNLP:   !noHints,
		UseFuzzy: true,
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
		if isRelevant(r.Command.Command, q.Relevant) {
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


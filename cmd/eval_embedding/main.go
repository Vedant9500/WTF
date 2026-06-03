package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Vedant9500/WTF/internal/database"
)

const evalModeHybrid = "hybrid"

// EvalQuery represents a single evaluation query.
type EvalQuery struct {
	Query    string   `yaml:"query"`
	Slice    string   `yaml:"slice"`
	Set      string   `yaml:"set"`
	Relevant []string `yaml:"relevant"`
}

// EvalQueries represents the evaluation queries file.
type EvalQueries struct {
	Queries []EvalQuery `yaml:"queries"`
}

// Metrics holds evaluation metrics.
type Metrics struct {
	TotalQueries       int
	DevQueries         int
	TestQueries        int
	FoundQueries       int
	HitAt1             int
	HitAt3             int
	HitAt5             int
	MeanRank           float64
	MeanReciprocalRank float64
	SliceMetrics       map[string]*SliceMetrics
}

// SliceMetrics holds per-slice metrics.
type SliceMetrics struct {
	Total  int
	Found  int
	HitAt1 int
	HitAt3 int
	HitAt5 int
}

func main() {
	mode := flag.String("mode", "hybrid", "search mode: hybrid, bm25, or sentence")
	flag.Parse()
	evalMode := normalizeEvalMode(*mode)

	fmt.Println("=== WTF Embedding-Based Search Evaluation ===")
	fmt.Printf("Mode: %s\n\n", evalMode)

	fmt.Println("Loading command database...")
	db, err := database.LoadDatabase("assets/commands.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load database: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d commands\n\n", len(db.Commands))

	fmt.Println("Loading evaluation queries...")
	queries, err := loadEvalQueries("assets/eval_queries.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load evaluation queries: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d queries\n\n", len(queries.Queries))

	fmt.Println("Running evaluation...")
	metrics := evaluate(db, queries, evalMode)
	printMetrics(metrics)
}

func loadEvalQueries(filename string) (*EvalQueries, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var queries EvalQueries
	if err := yaml.Unmarshal(data, &queries); err != nil {
		return nil, err
	}

	return &queries, nil
}

func evaluate(db *database.Database, queries *EvalQueries, mode string) *Metrics {
	metrics := &Metrics{
		SliceMetrics: make(map[string]*SliceMetrics),
	}

	var totalRank float64
	var totalMRR float64
	mode = strings.ToLower(strings.TrimSpace(mode))

	for idx, q := range queries.Queries {
		options := searchOptionsForMode(mode)
		results := db.SearchUniversal(q.Query, options)

		if idx < 5 {
			printDebugQuery(q, results)
		}

		rank := findBestRank(searchResultCommands(results), q.Relevant)
		totalRank, totalMRR = recordMetrics(metrics, q, rank, totalRank, totalMRR)
		printQueryOutcome(q, rank)
	}

	if metrics.FoundQueries > 0 {
		metrics.MeanRank = totalRank / float64(metrics.FoundQueries)
	}
	if metrics.TotalQueries > 0 {
		metrics.MeanReciprocalRank = totalMRR / float64(metrics.TotalQueries)
	}

	return metrics
}

func printDebugQuery(q EvalQuery, results []database.SearchResult) {
	fmt.Printf("  Query: %s\n", q.Query)
	fmt.Printf("  Relevant: %v\n", q.Relevant)
	fmt.Printf("  Results (%d):\n", len(results))
	for i, r := range results {
		if r.Command != nil {
			fmt.Printf("    [%d] %s (score=%.2f)\n", i+1, r.Command.Command, r.Score)
		}
	}
}

func searchResultCommands(results []database.SearchResult) []string {
	commands := make([]string, 0, len(results))
	for _, r := range results {
		if r.Command != nil {
			commands = append(commands, strings.ToLower(r.Command.Command))
		}
	}
	return commands
}

func recordMetrics(
	metrics *Metrics,
	q EvalQuery,
	rank int,
	totalRank float64,
	totalMRR float64,
) (nextTotalRank, nextTotalMRR float64) {
	nextTotalRank = totalRank
	nextTotalMRR = totalMRR

	metrics.TotalQueries++
	if q.Set == "dev" {
		metrics.DevQueries++
	} else {
		metrics.TestQueries++
	}

	if rank > 0 {
		metrics.FoundQueries++
		nextTotalRank += float64(rank)
		nextTotalMRR += 1.0 / float64(rank)
	}
	recordHitCounts(rank, &metrics.HitAt1, &metrics.HitAt3, &metrics.HitAt5)

	if _, ok := metrics.SliceMetrics[q.Slice]; !ok {
		metrics.SliceMetrics[q.Slice] = &SliceMetrics{}
	}
	sm := metrics.SliceMetrics[q.Slice]
	sm.Total++
	if rank > 0 {
		sm.Found++
	}
	recordHitCounts(rank, &sm.HitAt1, &sm.HitAt3, &sm.HitAt5)

	return nextTotalRank, nextTotalMRR
}

func recordHitCounts(rank int, hitAt1, hitAt3, hitAt5 *int) {
	if rank > 0 && rank <= 1 {
		(*hitAt1)++
	}
	if rank > 0 && rank <= 3 {
		(*hitAt3)++
	}
	if rank > 0 && rank <= 5 {
		(*hitAt5)++
	}
}

func printQueryOutcome(q EvalQuery, rank int) {
	status := "PASS"
	rankText := fmt.Sprintf("%d", rank)
	if rank == 0 {
		rankText = "-"
	}
	if rank == 0 || rank > 5 {
		status = "MISS"
	}
	fmt.Printf("  %-4s %-50s rank=%s relevant=%v\n", status, q.Query, rankText, q.Relevant)
}

func searchOptionsForMode(mode string) database.SearchOptions {
	options := database.SearchOptions{
		Limit:        5,
		UseNLP:       true,
		UseEmbedding: true,
	}

	switch mode {
	case "bm25":
		options.DisableEmbedding = true
	case "sentence":
		options.UseSentenceEmbeddings = true
	case evalModeHybrid, "":
		// Default: enhanced embedding + BM25F hybrid when enhanced assets are available.
	}

	return options
}

func normalizeEvalMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "bm25", "sentence", evalModeHybrid, "":
		if mode == "" {
			return evalModeHybrid
		}
		return mode
	default:
		fmt.Fprintf(os.Stderr, "warning: unknown mode %q, using %s\n", mode, evalModeHybrid)
		return evalModeHybrid
	}
}

func findBestRank(results, relevant []string) int {
	for i, result := range results {
		if isRelevant(result, relevant) {
			return i + 1
		}
	}
	return 0
}

func isRelevant(command string, relevantPrefixes []string) bool {
	cmdLower := strings.ToLower(command)
	cmdParts := strings.Fields(cmdLower)

	for _, prefix := range relevantPrefixes {
		prefixLower := strings.ToLower(prefix)
		if strings.HasPrefix(cmdLower, prefixLower) {
			return true
		}

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

func printMetrics(m *Metrics) {
	fmt.Println()
	fmt.Println("=== Evaluation Results ===")
	fmt.Println()

	hitRate1 := percent(m.HitAt1, m.TotalQueries)
	hitRate3 := percent(m.HitAt3, m.TotalQueries)
	hitRate5 := percent(m.HitAt5, m.TotalQueries)

	fmt.Printf("Total Queries:        %d\n", m.TotalQueries)
	fmt.Printf("  Dev Set:            %d\n", m.DevQueries)
	fmt.Printf("  Test Set:           %d\n", m.TestQueries)
	fmt.Printf("  Found Relevant:     %d\n", m.FoundQueries)
	fmt.Println()
	fmt.Printf("Hit@1:                %d/%d (%.1f%%)\n", m.HitAt1, m.TotalQueries, hitRate1)
	fmt.Printf("Hit@3:                %d/%d (%.1f%%)\n", m.HitAt3, m.TotalQueries, hitRate3)
	fmt.Printf("Hit@5:                %d/%d (%.1f%%)\n", m.HitAt5, m.TotalQueries, hitRate5)
	fmt.Printf("Mean Rank (hits):     %.2f\n", m.MeanRank)
	fmt.Printf("Mean Reciprocal Rank: %.3f\n", m.MeanReciprocalRank)
	fmt.Println()

	fmt.Println("=== Per-Slice Metrics ===")
	fmt.Println()

	for slice, sm := range m.SliceMetrics {
		fmt.Printf("%-25s: total=%d found=%d hit@1=%d/%d hit@3=%d/%d hit@5=%d/%d\n",
			slice, sm.Total, sm.Found,
			sm.HitAt1, sm.Total,
			sm.HitAt3, sm.Total,
			sm.HitAt5, sm.Total)
	}
}

func percent(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}

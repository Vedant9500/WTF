package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
	"gopkg.in/yaml.v3"
)

type evalDataset struct {
	Queries []evalQuery `yaml:"queries"`
}

type evalQuery struct {
	Query    string   `yaml:"query"`
	Relevant []string `yaml:"relevant"`
}

type perQueryResult struct {
	Query       string
	Top1Hit     bool
	TopKHit     bool
	RR          float64
	NDCG        float64
	BestRank    int
	TopCommands []string
}

func main() {
	var (
		dbPath       = flag.String("db", "assets/commands.yml", "Path to commands database YAML")
		evalPath     = flag.String("eval", "assets/eval_queries.yaml", "Path to evaluation query set YAML")
		limit        = flag.Int("limit", 10, "Search limit per query")
		k            = flag.Int("k", 3, "Evaluation cutoff for Hit@K and NDCG@K")
		useNLP       = flag.Bool("nlp", true, "Enable NLP enhancements")
		useFuzzy     = flag.Bool("fuzzy", true, "Enable fuzzy fallback")
		allPlatforms = flag.Bool("all-platforms", true, "Search across all platforms")
		verbose      = flag.Bool("verbose", true, "Print per-query details")
	)
	flag.Parse()

	if *k <= 0 {
		log.Fatal("k must be > 0")
	}
	if *limit <= 0 {
		log.Fatal("limit must be > 0")
	}

	dataset, err := loadEvalDataset(*evalPath)
	if err != nil {
		log.Fatalf("failed to load eval dataset: %v", err)
	}
	if len(dataset.Queries) == 0 {
		log.Fatal("evaluation dataset has no queries")
	}

	db, err := database.LoadDatabase(*dbPath)
	if err != nil {
		log.Fatalf("failed to load database: %v", err)
	}

	results := runEvaluation(db, dataset.Queries, *limit, *k, *useNLP, *useFuzzy, *allPlatforms)
	printSummary(results, *k, *verbose)
}

func loadEvalDataset(path string) (*evalDataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ds evalDataset
	if err := yaml.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}

func runEvaluation(db *database.Database, queries []evalQuery, limit, k int, useNLP, useFuzzy, allPlatforms bool) []perQueryResult {
	out := make([]perQueryResult, 0, len(queries))

	for _, q := range queries {
		res := db.SearchUniversal(q.Query, database.SearchOptions{
			Limit:        limit,
			UseNLP:       useNLP,
			UseFuzzy:     useFuzzy,
			AllPlatforms: allPlatforms,
		})

		pred := make([]string, 0, len(res))
		for _, r := range res {
			if r.Command == nil {
				continue
			}
			pred = append(pred, strings.ToLower(strings.TrimSpace(r.Command.Command)))
		}

		relevant := make([]string, 0, len(q.Relevant))
		for _, rel := range q.Relevant {
			relevant = append(relevant, strings.ToLower(strings.TrimSpace(rel)))
		}

		topCommands := pred
		if len(topCommands) > k {
			topCommands = topCommands[:k]
		}

		bestRank := firstRelevantRank(pred, relevant)
		rr := 0.0
		top1Hit := false
		topKHit := false
		if bestRank > 0 {
			rr = 1.0 / float64(bestRank)
			top1Hit = bestRank == 1
			topKHit = bestRank <= k
		}

		nDCG := ndcgAtK(pred, relevant, k)

		out = append(out, perQueryResult{
			Query:       q.Query,
			Top1Hit:     top1Hit,
			TopKHit:     topKHit,
			RR:          rr,
			NDCG:        nDCG,
			BestRank:    bestRank,
			TopCommands: topCommands,
		})
	}

	return out
}

func firstRelevantRank(pred, relevant []string) int {
	for i, p := range pred {
		if isRelevant(p, relevant) {
			return i + 1
		}
	}
	return 0
}

func isRelevant(pred string, relevant []string) bool {
	for _, rel := range relevant {
		if rel == "" {
			continue
		}
		if pred == rel || strings.HasPrefix(pred, rel+" ") {
			return true
		}
	}
	return false
}

func ndcgAtK(pred, relevant []string, k int) float64 {
	if k <= 0 {
		return 0
	}

	limit := k
	if len(pred) < limit {
		limit = len(pred)
	}

	dcg := 0.0
	for i := 0; i < limit; i++ {
		rel := 0.0
		if isRelevant(pred[i], relevant) {
			rel = 1.0
		}
		dcg += rel / math.Log2(float64(i+2))
	}

	idealHits := k
	if len(relevant) < idealHits {
		idealHits = len(relevant)
	}
	if idealHits == 0 {
		return 0
	}

	idcg := 0.0
	for i := 0; i < idealHits; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}
	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

func printSummary(results []perQueryResult, k int, verbose bool) {
	total := float64(len(results))
	if total == 0 {
		fmt.Println("No evaluation results")
		return
	}

	top1 := 0.0
	topK := 0.0
	mrr := 0.0
	meanNDCG := 0.0

	if verbose {
		fmt.Println("Per-query results:")
	}

	for _, r := range results {
		if r.Top1Hit {
			top1++
		}
		if r.TopKHit {
			topK++
		}
		mrr += r.RR
		meanNDCG += r.NDCG

		if verbose {
			bestRank := "none"
			if r.BestRank > 0 {
				bestRank = fmt.Sprintf("%d", r.BestRank)
			}
			fmt.Printf("- %s\n", r.Query)
			fmt.Printf("  best_rank=%s top1=%t hit@%d=%t rr=%.3f ndcg@%d=%.3f\n", bestRank, r.Top1Hit, k, r.TopKHit, r.RR, k, r.NDCG)
			if len(r.TopCommands) > 0 {
				fmt.Printf("  top_commands=%s\n", strings.Join(r.TopCommands, " | "))
			}
		}
	}

	summary := map[string]float64{
		"Top1":  top1 / total,
		"HitK":  topK / total,
		"MRR":   mrr / total,
		"NDCG":  meanNDCG / total,
		"Count": total,
	}

	fmt.Println("\nEvaluation summary:")
	keys := []string{"Count", "Top1", "HitK", "MRR", "NDCG"}
	for _, key := range keys {
		switch key {
		case "Count":
			fmt.Printf("- queries: %.0f\n", summary[key])
		case "HitK":
			fmt.Printf("- Hit@%d: %.3f\n", k, summary[key])
		case "NDCG":
			fmt.Printf("- NDCG@%d: %.3f\n", k, summary[key])
		default:
			fmt.Printf("- %s: %.3f\n", key, summary[key])
		}
	}

	worst := append([]perQueryResult(nil), results...)
	sort.Slice(worst, func(i, j int) bool {
		if worst[i].NDCG == worst[j].NDCG {
			return worst[i].RR < worst[j].RR
		}
		return worst[i].NDCG < worst[j].NDCG
	})

	show := 3
	if len(worst) < show {
		show = len(worst)
	}
	if show > 0 {
		fmt.Println("\nLowest-performing queries:")
		for i := 0; i < show; i++ {
			r := worst[i]
			fmt.Printf("- %s (rr=%.3f ndcg@%d=%.3f)\n", r.Query, r.RR, k, r.NDCG)
		}
	}
}

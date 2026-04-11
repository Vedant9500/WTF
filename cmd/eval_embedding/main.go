package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
	"gopkg.in/yaml.v3"
)

// EvalQuery represents a single evaluation query
type EvalQuery struct {
	Query    string   `yaml:"query"`
	Slice    string   `yaml:"slice"`
	Set      string   `yaml:"set"`
	Relevant []string `yaml:"relevant"`
}

// EvalQueries represents the evaluation queries file
type EvalQueries struct {
	Queries []EvalQuery `yaml:"queries"`
}

// Metrics holds evaluation metrics
type Metrics struct {
	TotalQueries   int
	DevQueries     int
	TestQueries    int
	HitAt1         int
	HitAt3         int
	HitAt5         int
	MeanRank       float64
	MeanReciprocalRank float64
	SliceMetrics   map[string]*SliceMetrics
}

// SliceMetrics holds per-slice metrics
type SliceMetrics struct {
	Total int
	HitAt1 int
	HitAt3 int
	HitAt5 int
}

func main() {
	fmt.Println("=== WTF Embedding-Based Search Evaluation ===")
	fmt.Println()

	// Load database
	fmt.Println("📦 Loading command database...")
	dbPath := "assets/commands.yml"
	db, err := database.LoadDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load database: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Loaded %d commands\n\n", len(db.Commands))

	// Load evaluation queries
	fmt.Println("📋 Loading evaluation queries...")
	queries, err := loadEvalQueries("assets/eval_queries.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load evaluation queries: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Loaded %d queries\n\n", len(queries.Queries))

	// Run evaluation
	fmt.Println("🔍 Running evaluation...")
	metrics := evaluate(db, queries)

	// Print results
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

func evaluate(db *database.Database, queries *EvalQueries) *Metrics {
	metrics := &Metrics{
		SliceMetrics: make(map[string]*SliceMetrics),
	}

	var totalRank float64
	var totalMRR float64

	for idx, q := range queries.Queries {
		// Search for the query
		results := db.SearchUniversal(q.Query, database.SearchOptions{
			Limit:        5,
			UseNLP:       true,
			UseEmbedding: true,
		})

		// Print first 5 queries for debugging
		if idx < 5 {
			fmt.Printf("  Query: %s\n", q.Query)
			fmt.Printf("  Relevant: %v\n", q.Relevant)
			fmt.Printf("  Results (%d):\n", len(results))
			for i, r := range results {
				if r.Command != nil {
					fmt.Printf("    [%d] %s (score=%.2f)\n", i+1, r.Command.Command, r.Score)
				}
			}
		}

		// Extract command names from results
		resultCommands := make([]string, 0, len(results))
		for _, r := range results {
			if r.Command != nil {
				resultCommands = append(resultCommands, strings.ToLower(r.Command.Command))
			}
		}

		// Check if relevant commands are in results
		rank := findBestRank(resultCommands, q.Relevant)
		
		// Update metrics
		metrics.TotalQueries++
		if q.Set == "dev" {
			metrics.DevQueries++
		} else {
			metrics.TestQueries++
		}

		if rank <= 1 {
			metrics.HitAt1++
		}
		if rank <= 3 {
			metrics.HitAt3++
		}
		if rank <= 5 {
			metrics.HitAt5++
		}

		if rank > 0 {
			totalRank += float64(rank)
			totalMRR += 1.0 / float64(rank)
		}

		// Update slice metrics
		if _, ok := metrics.SliceMetrics[q.Slice]; !ok {
			metrics.SliceMetrics[q.Slice] = &SliceMetrics{}
		}
		sm := metrics.SliceMetrics[q.Slice]
		sm.Total++
		if rank <= 1 {
			sm.HitAt1++
		}
		if rank <= 3 {
			sm.HitAt3++
		}
		if rank <= 5 {
			sm.HitAt5++
		}

		// Print per-query result
		status := "✓"
		if rank > 5 {
			status = "✗"
		}
		fmt.Printf("  %s %-50s rank=%d relevant=%v\n", status, q.Query, rank, q.Relevant)
	}

	if metrics.HitAt5 > 0 {
		metrics.MeanRank = totalRank / float64(metrics.HitAt5)
		metrics.MeanReciprocalRank = totalMRR / float64(metrics.TotalQueries)
	}

	return metrics
}

func findBestRank(results []string, relevant []string) int {
	for i, result := range results {
		for _, rel := range relevant {
			relLower := strings.ToLower(rel)
			resultLower := strings.ToLower(result)
			// Check if result contains relevant command or vice versa
			if strings.Contains(resultLower, relLower) || 
			   strings.Contains(relLower, resultLower) ||
			   strings.HasPrefix(resultLower, relLower) ||
			   strings.HasPrefix(relLower, resultLower) {
				return i + 1
			}
			// Check for command name matching (e.g., "git branch -d" matches "git branch")
			resultParts := strings.Fields(resultLower)
			relParts := strings.Fields(relLower)
			if len(resultParts) > 0 && len(relParts) > 0 {
				if resultParts[0] == relParts[0] {
					// First word matches, check more
					matchCount := 0
					for j := 0; j < len(relParts) && j < len(resultParts); j++ {
						if resultParts[j] == relParts[j] {
							matchCount++
						}
					}
					if matchCount >= len(relParts)/2 {
						return i + 1
					}
				}
			}
		}
	}
	return math.MaxInt64
}

func printMetrics(m *Metrics) {
	fmt.Println()
	fmt.Println("=== Evaluation Results ===")
	fmt.Println()
	
	hitRate1 := float64(m.HitAt1) / float64(m.TotalQueries) * 100
	hitRate3 := float64(m.HitAt3) / float64(m.TotalQueries) * 100
	hitRate5 := float64(m.HitAt5) / float64(m.TotalQueries) * 100

	fmt.Printf("Total Queries:      %d\n", m.TotalQueries)
	fmt.Printf("  Dev Set:          %d\n", m.DevQueries)
	fmt.Printf("  Test Set:         %d\n", m.TestQueries)
	fmt.Println()
	fmt.Printf("Hit@1:              %d/%d (%.1f%%)\n", m.HitAt1, m.TotalQueries, hitRate1)
	fmt.Printf("Hit@3:              %d/%d (%.1f%%)\n", m.HitAt3, m.TotalQueries, hitRate3)
	fmt.Printf("Hit@5:              %d/%d (%.1f%%)\n", m.HitAt5, m.TotalQueries, hitRate5)
	fmt.Printf("Mean Rank:          %.2f\n", m.MeanRank)
	fmt.Printf("Mean Reciprocal Rank: %.3f\n", m.MeanReciprocalRank)
	fmt.Println()

	// Print per-slice metrics
	fmt.Println("=== Per-Slice Metrics ===")
	fmt.Println()
	
	for slice, sm := range m.SliceMetrics {
		fmt.Printf("%-25s: total=%d hit@1=%d/%d hit@3=%d/%d hit@5=%d/%d\n",
			slice, sm.Total,
			sm.HitAt1, sm.Total,
			sm.HitAt3, sm.Total,
			sm.HitAt5, sm.Total)
	}
}

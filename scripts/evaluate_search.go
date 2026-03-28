package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Vedant9500/WTF/internal/database"
)

type evalDataset struct {
	Queries []evalQuery `yaml:"queries"`
}

type evalQuery struct {
	Query    string   `yaml:"query"`
	Relevant []string `yaml:"relevant"`
	Slice    string   `yaml:"slice,omitempty"`
}

type perQueryResult struct {
	Query       string
	Slice       string
	Top1Hit     bool
	TopKHit     bool
	RR          float64
	NDCG        float64
	BestRank    int
	TopCommands []string
}

type metricsSummary struct {
	Count float64
	Top1  float64
	HitK  float64
	MRR   float64
	NDCG  float64
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
}

type sweepResult struct {
	Config  bm25EvalConfig
	Summary metricsSummary
}

type cliOptions struct {
	dbPath       string
	evalPath     string
	limit        int
	k            int
	useNLP       bool
	useFuzzy     bool
	allPlatforms bool
	verbose      bool
	minTop1      float64
	minHitK      float64
	minMRR       float64
	minNDCG      float64

	baseConfig bm25EvalConfig

	sweepMode            bool
	sweepTopN            int
	sweepWorkers         int
	sweepProgressSeconds int
	gridK1               string
	gridBCmd             string
	gridBDesc            string
	gridBKeys            string
	gridBTags            string
	gridWCmd             string
	gridWDesc            string
	gridWKeys            string
	gridWTags            string
	gridMinIDF           string
	gridTopTerms         string
}

type sweepFlagRefs struct {
	sweepMode            *bool
	sweepTopN            *int
	sweepWorkers         *int
	sweepProgressSeconds *int
	gridK1               *string
	gridBCmd             *string
	gridBDesc            *string
	gridBKeys            *string
	gridBTags            *string
	gridWCmd             *string
	gridWDesc            *string
	gridWKeys            *string
	gridWTags            *string
	gridMinIDF           *string
	gridTopTerms         *string
}

func main() {
	opts := parseCLIOptions()

	if opts.k <= 0 {
		log.Fatal("k must be > 0")
	}
	if opts.limit <= 0 {
		log.Fatal("limit must be > 0")
	}

	dataset, err := loadEvalDataset(opts.evalPath)
	if err != nil {
		log.Fatalf("failed to load eval dataset: %v", err)
	}
	if len(dataset.Queries) == 0 {
		log.Fatal("evaluation dataset has no queries")
	}

	db, err := database.LoadDatabase(opts.dbPath)
	if err != nil {
		log.Fatalf("failed to load database: %v", err)
	}

	if opts.sweepMode {
		runSweep(
			db,
			dataset.Queries,
			opts.limit,
			opts.k,
			opts.useNLP,
			opts.useFuzzy,
			opts.allPlatforms,
			opts.baseConfig,
			opts.sweepTopN,
			opts.sweepWorkers,
			opts.sweepProgressSeconds,
			opts.gridK1,
			opts.gridBCmd,
			opts.gridBDesc,
			opts.gridBKeys,
			opts.gridBTags,
			opts.gridWCmd,
			opts.gridWDesc,
			opts.gridWKeys,
			opts.gridWTags,
			opts.gridMinIDF,
			opts.gridTopTerms,
		)
		return
	}

	options := searchOptionsFromConfig(opts.baseConfig, opts.limit, opts.useNLP, opts.useFuzzy, opts.allPlatforms)
	results := runEvaluation(db, dataset.Queries, opts.k, options)
	summary := printSummary(results, opts.k, opts.verbose)

	failures := checkThresholdFailures(summary, opts.k, opts.minTop1, opts.minHitK, opts.minMRR, opts.minNDCG)
	if len(failures) > 0 {
		fmt.Println("\nThreshold check failed:")
		for _, f := range failures {
			fmt.Printf("- %s\n", f)
		}
		os.Exit(2)
	}
}

func parseCLIOptions() cliOptions {
	sweepFlags := defineSweepFlags()

	var (
		dbPath       = flag.String("db", "assets/commands.yml", "Path to commands database YAML")
		evalPath     = flag.String("eval", "assets/eval_queries.yaml", "Path to evaluation query set YAML")
		limit        = flag.Int("limit", 10, "Search limit per query")
		k            = flag.Int("k", 3, "Evaluation cutoff for Hit@K and NDCG@K")
		useNLP       = flag.Bool("nlp", true, "Enable NLP enhancements")
		useFuzzy     = flag.Bool("fuzzy", true, "Enable fuzzy fallback")
		allPlatforms = flag.Bool("all-platforms", true, "Search across all platforms")
		verbose      = flag.Bool("verbose", true, "Print per-query details")
		minTop1      = flag.Float64("min-top1", -1, "Fail if Top1 is below this value (0-1); disabled when < 0")
		minHitK      = flag.Float64("min-hitk", -1, "Fail if Hit@K is below this value (0-1); disabled when < 0")
		minMRR       = flag.Float64("min-mrr", -1, "Fail if MRR is below this value (0-1); disabled when < 0")
		minNDCG      = flag.Float64("min-ndcg", -1, "Fail if NDCG@K is below this value (0-1); disabled when < 0")

		bm25K1                 = flag.Float64("bm25-k1", 1.2, "BM25F k1 override")
		bm25BCmd               = flag.Float64("bm25-b-cmd", 0.75, "BM25F b (command field)")
		bm25BDesc              = flag.Float64("bm25-b-desc", 0.75, "BM25F b (description field)")
		bm25BKeys              = flag.Float64("bm25-b-keys", 0.7, "BM25F b (keywords field)")
		bm25BTags              = flag.Float64("bm25-b-tags", 0.7, "BM25F b (tags field)")
		bm25WCmd               = flag.Float64("bm25-w-cmd", 3.5, "BM25F weight (command field)")
		bm25WDesc              = flag.Float64("bm25-w-desc", 1.0, "BM25F weight (description field)")
		bm25WKeys              = flag.Float64("bm25-w-keys", 2.0, "BM25F weight (keywords field)")
		bm25WTags              = flag.Float64("bm25-w-tags", 1.2, "BM25F weight (tags field)")
		bm25MinIDF             = flag.Float64("bm25-min-idf", 0.0, "BM25F minimum IDF threshold")
		topTermsCap            = flag.Int("top-terms-cap", 10, "Top-IDF terms cap used for long query scoring")
		familyExpansionProfile = flag.String(
			"family-expansion-profile",
			"custom",
			"Family expansion profile: off, safe, experimental, custom",
		)
		enableFamilyExpansion     = flag.Bool("family-expansion", false, "Enable Phase 2 corpus-native family expansion")
		familyExpansionMaxBases   = flag.Int("family-expansion-max-bases", 3, "Max learned command bases considered for expansion")
		familyExpansionMaxTerms   = flag.Int("family-expansion-max-terms", 4, "Max expansion terms appended")
		familyExpansionClarityMax = flag.Float64(
			"family-expansion-clarity-max",
			0.55,
			"Expand only when family clarity/confidence is <= threshold",
		)
		familyExpansionBlendWeight = flag.Float64("family-expansion-blend-weight", 0.25, "Additive blend weight for expansion channel")

		disableBigrams   = flag.Bool("disable-bigrams", false, "Disable command/keyword bigram channel")
		disableCharNGram = flag.Bool("disable-char-ngram", false, "Disable character n-gram channel")
	)
	flag.Parse()

	baseConfig := bm25EvalConfig{
		K1:                         *bm25K1,
		B:                          database.BM25FieldValues{Cmd: *bm25BCmd, Desc: *bm25BDesc, Keys: *bm25BKeys, Tags: *bm25BTags},
		W:                          database.BM25FieldValues{Cmd: *bm25WCmd, Desc: *bm25WDesc, Keys: *bm25WKeys, Tags: *bm25WTags},
		MinIDF:                     *bm25MinIDF,
		TopTermsCap:                *topTermsCap,
		EnableFamilyExpansion:      *enableFamilyExpansion,
		FamilyExpansionMaxBases:    *familyExpansionMaxBases,
		FamilyExpansionMaxTerms:    *familyExpansionMaxTerms,
		FamilyExpansionClarityMax:  *familyExpansionClarityMax,
		FamilyExpansionBlendWeight: *familyExpansionBlendWeight,
		DisableBigrams:             *disableBigrams,
		DisableCharNGram:           *disableCharNGram,
	}
	applyFamilyExpansionProfile(&baseConfig, strings.ToLower(strings.TrimSpace(*familyExpansionProfile)))

	return cliOptions{
		dbPath:               *dbPath,
		evalPath:             *evalPath,
		limit:                *limit,
		k:                    *k,
		useNLP:               *useNLP,
		useFuzzy:             *useFuzzy,
		allPlatforms:         *allPlatforms,
		verbose:              *verbose,
		minTop1:              *minTop1,
		minHitK:              *minHitK,
		minMRR:               *minMRR,
		minNDCG:              *minNDCG,
		baseConfig:           baseConfig,
		sweepMode:            *sweepFlags.sweepMode,
		sweepTopN:            *sweepFlags.sweepTopN,
		sweepWorkers:         *sweepFlags.sweepWorkers,
		sweepProgressSeconds: *sweepFlags.sweepProgressSeconds,
		gridK1:               *sweepFlags.gridK1,
		gridBCmd:             *sweepFlags.gridBCmd,
		gridBDesc:            *sweepFlags.gridBDesc,
		gridBKeys:            *sweepFlags.gridBKeys,
		gridBTags:            *sweepFlags.gridBTags,
		gridWCmd:             *sweepFlags.gridWCmd,
		gridWDesc:            *sweepFlags.gridWDesc,
		gridWKeys:            *sweepFlags.gridWKeys,
		gridWTags:            *sweepFlags.gridWTags,
		gridMinIDF:           *sweepFlags.gridMinIDF,
		gridTopTerms:         *sweepFlags.gridTopTerms,
	}
}

func defineSweepFlags() sweepFlagRefs {
	return sweepFlagRefs{
		sweepMode:            flag.Bool("sweep", false, "Run BM25 grid sweep instead of single evaluation"),
		sweepTopN:            flag.Int("sweep-topn", 10, "Top N configs to print in sweep mode"),
		sweepWorkers:         flag.Int("sweep-workers", runtime.NumCPU(), "Number of concurrent workers for sweep evaluation"),
		sweepProgressSeconds: flag.Int("sweep-progress-seconds", 10, "Progress log interval in seconds during sweep mode"),
		gridK1:               flag.String("grid-k1", "", "Comma-separated k1 values for sweep mode"),
		gridBCmd:             flag.String("grid-b-cmd", "", "Comma-separated b(cmd) values for sweep mode"),
		gridBDesc:            flag.String("grid-b-desc", "", "Comma-separated b(desc) values for sweep mode"),
		gridBKeys:            flag.String("grid-b-keys", "", "Comma-separated b(keys) values for sweep mode"),
		gridBTags:            flag.String("grid-b-tags", "", "Comma-separated b(tags) values for sweep mode"),
		gridWCmd:             flag.String("grid-w-cmd", "", "Comma-separated w(cmd) values for sweep mode"),
		gridWDesc:            flag.String("grid-w-desc", "", "Comma-separated w(desc) values for sweep mode"),
		gridWKeys:            flag.String("grid-w-keys", "", "Comma-separated w(keys) values for sweep mode"),
		gridWTags:            flag.String("grid-w-tags", "", "Comma-separated w(tags) values for sweep mode"),
		gridMinIDF:           flag.String("grid-min-idf", "", "Comma-separated minIDF values for sweep mode"),
		gridTopTerms:         flag.String("grid-top-terms", "", "Comma-separated TopTermsCap values for sweep mode"),
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

func runEvaluation(db *database.Database, queries []evalQuery, k int, options database.SearchOptions) []perQueryResult {
	out := make([]perQueryResult, 0, len(queries))

	for _, q := range queries {
		res := db.SearchUniversal(q.Query, options)

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
			Slice:       normalizeSlice(q.Slice),
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

func searchOptionsFromConfig(cfg bm25EvalConfig, limit int, useNLP, useFuzzy, allPlatforms bool) database.SearchOptions {
	k1 := cfg.K1
	minIDF := cfg.MinIDF
	b := cfg.B
	w := cfg.W

	return database.SearchOptions{
		Limit:                      limit,
		UseNLP:                     useNLP,
		UseFuzzy:                   useFuzzy,
		AllPlatforms:               allPlatforms,
		TopTermsCap:                cfg.TopTermsCap,
		EnableFamilyExpansion:      cfg.EnableFamilyExpansion,
		FamilyExpansionMaxBases:    cfg.FamilyExpansionMaxBases,
		FamilyExpansionMaxTerms:    cfg.FamilyExpansionMaxTerms,
		FamilyExpansionClarityMax:  cfg.FamilyExpansionClarityMax,
		FamilyExpansionBlendWeight: cfg.FamilyExpansionBlendWeight,
		DisableBigrams:             cfg.DisableBigrams,
		DisableCharNGram:           cfg.DisableCharNGram,
		BM25Overrides: &database.BM25Overrides{
			K1:     &k1,
			B:      &b,
			W:      &w,
			MinIDF: &minIDF,
		},
	}
}

func runSweep(
	db *database.Database,
	queries []evalQuery,
	limit, k int,
	useNLP, useFuzzy, allPlatforms bool,
	base bm25EvalConfig,
	topN int,
	workers int,
	progressSeconds int,
	gridK1, gridBCmd, gridBDesc, gridBKeys, gridBTags, gridWCmd, gridWDesc, gridWKeys, gridWTags, gridMinIDF, gridTopTerms string,
) {
	configs := buildSweepConfigs(
		base,
		gridK1,
		gridBCmd,
		gridBDesc,
		gridBKeys,
		gridBTags,
		gridWCmd,
		gridWDesc,
		gridWKeys,
		gridWTags,
		gridMinIDF,
		gridTopTerms,
	)

	if len(configs) == 0 {
		log.Fatal("sweep produced no configurations")
	}

	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(configs) {
		workers = len(configs)
	}
	if progressSeconds <= 0 {
		progressSeconds = 10
	}

	results := evaluateConfigsParallel(
		db,
		queries,
		k,
		limit,
		useNLP,
		useFuzzy,
		allPlatforms,
		configs,
		workers,
		time.Duration(progressSeconds)*time.Second,
	)

	sort.Slice(results, func(i, j int) bool {
		if results[i].Summary.NDCG != results[j].Summary.NDCG {
			return results[i].Summary.NDCG > results[j].Summary.NDCG
		}
		if results[i].Summary.HitK != results[j].Summary.HitK {
			return results[i].Summary.HitK > results[j].Summary.HitK
		}
		return results[i].Summary.Top1 > results[j].Summary.Top1
	})

	if topN <= 0 {
		topN = 10
	}
	if topN > len(results) {
		topN = len(results)
	}

	fmt.Printf("Evaluated %d BM25 configurations using %d workers\n", len(results), workers)
	fmt.Printf("Top %d by NDCG@%d:\n", topN, k)
	for i := 0; i < topN; i++ {
		r := results[i]
		fmt.Printf(
			"%2d) ndcg=%.4f hit@%d=%.4f top1=%.4f mrr=%.4f | k1=%.2f minIDF=%.3f topTerms=%d | b=[%.2f %.2f %.2f %.2f] w=[%.2f %.2f %.2f %.2f]\n",
			i+1,
			r.Summary.NDCG,
			k,
			r.Summary.HitK,
			r.Summary.Top1,
			r.Summary.MRR,
			r.Config.K1,
			r.Config.MinIDF,
			r.Config.TopTermsCap,
			r.Config.B.Cmd,
			r.Config.B.Desc,
			r.Config.B.Keys,
			r.Config.B.Tags,
			r.Config.W.Cmd,
			r.Config.W.Desc,
			r.Config.W.Keys,
			r.Config.W.Tags,
		)
	}
}

func evaluateConfigsParallel(
	db *database.Database,
	queries []evalQuery,
	k, limit int,
	useNLP, useFuzzy, allPlatforms bool,
	configs []bm25EvalConfig,
	workers int,
	progressInterval time.Duration,
) []sweepResult {
	results := make([]sweepResult, len(configs))
	jobs := make(chan int, workers*2)
	var wg sync.WaitGroup
	var completed int64
	start := time.Now()

	if progressInterval > 0 {
		go func(total int) {
			ticker := time.NewTicker(progressInterval)
			defer ticker.Stop()
			for range ticker.C {
				done := int(atomic.LoadInt64(&completed))
				if done >= total {
					return
				}
				pct := 100.0 * float64(done) / float64(total)
				fmt.Printf("Sweep progress: %d/%d (%.1f%%), elapsed=%s\n", done, total, pct, time.Since(start).Truncate(time.Second))
			}
		}(len(configs))
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				cfg := configs[idx]
				opts := searchOptionsFromConfig(cfg, limit, useNLP, useFuzzy, allPlatforms)
				summary := runEvaluationSummary(db, queries, k, opts)
				results[idx] = sweepResult{Config: cfg, Summary: summary}
				atomic.AddInt64(&completed, 1)
			}
		}()
	}

	for i := range configs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	fmt.Printf("Sweep completed in %s\n", time.Since(start).Truncate(time.Second))
	return results
}

func buildSweepConfigs(
	base bm25EvalConfig,
	gridK1, gridBCmd, gridBDesc, gridBKeys, gridBTags, gridWCmd, gridWDesc, gridWKeys, gridWTags, gridMinIDF, gridTopTerms string,
) []bm25EvalConfig {
	k1Values := parseFloatGridOrDefault(gridK1, base.K1)
	bCmdValues := parseFloatGridOrDefault(gridBCmd, base.B.Cmd)
	bDescValues := parseFloatGridOrDefault(gridBDesc, base.B.Desc)
	bKeysValues := parseFloatGridOrDefault(gridBKeys, base.B.Keys)
	bTagsValues := parseFloatGridOrDefault(gridBTags, base.B.Tags)
	wCmdValues := parseFloatGridOrDefault(gridWCmd, base.W.Cmd)
	wDescValues := parseFloatGridOrDefault(gridWDesc, base.W.Desc)
	wKeysValues := parseFloatGridOrDefault(gridWKeys, base.W.Keys)
	wTagsValues := parseFloatGridOrDefault(gridWTags, base.W.Tags)
	minIDFValues := parseFloatGridOrDefault(gridMinIDF, base.MinIDF)
	topTermsValues := parseIntGridOrDefault(gridTopTerms, base.TopTermsCap)

	configs := make([]bm25EvalConfig, 0, 128)
	for _, k1 := range k1Values {
		for _, bCmd := range bCmdValues {
			for _, bDesc := range bDescValues {
				for _, bKeys := range bKeysValues {
					for _, bTags := range bTagsValues {
						for _, wCmd := range wCmdValues {
							for _, wDesc := range wDescValues {
								for _, wKeys := range wKeysValues {
									for _, wTags := range wTagsValues {
										for _, minIDF := range minIDFValues {
											for _, cap := range topTermsValues {
												configs = append(configs, bm25EvalConfig{
													K1:                         k1,
													B:                          database.BM25FieldValues{Cmd: bCmd, Desc: bDesc, Keys: bKeys, Tags: bTags},
													W:                          database.BM25FieldValues{Cmd: wCmd, Desc: wDesc, Keys: wKeys, Tags: wTags},
													MinIDF:                     minIDF,
													TopTermsCap:                cap,
													EnableFamilyExpansion:      base.EnableFamilyExpansion,
													FamilyExpansionMaxBases:    base.FamilyExpansionMaxBases,
													FamilyExpansionMaxTerms:    base.FamilyExpansionMaxTerms,
													FamilyExpansionClarityMax:  base.FamilyExpansionClarityMax,
													FamilyExpansionBlendWeight: base.FamilyExpansionBlendWeight,
													DisableBigrams:             base.DisableBigrams,
													DisableCharNGram:           base.DisableCharNGram,
												})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return configs
}

func runEvaluationSummary(db *database.Database, queries []evalQuery, k int, options database.SearchOptions) metricsSummary {
	if len(queries) == 0 {
		return metricsSummary{}
	}

	total := float64(len(queries))
	top1 := 0.0
	topK := 0.0
	mrr := 0.0
	meanNDCG := 0.0

	for _, q := range queries {
		res := db.SearchUniversal(q.Query, options)

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

		bestRank := firstRelevantRank(pred, relevant)
		if bestRank > 0 {
			mrr += 1.0 / float64(bestRank)
			if bestRank == 1 {
				top1++
			}
			if bestRank <= k {
				topK++
			}
		}

		meanNDCG += ndcgAtK(pred, relevant, k)
	}

	return metricsSummary{
		Count: total,
		Top1:  top1 / total,
		HitK:  topK / total,
		MRR:   mrr / total,
		NDCG:  meanNDCG / total,
	}
}

func parseFloatGridOrDefault(raw string, def float64) []float64 {
	if strings.TrimSpace(raw) == "" {
		return []float64{def}
	}
	parts := strings.Split(raw, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			log.Fatalf("invalid float grid value %q: %v", p, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return []float64{def}
	}
	return out
}

func parseIntGridOrDefault(raw string, def int) []int {
	if strings.TrimSpace(raw) == "" {
		return []int{def}
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("invalid int grid value %q: %v", p, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return []int{def}
	}
	return out
}

func normalizeSlice(slice string) string {
	s := strings.TrimSpace(strings.ToLower(slice))
	if s == "" {
		return "unspecified"
	}
	return s
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

func printSummary(results []perQueryResult, k int, verbose bool) metricsSummary {
	total := float64(len(results))
	if total == 0 {
		fmt.Println("No evaluation results")
		return metricsSummary{}
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
			fmt.Printf("- [%s] %s\n", r.Slice, r.Query)
			fmt.Printf("  best_rank=%s top1=%t hit@%d=%t rr=%.3f ndcg@%d=%.3f\n", bestRank, r.Top1Hit, k, r.TopKHit, r.RR, k, r.NDCG)
			if len(r.TopCommands) > 0 {
				fmt.Printf("  top_commands=%s\n", strings.Join(r.TopCommands, " | "))
			}
		}
	}

	summary := metricsSummary{
		Top1:  top1 / total,
		HitK:  topK / total,
		MRR:   mrr / total,
		NDCG:  meanNDCG / total,
		Count: total,
	}

	fmt.Println("\nEvaluation summary:")
	fmt.Printf("- queries: %.0f\n", summary.Count)
	fmt.Printf("- Top1: %.3f\n", summary.Top1)
	fmt.Printf("- Hit@%d: %.3f\n", k, summary.HitK)
	fmt.Printf("- MRR: %.3f\n", summary.MRR)
	fmt.Printf("- NDCG@%d: %.3f\n", k, summary.NDCG)

	printSliceSummary(results, k)

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
			fmt.Printf("- [%s] %s (rr=%.3f ndcg@%d=%.3f)\n", r.Slice, r.Query, r.RR, k, r.NDCG)
		}
	}

	return summary
}

func printSliceSummary(results []perQueryResult, k int) {
	bySlice := make(map[string][]perQueryResult)
	for _, r := range results {
		bySlice[r.Slice] = append(bySlice[r.Slice], r)
	}

	if len(bySlice) == 0 {
		return
	}

	slices := make([]string, 0, len(bySlice))
	for s := range bySlice {
		slices = append(slices, s)
	}
	sort.Strings(slices)

	fmt.Println("\nPer-slice metrics:")
	for _, sliceName := range slices {
		metrics := summarizeMetrics(bySlice[sliceName])
		fmt.Printf("- %s (n=%.0f): top1=%.3f hit@%d=%.3f mrr=%.3f ndcg@%d=%.3f\n",
			sliceName, metrics.Count, metrics.Top1, k, metrics.HitK, metrics.MRR, k, metrics.NDCG)
	}
}

func summarizeMetrics(results []perQueryResult) metricsSummary {
	total := float64(len(results))
	if total == 0 {
		return metricsSummary{}
	}

	top1 := 0.0
	topK := 0.0
	mrr := 0.0
	ndcg := 0.0
	for _, r := range results {
		if r.Top1Hit {
			top1++
		}
		if r.TopKHit {
			topK++
		}
		mrr += r.RR
		ndcg += r.NDCG
	}

	return metricsSummary{
		Count: total,
		Top1:  top1 / total,
		HitK:  topK / total,
		MRR:   mrr / total,
		NDCG:  ndcg / total,
	}
}

func checkThresholdFailures(summary metricsSummary, k int, minTop1, minHitK, minMRR, minNDCG float64) []string {
	failures := make([]string, 0, 4)
	if minTop1 >= 0 && summary.Top1 < minTop1 {
		failures = append(failures, fmt.Sprintf("Top1 %.3f is below min-top1 %.3f", summary.Top1, minTop1))
	}
	if minHitK >= 0 && summary.HitK < minHitK {
		failures = append(failures, fmt.Sprintf("Hit@%d %.3f is below min-hitk %.3f", k, summary.HitK, minHitK))
	}
	if minMRR >= 0 && summary.MRR < minMRR {
		failures = append(failures, fmt.Sprintf("MRR %.3f is below min-mrr %.3f", summary.MRR, minMRR))
	}
	if minNDCG >= 0 && summary.NDCG < minNDCG {
		failures = append(failures, fmt.Sprintf("NDCG@%d %.3f is below min-ndcg %.3f", k, summary.NDCG, minNDCG))
	}
	return failures
}

package database

import (
    "math"
    "sort"
    "strings"
    "unicode"

    "github.com/Vedant9500/WTF/internal/nlp"
)

// universalIndex implements a scalable, data-driven inverted index with BM25F scoring.
// It avoids hardcoded domain rules and scales with database size.
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
    cmd int
    desc int
    keys int
    tags int
}

type docLens struct {
    cmd int
    desc int
    keys int
    tags int
}

type docLensF struct {
    cmd float64
    desc float64
    keys float64
    tags float64
}

type bm25fParams struct {
    k1 float64
    b  docLensF     // b per field
    w  docLensF     // field weights
    minIDF float64  // minimum idf to count
}

// tokenizer for index/query
func normalizeAndTokenize(s string) []string {
    if s == "" {
        return nil
    }
    // Lowercase and split on non letters/numbers
    lower := strings.ToLower(s)
    words := strings.FieldsFunc(lower, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })

    out := make([]string, 0, len(words))
    for _, w := range words {
        if len(w) < 2 { // drop very short tokens
            continue
        }
        if stopWords[w] {
            continue
        }
        out = append(out, w)
    }
    return out
}

var stopWords = map[string]bool{
    "the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
    "in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
    "with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
    "be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
    "does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
    "this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
    "you": true, "your": true, "all": true, "any": true, "can": true, "from": true,
    "not": true, "no": true, "if": true, "when": true, "where": true, "how": true,
    "what": true, "which": true, "who": true, "why": true, "use": true, "used": true,
    "using": true,
}

func defaultParams() bm25fParams {
    return bm25fParams{
        k1: 1.2,
        b:  docLensF{cmd: 0.75, desc: 0.75, keys: 0.7, tags: 0.7},
        w:  docLensF{cmd: 2.5, desc: 1.0, keys: 1.8, tags: 1.2},
        minIDF: 0.0,
    }
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
        cmd := &db.Commands[i]

        // prefer cached lowercase fields if available
        cmdText := cmd.Command
        if cmd.CommandLower != "" { cmdText = cmd.CommandLower }
        descText := cmd.Description
        if cmd.DescriptionLower != "" { descText = cmd.DescriptionLower }

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
        idx.docLens[i] = docLens{cmd: len(cmdTokens), desc: len(descTokens), keys: len(keysTokens), tags: len(tagsTokens)}

        // term frequencies
        tf := make(map[string]fieldTF)
        inc := func(tok string, f string) {
            entry := tf[tok]
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
            tf[tok] = entry
        }

        for _, t := range cmdTokens { inc(t, "cmd") }
        for _, t := range descTokens { inc(t, "desc") }
        for _, t := range keysTokens { inc(t, "keys") }
        for _, t := range tagsTokens { inc(t, "tags") }

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
    idx.avgLen = docLensF{cmd: float64(sum.cmd)/n, desc: float64(sum.desc)/n, keys: float64(sum.keys)/n, tags: float64(sum.tags)/n}

    // build postings
    for docID, tf := range perDocTFs {
        for term, ftf := range tf {
            idx.postings[term] = append(idx.postings[term], posting{docID: docID, tf: ftf})
        }
    }

    db.uIndex = idx
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
    // If NLP enabled, enhance the query terms with actions/targets/keywords
    if options.UseNLP {
        processor := nlp.NewQueryProcessor()
        pq = processor.ProcessQuery(query)
        enh := pq.GetEnhancedKeywords()
        if len(enh) > 0 {
            terms = enh
        }
    }
    if len(terms) == 0 {
        return nil
    }

    idx := db.uIndex
    scores := make(map[int]float64, len(db.Commands)/4)
    currentPlatform := getCurrentPlatform()

    // Prepare per-term boosts (context + NLP action/target emphasis)
    termBoost := map[string]float64{}
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
        for _, p := range postings {
            doc := &db.Commands[p.docID]

            // Platform filtering (fast, non-hardcoded)
            if len(doc.Platform) > 0 {
                if !isPlatformCompatible(doc.Platform, currentPlatform) && !isCrossPlatformTool(doc.Command) {
                    continue
                }
            }

            // Optional pipeline-only filter
            if options.PipelineOnly && !isPipelineCommand(doc) {
                continue
            }

            s := scores[p.docID]
            s += (idf * boost) * idx.termBM25F(p.docID, p.tf)
            scores[p.docID] = s
        }
    }

    if len(scores) == 0 {
        return nil
    }

    // Collect and optionally apply pipeline boost
    results := make([]SearchResult, 0, min(len(scores), options.Limit*3))
    for docID, score := range scores {
        cmd := &db.Commands[docID]
        // Apply intent-based boost if NLP is active
        if pq != nil {
            score *= calculateIntentBoost(cmd, pq)
        }
        if isPipelineCommand(cmd) && options.PipelineBoost > 0 {
            score *= options.PipelineBoost
        }
        results = append(results, SearchResult{Command: cmd, Score: score})
    }

    // Sort and limit
    sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
    if len(results) > options.Limit {
        results = results[:options.Limit]
    }
    return results
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

func bm25IDF(N, df int) float64 {
    // Okapi BM25 idf with 0.5 adjustments
    return math.Log((float64(N)-float64(df)+0.5)/(float64(df)+0.5) + 1)
}

func isPlatformCompatible(platforms []string, current string) bool {
    for _, p := range platforms {
        if strings.EqualFold(p, "cross-platform") || strings.EqualFold(p, current) {
            return true
        }
    }
    return false
}

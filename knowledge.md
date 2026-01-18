# Project: WTF (What's The Function) - System Knowledge

> **Generated:** January 18, 2026  
> **Version:** 1.2.0  
> **Purpose:** AI-consumable knowledge base for instant codebase understanding

---

## 1. High-Level Architecture

### Core Loop
User query enters via `cmd/wtf/main.go` → `cli.Execute()`. The CLI validates input, detects project context (Git/Docker/Node/etc.), loads the YAML command database with BM25F inverted index, executes multi-stage search (NLP enhancement → BM25F scoring → Cascading Boost → TF-IDF rerank), and outputs results in list/table/JSON format.

### Package Dependency Graph
```
cmd/wtf/main.go
    └── cli/root.go (Cobra CLI)
            ├── cli/search.go ─────────────┐
            ├── cli/wizard.go              │
            ├── cli/history.go             │
            ├── cli/alias.go               │
            ├── cli/save.go                │
            └── cli/pipeline.go            │
                                           ▼
                    ┌──────────────────────────────────────┐
                    │         SEARCH ORCHESTRATION         │
                    │  validation → context → database     │
                    └──────────────────────────────────────┘
                           │         │           │
                           ▼         ▼           ▼
                    ┌─────────┐ ┌─────────┐ ┌──────────┐
                    │validation│ │ context │ │ database │
                    └─────────┘ └─────────┘ └──────────┘
                                     │           │
                                     │     ┌─────┴─────┐
                                     │     ▼           ▼
                                     │  ┌─────┐    ┌───────┐
                                     │  │ nlp │    │ cache │
                                     │  └─────┘    └───────┘
                                     │
                              ┌──────┴──────┐
                              │  recovery   │
                              │  errors     │
                              │  history    │
                              └─────────────┘
```

### Key Packages

| Package | Responsibility | Key Structs/Functions |
|---------|----------------|----------------------|
| `cli` | Cobra CLI commands, search orchestration | `searchCmd`, `Execute()`, `rootCmd` |
| `database` | Command storage, BM25F search, TF-IDF | `Database`, `Command`, `SearchResult`, `SearchUniversal()` |
| `nlp` | Query processing, intent detection, synonyms | `QueryProcessor`, `ProcessedQuery`, `TFIDFSearcher` |
| `context` | Project type detection, context boosts | `Analyzer`, `Context`, `ProjectType` |
| `config` | App configuration, path resolution | `Config`, `DefaultConfig()`, `GetDatabasePath()` |
| `history` | Search history persistence (JSON) | `SearchHistory`, `SearchEntry` |
| `cache` | TTL-based in-memory cache | `Cache`, `Item` |
| `recovery` | Database fallback strategies | `DatabaseRecovery`, `SearchRecovery` |
| `errors` | Structured errors with user messages | `AppError`, `DatabaseError`, `SearchError` |
| `validation` | Input sanitization | `ValidateQuery()`, `ValidateLimit()` |
| `embedding` | GloVe word vectors (optional) | `Index`, `LoadWordVectors()` |
| `constants` | All scoring weights & thresholds | Score constants, limits, TTLs |
| `version` | Build-time version info | `Version`, `GitHash`, `Build` |

---

## 2. The Brain (Search & NLP)

### Search Pipeline (4 Stages)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  STAGE 1: NLP ENHANCEMENT                                               │
│  ─────────────────────────                                              │
│  QueryProcessor.ProcessQuery(query)                                     │
│    → Clean query (remove special chars)                                 │
│    → Extract Actions: ["compress", "find", "create", ...]              │
│    → Extract Targets: ["file", "directory", "package", ...]            │
│    → Detect Intent: IntentFind|Create|Delete|Modify|View|Install|...   │
│    → Expand with synonyms (limited to 1 best synonym per word)         │
│    → GetEnhancedKeywords() → merged search terms                       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STAGE 2: BM25F INVERTED INDEX SEARCH                                   │
│  ────────────────────────────────────                                   │
│  universalIndex.postings[term] → list of (docID, fieldTF)              │
│                                                                         │
│  Score = Σ (IDF × boost × termBM25F)                                   │
│                                                                         │
│  Field Weights:     │  BM25 Params:                                    │
│    cmd:  3.5        │    k1: 1.2                                       │
│    desc: 1.0        │    b:  0.75 (per field)                          │
│    keys: 2.0        │                                                   │
│    tags: 1.2        │                                                   │
│                                                                         │
│  IDF = log((N - df + 0.5) / (df + 0.5) + 1)                            │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STAGE 3: CASCADING BOOST                                               │
│  ────────────────────────                                               │
│  Boost weights applied based on token type match:                       │
│                                                                         │
│    Action terms (compress, find, delete):     +3.0x                    │
│    Context terms (git, docker, npm, kubectl): +2.5x                    │
│    Target terms (file, directory, package):   +2.0x                    │
│    Keyword terms (general matches):           +1.5x                    │
│                                                                         │
│  + Intent-specific boosts from intentKeywords map                       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STAGE 4: TF-IDF RERANKING (Top Results Only)                          │
│  ────────────────────────────────────────────                          │
│  Cosine similarity refinement using TFIDFSearcher                       │
│  Combines BM25F score with TF-IDF similarity for final ranking         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Intent Detection Logic

**Location:** `nlp/processor.go` → `detectIntent()`

| Intent | Trigger Words |
|--------|---------------|
| `IntentFind` | find, search, locate, list |
| `IntentView` | show, display, view, see, read, cat |
| `IntentCreate` | create, make, build, generate, new |
| `IntentDelete` | delete, remove, destroy, clean, clear |
| `IntentModify` | modify, change, edit, update, alter |
| `IntentInstall` | install, add, download |
| `IntentRun` | run, execute, start, launch |
| `IntentConfigure` | configure, config, setup, set |

**Intent is probabilistic via keyword matching, NOT ML-based.**

### Context Awareness

**Location:** `context/analyzer.go`

Detects 18+ project types by scanning current directory for marker files:

| Project Type | Detection Files |
|--------------|-----------------|
| Git | `.git/` |
| Docker | `Dockerfile`, `docker-compose.yml` |
| Node.js | `package.json`, `node_modules/`, `yarn.lock` |
| Python | `requirements.txt`, `setup.py`, `pyproject.toml` |
| Go | `go.mod`, `go.sum` |
| Rust | `Cargo.toml` |
| Java | `pom.xml`, `build.gradle` |
| .NET | `*.csproj`, `global.json` |
| Kubernetes | `*k8s*.yaml`, `kustomization.yaml` |
| Terraform | `*.tf`, `*.tfvars` |

**Context Boosts:** Each project type has a boost map (e.g., Git context boosts "git", "commit", "branch" by 1.5-2.0x).

---

## 3. Data & Persistence

### Command Database

**Location:** `assets/commands.yml` (primary), `~/.config/cmd-finder/personal.yml` (user)

**Format:** YAML array with ~3,850 commands

```yaml
- command: "git fetch"
  description: "Download objects and refs from a remote repository"
  keywords: ["git", "fetch", "remote", "download", "repository"]
  niche: "version-control"           # Category/domain
  platform: [linux, macos, windows]  # Supported platforms
  pipeline: false                    # Pipeline-friendly command?
```

**Schema:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | ✓ | The shell command |
| `description` | string | ✓ | Human-readable explanation |
| `keywords` | []string | ✓ | Searchable terms |
| `tags` | []string | | Additional categorization |
| `niche` | string | | Domain (git, docker, system, etc.) |
| `platform` | []string | | linux, macos, windows, cross-platform |
| `pipeline` | bool | | True if commonly piped |

**Lowercased Cache Fields:** At load time, `CommandLower`, `DescriptionLower`, `KeywordsLower`, `TagsLower` are populated for O(1) case-insensitive matching.

### Database Loading Flow

```
LoadDatabase(filename)
    → os.ReadFile()
    → yaml.Unmarshal()
    → Populate lowercase cache fields
    → BuildUniversalIndex()  ← BM25F inverted index
    → buildTFIDFSearcher()   ← TF-IDF for reranking
```

**Merge Logic:** `LoadDatabaseWithPersonal()` loads main DB, then appends personal DB commands, rebuilds indices.

### Data Sources

**Primary:** [TLDR Pages](https://github.com/tldr-pages/tldr), [Cheat/Cheatsheets](https://github.com/cheat/cheatsheets)

**Ingestion Scripts:**
- `scripts/embed_commands.py` - Generate GloVe embeddings for semantic search
- `scripts/prepare_glove.py` - Convert GloVe text to binary format

### History Persistence

**Location:** `~/.config/wtf/search_history.json` (or `UserConfigDir()`)

```json
{
  "entries": [
    {
      "query": "compress files",
      "timestamp": "2026-01-18T10:30:00Z",
      "results_count": 5,
      "context": "go",
      "duration": 45
    }
  ],
  "max_size": 100
}
```

### Configuration

**Location:** `~/.config/cmd-finder/`

**Defaults (from `config.DefaultConfig()`):**
- `DatabasePath`: `assets/commands.yml`
- `PersonalDBPath`: `~/.config/cmd-finder/personal.yml`
- `MaxResults`: 5
- `CacheEnabled`: true

**Path Resolution:** Falls back through: configured path → `/usr/local/share/` → `/usr/share/` → `assets/` → legacy paths

---

## 4. Developer Constraints & Gotchas

### Error Handling Patterns

**Pattern: Structured Errors with User Messages**

```go
// CORRECT: Use AppError with context and suggestions
return errors.NewAppError(errors.ErrorTypeDatabase, "load failed", cause).
    WithUserMessage("Could not load the command database").
    WithContext("path", filepath).
    WithSuggestions("Run 'wtf setup' to reinitialize")

// CORRECT: Use typed error constructors
return errors.NewDatabaseErrorWithContext("read", filename, err)
return errors.NewQueryEmptyError()
return errors.NewQueryTooLongError(len, max)

// INCORRECT: Raw error wrapping
return fmt.Errorf("failed to load: %w", err)  // ❌ No user message
```

**Error Types:** `ErrorTypeDatabase`, `ErrorTypeValidation`, `ErrorTypeSearch`, `ErrorTypeConfig`, `ErrorTypePermission`

**Recovery Pattern:** `recovery.DatabaseRecovery` attempts: retry → embedded DB → backup → minimal DB

### Concurrency Patterns

**Pattern: sync.RWMutex for Cache**

```go
// Cache uses RWMutex for thread-safe read/write
type Cache struct {
    items map[string]*Item
    mutex sync.RWMutex  // ← RWMutex, not Mutex
    // ...
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()      // ← Read lock
    defer c.mutex.RUnlock()
    // ...
}

func (c *Cache) Set(key string, value interface{}) {
    c.mutex.Lock()       // ← Write lock
    defer c.mutex.Unlock()
    // ...
}
```

**Pattern: WaitGroup for Cleanup**

```go
// Cache cleanup uses WaitGroup for graceful shutdown
c.wg.Add(1)
go c.autoCleanup()

func (c *Cache) Stop() {
    close(c.stopCleanup)
    c.wg.Wait()  // ← Wait for goroutine to finish
}
```

**NO channels used for main data flow** - all search is synchronous.

### Do's & Don'ts

| ✅ DO | ❌ DON'T |
|-------|----------|
| Use `errors.NewAppError()` with user messages | Use `fmt.Errorf()` for user-facing errors |
| Validate input via `validation.ValidateQuery()` | Trust raw user input |
| Use constants from `constants/constants.go` | Hardcode scoring weights |
| Populate lowercase cache fields on load | Compare strings case-sensitively |
| Use `utils.Min()`/`utils.Max()` | Use raw conditionals for min/max |
| Return `nil, nil` for optional missing files | Fail hard on optional file absence |
| Use `SearchUniversal()` for new code | Use deprecated `Search()` or `SearchWithOptions()` |

### Adding a New Command Source

1. Add commands to `assets/commands.yml` following the schema
2. Ensure all required fields: `command`, `description`, `keywords`
3. Run `go build` - indices are built at load time
4. (Optional) Run `scripts/embed_commands.py` if using semantic search

### Adding a New Project Context

1. Add constant to `context/analyzer.go`:
   ```go
   ProjectTypeMyTool ProjectType = "mytool"
   ```

2. Add boost map to `projectBoosts`:
   ```go
   ProjectTypeMyTool: {
       "mytool": 2.0, "keyword1": 1.5, "keyword2": 1.3,
   },
   ```

3. Add detection in `analyzeFile()`:
   ```go
   func (a *Analyzer) checkMyTool(filename string, ctx *Context) {
       if filename == "mytool.config" {
           ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeMyTool)
       }
   }
   ```

### Build System

**Makefile Targets:**

| Target | Description |
|--------|-------------|
| `make build` | Build binary to `build/wtf` |
| `make build-release` | Optimized build with `-s -w` flags |
| `make build-all` | Cross-compile for linux/darwin/windows (amd64/arm64) |
| `make test` | Run all tests |
| `make quality` | Run linters and formatters |
| `make clean` | Remove build artifacts |

**Build Flags (ldflags):**
```
-X github.com/Vedant9500/WTF/internal/version.Version=$(VERSION)
-X github.com/Vedant9500/WTF/internal/version.GitHash=$(GIT_HASH)
-X github.com/Vedant9500/WTF/internal/version.Build=$(BUILD_TIME)
```

**Windows:** Use `build.bat build` instead of Make.

---

## 5. Critical Constants Reference

**Location:** `internal/constants/constants.go`

### Scoring Weights

| Constant | Value | Usage |
|----------|-------|-------|
| `ScoreDirectCommandMatch` | 15.0 | Exact command name match |
| `ScoreCommandMatch` | 10.0 | Partial command match |
| `ScoreDescriptionMatch` | 6.0 | Description word match |
| `ScoreKeywordExactMatch` | 4.0 | Exact keyword match |
| `ScoreDomainSpecificMatch` | 12.0 | Niche/domain match |
| `ExactCommandMatchMultiplier` | 2.0 | Boost for exact matches |

### BM25F Weights (in `search_universal.go`)

| Field | Weight | b Parameter |
|-------|--------|-------------|
| cmd | 3.5 | 0.75 |
| desc | 1.0 | 0.75 |
| keys | 2.0 | 0.70 |
| tags | 1.2 | 0.70 |

**k1 = 1.2** (term frequency saturation)

### Limits

| Constant | Value |
|----------|-------|
| `DefaultSearchLimit` | 5 |
| `MaxQueryLength` | 1000 chars |
| `DefaultHistorySize` | 100 entries |
| `DefaultCacheTTL` | 5 minutes |
| `DefaultCacheCapacity` | 1000 items |

---

## 6. Testing Patterns

**Test Utilities:** `internal/testutil/` provides:
- `CommandBuilder` - Fluent API for creating test commands
- `fixtures.go` - Pre-built test fixtures
- `generators.go` - Random data generators
- `helpers.go` - Common test assertions

**Integration Tests:**
- `cli/integration_test.go` - Full CLI flow tests
- `cli/search_integration_test.go` - Search-specific integration

**Benchmarks:**
- `database/benchmark_search_test.go` - Search performance
- `database/benchmark_test.go` - Database operations

---

## 7. File Quick Reference

| Need to... | Look in... |
|------------|------------|
| Add CLI command | `internal/cli/` + register in `root.go` init() |
| Modify search scoring | `internal/database/search_universal.go`, `cascading_boost.go` |
| Add NLP synonyms | `internal/nlp/processor.go` → `buildSynonyms()` |
| Add intent type | `internal/nlp/processor.go` → `QueryIntent` const + detection |
| Detect new project type | `internal/context/analyzer.go` |
| Change defaults | `internal/config/config.go`, `internal/constants/constants.go` |
| Add error type | `internal/errors/errors.go` |
| Modify validation | `internal/validation/validation.go` |

---

## 8. Extracted Patterns & Idioms

### Error Handling Pattern (Fluent Builder)

**All user-facing errors use the `AppError` builder pattern:**

```go
// Template: NewAppError → WithUserMessage → WithContext → WithSuggestions
return errors.NewAppError(errors.ErrorTypeDatabase, "technical message", cause).
    WithUserMessage("User-friendly explanation").
    WithContext("key", value).
    WithSuggestions(
        "Actionable suggestion 1",
        "Actionable suggestion 2",
    )
```

**Pre-built Error Constructors (prefer these):**
- `NewDatabaseNotFoundError(path, cause)`
- `NewDatabaseParseError(path, cause)`
- `NewDatabasePermissionError(path, cause)`
- `NewQueryEmptyError()`
- `NewQueryTooLongError(len, max)`
- `NewQueryInvalidCharsError(chars)`
- `NewLimitInvalidError(limit, max)`
- `NewSearchFailedError(query, cause)`
- `NewNoResultsError(query, suggestions)`

**Error Extraction:**
```go
userMsg := errors.GetUserFriendlyMessage(err)      // Safe for any error
suggestions := errors.GetErrorSuggestions(err)    // nil for non-AppError
```

### Concurrency Patterns

**Pattern 1: RWMutex for Read-Heavy Caches**
```go
type Cache struct {
    items map[string]*Item
    mutex sync.RWMutex  // RWMutex, not Mutex
}

func (c *Cache) Get(key string) {
    c.mutex.RLock()       // Read lock for gets
    defer c.mutex.RUnlock()
}

func (c *Cache) Set(key string, value interface{}) {
    c.mutex.Lock()        // Write lock for sets
    defer c.mutex.Unlock()
}
```

**Pattern 2: Atomic Operations for Counters**
```go
type Counter struct {
    value int64  // int64 for atomic
}

func (c *Counter) Inc() {
    atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Value() int64 {
    return atomic.LoadInt64(&c.value)
}
```

**Pattern 3: WaitGroup + Channel for Graceful Shutdown**
```go
type Cache struct {
    stopCleanup chan struct{}
    wg          sync.WaitGroup
}

func NewCache() *Cache {
    c := &Cache{stopCleanup: make(chan struct{})}
    c.wg.Add(1)
    go c.backgroundTask()
    return c
}

func (c *Cache) backgroundTask() {
    defer c.wg.Done()
    for {
        select {
        case <-c.stopCleanup:
            return
        case <-time.After(interval):
            // do work
        }
    }
}

func (c *Cache) Stop() {
    close(c.stopCleanup)
    c.wg.Wait()  // Wait for goroutine to finish
}
```

### Recovery Pattern (Strategy Chain)

**Location:** `recovery/recovery.go`

```go
// Database loading uses cascading fallback strategies
func LoadDatabaseWithFallback(primary, personal string) (*Database, error) {
    // Try 1: Load with retry (exponential backoff)
    db, err := loadWithRetry(primary, personal)
    if err == nil { return db, nil }
    
    // Try 2: Embedded minimal database
    db, err = loadEmbeddedDatabase()
    if err == nil { return db, nil }
    
    // Try 3: Backup database
    db, err = loadBackupDatabase(primary)
    if err == nil { return db, nil }
    
    // Try 4: Create minimal database
    db, err = createMinimalDatabase()
    if err == nil { return db, nil }
    
    // All failed
    return nil, allStrategiesFailedError
}
```

**Retry Configuration:**
```go
RetryConfig{
    MaxAttempts:   3,
    BaseDelay:     100ms,
    MaxDelay:      5s,
    BackoffFactor: 2.0,  // Exponential backoff
}
```

### LRU Cache Pattern

**Location:** `cache/lru_cache.go`

```go
type LRUCache struct {
    items     map[string]*list.Element  // O(1) lookup
    evictList *list.List                // Doubly-linked for LRU ordering
    capacity  int
    ttl       time.Duration
}

// Get moves item to front (most recently used)
func (c *LRUCache) Get(key string) {
    if expired { c.removeElement(element); return nil }
    c.evictList.MoveToFront(element)  // LRU update
    return value
}

// Put evicts oldest if at capacity
func (c *LRUCache) Put(key string, value interface{}) {
    element := c.evictList.PushFront(entry)
    if c.evictList.Len() > c.capacity {
        c.evictOldest()  // Remove from back
    }
}
```

### Metrics Pattern (Atomic + Histogram)

**Location:** `metrics/metrics.go`

- **Counter:** `atomic.AddInt64` for thread-safe increment
- **Gauge:** `atomic.StoreInt64` / `atomic.LoadInt64`
- **Histogram:** Bucketed value distribution with mutex protection
- **Timer:** Measures duration via `defer timer.Time()()`

### Validation Pattern

**Location:** `validation/validation.go`

```go
// Always sanitize + validate before use
func ValidateQuery(query string) (string, error) {
    // 1. Empty check
    if strings.TrimSpace(query) == "" {
        return "", errors.NewQueryEmptyError()
    }
    
    // 2. Length check
    if len(query) > constants.MaxQueryLength {
        return "", errors.NewQueryTooLongError(...)
    }
    
    // 3. Remove control characters
    cleaned := strings.Map(sanitizeRune, query)
    
    // 4. Check for dangerous characters
    if dangerousChars.MatchString(cleaned) {
        return "", errors.NewQueryInvalidCharsError(...)
    }
    
    // 5. Normalize whitespace
    cleaned = strings.Join(strings.Fields(cleaned), " ")
    
    return cleaned, nil
}
```

---

## 9. Platform-Specific Behavior

### Runtime Platform Detection

```go
func getCurrentPlatform() string {
    switch runtime.GOOS {
    case "windows": return constants.PlatformWindows
    case "darwin":  return constants.PlatformMacOS
    default:        return constants.PlatformLinux
    }
}
```

### Platform Filtering in Search

- Commands with `platform: [linux]` only shown on Linux (unless `--all-platforms`)
- `cross-platform` always shown
- Platform variants mapped: `darwin` → `macos`, `unix/bash/zsh` → `linux`

### Alias System (Cross-Platform)

**Windows:** Creates `.bat` files in alias directory
```batch
@echo off
"C:\path\to\wtf.exe" %*
```

**Unix:** Creates shell scripts
```bash
#!/bin/bash
exec "/usr/local/bin/wtf" "$@"
```

---

## 10. Performance Characteristics

| Operation | Complexity | Typical Time |
|-----------|------------|--------------|
| Database load | O(n) | ~50ms for 3850 commands |
| Index build | O(n × avg_tokens) | ~20ms |
| BM25F search | O(query_tokens × avg_postings) | ~5-10ms |
| TF-IDF rerank | O(top_k × vocab) | ~2-5ms |
| Total search | End-to-end | ~200ms |

**Optimizations in place:**
1. Lowercased cache fields (avoid repeated `ToLower()`)
2. Inverted index (avoid full scan)
3. Top-K limiting before TF-IDF
4. Term selection by IDF (reduce noise)
5. LRU cache for repeated queries

---

## 11. Extension Points

### Adding a New CLI Command

1. Create `internal/cli/mycommand.go`:
   ```go
   var myCmd = &cobra.Command{
       Use:   "mycommand",
       Short: "Description",
       Run: func(cmd *cobra.Command, args []string) {
           // implementation
       },
   }
   
   func init() {
       myCmd.Flags().StringP("flag", "f", "", "Flag description")
   }
   ```

2. Register in `internal/cli/root.go`:
   ```go
   func init() {
       rootCmd.AddCommand(myCmd)
   }
   ```

### Adding a New Wizard

1. Add to `internal/cli/wizard.go`:
   ```go
   case "mytool":
       runMyToolWizard()
   ```

2. Implement `runMyToolWizard()` using `readInput()`, `readChoice()`, `readYesNo()` helpers.

### Adding Scoring Weight

1. Add constant to `internal/constants/constants.go`
2. Reference in `internal/database/search_universal.go` or `cascading_boost.go`

---

*End of Knowledge Base*

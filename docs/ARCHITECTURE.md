# Project Structure Documentation

## 📁 Directory Organization

```
WTF/
├── cmd/                    # Application entry points
│   └── wtf/               # Main WTF CLI application
│       └── main.go        # Application entry point
├── internal/              # Private application code
│   ├── cli/              # Command-line interface
│   │   ├── alias.go      # Alias management commands
│   │   ├── pipeline.go   # Pipeline search commands
│   │   ├── root.go       # Root command and CLI setup
│   │   ├── root_test.go  # CLI tests
│   │   ├── save.go       # Save command functionality
│   │   ├── search.go     # Search command implementation
│   │   ├── setup.go      # Setup and configuration commands
│   │   └── wizard.go     # Interactive command wizards
│   ├── config/           # Configuration management
│   │   ├── config.go     # Configuration structures and loading
│   │   └── config_test.go # Configuration tests
│   ├── context/          # Context-aware analysis
│   │   ├── analyzer.go   # Directory and project analysis
│   │   └── analyzer_test.go # Context analysis tests
│   ├── database/         # Database operations and search
│   │   ├── cascading_boost.go # Token-based weighted boost scoring
│   │   ├── embedding_loader.go # Word vector embedding loader (optional)
│   │   ├── loader.go     # Database loading and management
│   │   ├── loader_test.go # Database loading tests
│   │   ├── models.go     # Data structures and models
│   │   ├── pipeline_test.go # Pipeline functionality tests
│   │   ├── search.go     # Advanced search algorithms and scoring
│   │   ├── search_universal.go # BM25F universal search with NLP
│   │   └── search_test.go # Search functionality tests
│   ├── errors/           # Error handling utilities
│   │   └── errors.go     # Custom error types
│   ├── history/          # Search history and analytics
│   │   └── history.go    # Search tracking and statistics
│   ├── embedding/        # Word vector embeddings (optional)
│   │   ├── embedding.go  # Embedding loader and cosine similarity
│   │   └── embedding_test.go # Embedding tests
│   ├── nlp/              # Natural Language Processing
│   │   ├── processor.go  # Intent detection and query processing
│   │   └── tfidf.go      # TF-IDF based search and reranking
│   ├── recovery/         # Error recovery mechanisms
│   │   ├── recovery.go   # Database and search recovery
│   │   └── recovery_test.go # Recovery mechanism tests
│   ├── search/           # Enhanced search utilities
│   │   ├── enhanced_search.go # Advanced search with platform filtering
│   │   ├── pattern_learner.go # Dynamic pattern learning
│   │   └── semantic_search.go # Semantic similarity search
│   └── version/          # Version information
│       ├── version.go    # Version constants and build info
│       └── version_test.go # Version tests
├── pkg/                   # Public library code (for future use)
├── assets/               # Static assets and data
│   └── commands.yml      # Main command database (3,845+ commands from TLDR & Cheatsheets)
├── docs/                 # Project documentation
│   ├── ARCHITECTURE.md   # Project Architecture
│   └── ALIASES.md        # Alias setup documentation
├── configs/              # Configuration files
├── build/                # Build artifacts (generated)
├── scripts/              # Build and utility scripts
│   ├── fetch_cheatsheets.go # Database update script
│   ├── prepare_glove.py  # GloVe word vector preparation (optional)
│   └── embed_commands.py # Command embedding generation (optional)
├── .editorconfig         # Editor configuration
├── .gitignore           # Git ignore rules
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build automation (Linux/macOS)
├── build.bat            # Build automation (Windows)
├── README.md            # Main project documentation
└── RELEASE_NOTES.md     # Release documentation
```

## 🏗️ Architecture Overview

### Package Organization

#### `/cmd/wtf`
The main application entry point following Go project layout standards. Contains only the `main.go` file that initializes and runs the CLI application.

#### `/internal`
Private application code that cannot be imported by other projects. Organized by functional domains:

- **`cli/`**: All command-line interface logic, using Cobra framework
- **`config/`**: Configuration management and file handling  
- **`context/`**: Project type detection and context-aware suggestions (15+ project types)
- **`database/`**: Command database operations, BM25F search, cascading boost scoring
- **`embedding/`**: Optional word vector embeddings for semantic search
- **`errors/`**: Custom error types and error handling utilities
- **`history/`**: Search history tracking, analytics, and usage statistics
- **`nlp/`**: Natural language processing for intent detection, TF-IDF reranking
- **`recovery/`**: Database and search error recovery mechanisms
- **`search/`**: Fuzzy search utilities and typo tolerance algorithms
- **`version/`**: Version information and build metadata

#### `/pkg` (Future Use)
Reserved for public library code that could be imported by other projects. Currently empty but available for future extensibility.

#### `/assets`
Static files and data:
- `commands.yml`: Main curated command database (3,845+ commands from TLDR Pages and Cheat/Cheatsheets)

#### `/docs`
Project documentation:
- Development docs (design, requirements, tasks)
- User documentation and guides
- Architecture and alias setup instructions

### Key Design Principles

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Dependency Direction**: Dependencies flow inward (CLI → NLP/context/database)
3. **Testability**: All packages have comprehensive test coverage
4. **Modularity**: Loosely coupled components with clean interfaces
5. **Extensibility**: Easy to add new commands, data sources, and NLP features
6. **Performance**: Optimized algorithms for sub-50ms search response times

### Import Path Structure

```go
// Public module path
github.com/Vedant9500/WTF

// Internal packages  
github.com/Vedant9500/WTF/internal/cli
github.com/Vedant9500/WTF/internal/config
github.com/Vedant9500/WTF/internal/context
github.com/Vedant9500/WTF/internal/database
github.com/Vedant9500/WTF/internal/errors
github.com/Vedant9500/WTF/internal/history
github.com/Vedant9500/WTF/internal/nlp
github.com/Vedant9500/WTF/internal/search
github.com/Vedant9500/WTF/internal/version
```

### Build System

- **Makefile**: Cross-platform builds for Linux/macOS users
- **build.bat**: Windows-native build script
- **go.mod**: Go module with minimal dependencies (Cobra + YAML)

### Configuration

- **`.editorconfig`**: Consistent code formatting across editors
- **`.gitignore`**: Appropriate exclusions for Go projects
- **`assets/commands.yml`**: Main database with 3,845+ commands from TLDR Pages and Cheat/Cheatsheets

### Core Technologies

- **BM25F Search**: Field-weighted Best Match 25 scoring algorithm
- **Cascading Boost**: Token-based weighted boosting (action 3x, context 2.5x, target 2x)
- **NLP Processing**: Intent detection, synonym expansion, and query understanding
- **TF-IDF Reranking**: Term frequency-inverse document frequency refinement
- **Fuzzy Search**: Levenshtein distance-based typo tolerance
- **Platform Filtering**: Filter by OS with --platform and --all-platforms flags
- **Context Detection**: 15+ project types with smart command prioritization
- **Search Analytics**: JSON-based history tracking and usage statistics

This structure follows Go community best practices and makes the project easy to understand, maintain, and extend with advanced search capabilities.

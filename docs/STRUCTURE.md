# Project Structure

This document describes the organization of the WTF codebase following Go best practices and conventions.

## 📁 Directory Structure

```
WTF/
├── cmd/                    # Main applications
│   └── wtf/               # WTF CLI application
│       └── main.go        # Application entry point
├── internal/              # Private application code
│   ├── cli/              # Command-line interface
│   ├── config/           # Configuration management
│   ├── context/          # Context-aware suggestions
│   ├── database/         # Database operations
│   ├── errors/           # Error handling
│   └── version/          # Version information
├── pkg/                   # Public packages (empty for now)
├── assets/               # Static assets
│   └── commands.yml      # Command database
├── configs/              # Configuration files
├── docs/                 # Documentation
│   ├── design.md        # Design documentation
│   ├── requirements.md  # Requirements specification
│   ├── tasks.md         # Development tasks
│   └── ALIASES.md       # Aliases documentation
├── build/               # Build artifacts (generated)
├── scripts/             # Build and utility scripts
├── .editorconfig        # Editor configuration
├── .air.toml           # Live reload configuration
├── .gitignore          # Git ignore rules
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation (Unix)
├── build.bat           # Build automation (Windows)
├── README.md           # Project documentation
├── CONTRIBUTING.md     # Contribution guidelines
├── RELEASE_NOTES.md    # Release documentation
└── LICENSE             # License file
```

## 📝 Package Organization

### `/cmd/wtf/`
- **Purpose**: Main application entry point
- **Contents**: `main.go` with application bootstrapping
- **Dependencies**: Imports from `internal/` packages only

### `/internal/`
Private application packages following the [internal package pattern](https://golang.org/doc/go1.4#internalpackages).

#### `/internal/cli/`
- **Purpose**: Command-line interface implementation
- **Key Files**:
  - `root.go` - Root command and CLI setup
  - `search.go` - Search command implementation
  - `save.go` - Save command implementation
  - `wizard.go` - Interactive command wizards
  - `pipeline.go` - Pipeline search functionality
  - `alias.go` - Alias management
  - `setup.go` - Setup command

#### `/internal/database/`
- **Purpose**: Database operations and search functionality
- **Key Files**:
  - `loader.go` - Database loading and parsing
  - `search.go` - Search algorithms and scoring
  - `models.go` - Data structures
  - `pipeline_test.go` - Pipeline-specific tests

#### `/internal/config/`
- **Purpose**: Configuration management
- **Key Files**:
  - `config.go` - Configuration structure and loading
  - `config_test.go` - Configuration tests

#### `/internal/context/`
- **Purpose**: Context-aware suggestions
- **Key Files**:
  - `analyzer.go` - Directory analysis
  - `context.go` - Context detection and scoring

#### `/internal/version/`
- **Purpose**: Version information management
- **Key Files**:
  - `version.go` - Version constants and build info

### `/pkg/`
- **Purpose**: Public packages that could be imported by external projects
- **Status**: Currently empty, reserved for future public APIs

### `/assets/`
- **Purpose**: Static assets and data files
- **Contents**: `commands.yml` - Main command database

### `/docs/`
- **Purpose**: Project documentation
- **Contents**: Design docs, requirements, development notes

## 🔧 Build and Development

### Build Commands
```bash
# Build for current platform
make build                    # Unix/Linux/macOS
build.bat build              # Windows

# Cross-platform builds
make build-all               # All platforms
build.bat build-all         # All platforms

# Testing
make test                    # Run tests
build.bat test              # Run tests
```

### Development Workflow
1. **Live Reload**: Use `air` for development with automatic rebuilds
2. **Testing**: Run tests in specific packages or all packages
3. **Linting**: Follow Go conventions and use `gofmt`
4. **Building**: Use build scripts for consistent builds

## 📦 Import Path Convention

All internal imports use the full module path:
```go
import "github.com/Vedant9500/WTF/internal/cli"
import "github.com/Vedant9500/WTF/internal/database"
```

## 🔍 Key Design Principles

1. **Separation of Concerns**: Each package has a single responsibility
2. **Dependency Direction**: CLI depends on business logic, not vice versa
3. **Internal Packages**: Use Go's internal package pattern for private code
4. **Standard Layout**: Follow [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
5. **Testability**: Each package is independently testable

## 📋 Package Dependencies

```
cmd/wtf/main.go
    └── internal/cli
            ├── internal/config
            ├── internal/context
            ├── internal/database
            └── internal/version

internal/database
    └── internal/errors

internal/config
    └── (no internal dependencies)

internal/context  
    └── (no internal dependencies)

internal/version
    └── (no internal dependencies)

internal/errors
    └── (no internal dependencies)
```

This structure ensures clean dependency management and maintainable code organization.

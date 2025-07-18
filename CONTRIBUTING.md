# Contributing to WTF (What's The Function)

Thank you for your interest in contributing to WTF! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Code Guidelines](#code-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## 🚀 Getting Started

### Prerequisites

- **Go 1.24+**: [Download and install Go](https://golang.org/dl/)
- **Git**: Version control system
- **Make** (Linux/macOS) or **build.bat** (Windows): Build automation

### Quick Start

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/your-username/WTF.git
   cd WTF
   ```
3. **Install dependencies**:
   ```bash
   go mod tidy
   ```
4. **Build and test**:
   ```bash
   # Linux/macOS
   make test
   make build
   
   # Windows
   build.bat test
   build.bat build
   ```

## 🛠️ Development Setup

### Project Structure

```
WTF/
├── internal/
│   ├── cli/           # Command-line interface
│   ├── database/      # Database operations and search
│   ├── config/        # Configuration management
│   ├── context/       # Context-aware suggestions
│   ├── version/       # Version information
│   └── errors/        # Error handling
├── scripts/           # Build and utility scripts
├── build/            # Build artifacts (generated)
├── commands.yml      # Main command database
├── main.go           # Application entry point
├── Makefile          # Build automation (Linux/macOS)
├── build.bat         # Build automation (Windows)
└── README.md         # Documentation
```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the [Code Guidelines](#code-guidelines)

3. **Test thoroughly**:
   ```bash
   # Run all tests
   go test ./...
   
   # Run with coverage
   go test ./... -cover
   
   # Test specific package
   go test ./internal/database -v
   ```

4. **Build and verify**:
   ```bash
   # Build for current platform
   go build -o wtf .
   
   # Test the binary
   ./wtf "test query"
   ```

5. **Commit and push**:
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** on GitHub

## 🤝 How to Contribute

### 🐛 Bug Reports

When reporting bugs, please include:

- **OS and version** (Windows 10, macOS 13, Ubuntu 22.04, etc.)
- **Go version** (`go version`)
- **WTF version** (`wtf --version`)
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Error messages** (if any)

### 💡 Feature Requests

For new features:

- **Check existing issues** to avoid duplicates
- **Describe the problem** your feature would solve
- **Provide examples** of how it would work
- **Consider backwards compatibility**

### 🔧 Code Contributions

We welcome contributions in these areas:

#### 🗄️ Database Enhancements
- Add new commands and categories
- Improve command descriptions and keywords
- Add missing popular tools

#### 🎨 User Experience
- Improve output formatting
- Add new interactive wizards
- Enhance error messages

#### 🚀 Performance
- Optimize search algorithms
- Reduce memory usage
- Improve startup time

#### 🌐 Platform Support
- Windows-specific improvements
- macOS optimizations
- Linux distribution packaging

#### 📚 Documentation
- README improvements
- Code documentation
- Usage examples

## 📝 Code Guidelines

### Go Code Style

- Follow **standard Go conventions**
- Use **gofmt** for formatting
- Use **golint** for style checking
- Write **clear, descriptive variable names**
- Add **comments for exported functions**

### Example Code Style

```go
// SearchCommands performs a fuzzy search across the command database
// and returns ranked results based on relevance scoring.
func SearchCommands(query string, options SearchOptions) ([]SearchResult, error) {
    if query == "" {
        return nil, errors.New("search query cannot be empty")
    }
    
    // Normalize query for consistent searching
    normalizedQuery := strings.ToLower(strings.TrimSpace(query))
    
    // Implementation...
    return results, nil
}
```

### Package Organization

- **internal/cli**: Command-line interface components
- **internal/database**: All database-related operations
- **internal/config**: Configuration and settings
- **internal/context**: Context detection and analysis
- **internal/version**: Version information and build details

### Error Handling

- Use **descriptive error messages**
- Wrap errors with context: `fmt.Errorf("failed to load database: %w", err)`
- Handle errors gracefully in CLI commands
- Provide helpful suggestions when possible

## 🧪 Testing

### Test Requirements

- **All new code** must include tests
- **Maintain or improve** test coverage
- **Test edge cases** and error conditions
- **Use table-driven tests** for multiple scenarios

### Test Types

#### Unit Tests
```go
func TestSearchCommands(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        expected int
        wantErr  bool
    }{
        {"basic search", "git commit", 5, false},
        {"empty query", "", 0, true},
        {"no results", "nonexistentcommand", 0, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results, err := SearchCommands(tt.query, DefaultOptions)
            
            if tt.wantErr && err == nil {
                t.Error("expected error but got none")
            }
            
            if len(results) != tt.expected {
                t.Errorf("expected %d results, got %d", tt.expected, len(results))
            }
        })
    }
}
```

#### Integration Tests
Test CLI commands end-to-end with real database files.

#### Benchmark Tests
```go
func BenchmarkSearchCommands(b *testing.B) {
    for i := 0; i < b.N; i++ {
        SearchCommands("git commit", DefaultOptions)
    }
}
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/database

# With coverage
go test ./... -cover

# Verbose output
go test ./... -v

# Benchmarks
go test ./... -bench=.
```

## 📚 Documentation

### Code Documentation

- **Public functions** must have doc comments
- **Packages** should have package-level documentation
- **Complex algorithms** need inline comments
- **Examples** in doc comments when helpful

### User Documentation

- **README.md**: Keep up-to-date with new features
- **Command help**: Update CLI help text for new commands
- **Examples**: Add real-world usage examples

## 🚀 Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

### Release Checklist

1. **Update version** in `internal/version/version.go`
2. **Update CHANGELOG.md** with new features and fixes
3. **Run full test suite**: `go test ./...`
4. **Build for all platforms**: `make build-all`
5. **Test release binaries** on different platforms
6. **Create git tag**: `git tag v1.2.3`
7. **Push tag**: `git push origin v1.2.3`
8. **Create GitHub release** with release notes

## 🎯 Areas for Contribution

### High Priority
- 🗄️ **Database expansion**: Add more commands
- 🐛 **Bug fixes**: Fix reported issues
- 📚 **Documentation**: Improve guides and examples
- 🧪 **Test coverage**: Add missing tests

### Medium Priority
- 🎨 **UI/UX improvements**: Better output formatting
- 🚀 **Performance**: Optimize search algorithms
- 🌐 **Platform support**: Package manager integration

### Future Features
- 🔌 **Plugin system**: Extensible architecture
- 🌍 **Internationalization**: Multi-language support
- 📱 **Additional interfaces**: Web UI, mobile apps

## 💬 Communication

- **GitHub Issues**: Bug reports and feature requests
- **Pull Requests**: Code contributions and reviews
- **Discussions**: General questions and ideas

## 📄 License

By contributing to WTF, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

**Thank you for contributing to WTF! 🎉**

Your contributions help make command discovery easier for developers everywhere.

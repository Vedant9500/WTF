# Changelog

All notable changes to WTF (What's The Function) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-07-19

### 🚀 Major Features Added

#### Advanced Natural Language Processing
- **Intent Detection**: Automatically detects user intent (create, compress, search, download, install, etc.)
- **Query Processing**: Sophisticated parsing with action/target recognition
- **Synonym Expansion**: Understands alternatives like "folder" → "directory", "get" → "download"
- **Stop Word Removal**: Filters out noise words for cleaner matching

#### Enhanced Search Capabilities
- **Domain-Specific Matching**: Maps queries to relevant command domains
- **Fuzzy Search**: Built-in typo tolerance using Levenshtein distance
- **Hybrid Search Algorithm**: Combines exact, fuzzy, and semantic matching
- **Advanced Scoring**: Multi-factor relevance calculation with domain-specific boosts

#### Search History & Analytics
- **Usage Tracking**: Persistent search history with JSON storage
- **Analytics Dashboard**: Search statistics and success rates
- **Personalized Results**: Frequently used commands get priority boosts
- **Auto-Cleanup**: Smart history management

#### Improved Context Detection
- **15+ Project Types**: Expanded from basic Git/Docker to comprehensive project detection
- **Smart Boosts**: Project-aware command prioritization
- **Multi-Context Support**: Handles projects with multiple technologies

### ✨ Improvements

#### Search Quality
- **Better Relevance Scoring**: Fixed issues where wrong commands appeared at top of results
- **Category Relevance Boosting**: Commands get extra points for category-specific matches
- **Penalty System**: Reduces scores for mismatched tools (e.g., find commands for compression queries)
- **Exact Match Prioritization**: Heavily weights exact command name matches

#### Performance Optimizations
- **Sub-30ms Search**: Improved from ~50ms to often under 30ms response times
- **Efficient NLP**: Lightweight processing without external dependencies
- **Memory Management**: Smart caching and cleanup for long-running sessions
- **Real-time Fuzzy Search**: Typo correction with minimal performance impact

#### User Experience
- **Verbose Mode**: Shows NLP analysis and scoring details for debugging
- **Enhanced Output**: Relevance scores and better formatted results
- **Command Examples**: More comprehensive examples in search results

### 🐛 Bug Fixes
- **Search Relevance**: Fixed "compress files" returning find commands instead of tar/zip
- **Directory Creation**: Fixed "create directory" returning makepkg instead of mkdir
- **Keyword Expansion**: Limited over-aggressive synonym expansion that diluted results
- **Scoring Algorithm**: Improved prioritization of most relevant commands

### 📚 Documentation Updates
- **README.md**: Comprehensive rewrite with new features and examples
- **ARCHITECTURE.md**: Updated to reflect new NLP and search components
- **Performance Metrics**: Updated benchmarks and specifications
- **Feature Documentation**: Detailed explanations of NLP and search capabilities

### 🏗️ Technical Changes
- **New Packages**: Added `internal/nlp/`, `internal/search/`, `internal/history/`
- **Enhanced Database Schema**: Better command categorization and keyword support
- **Improved Test Coverage**: Additional tests for new functionality
- **Code Organization**: Better separation of concerns and modularity

---

## [1.0.0] - 2025-07-15

### 🎉 Initial Release

#### Core Features
- **Natural Language Search**: Basic command discovery using plain English
- **Command Database**: 3,845+ commands sourced from TLDR pages
- **Context Awareness**: Basic Git and project type detection
- **Personal Notebook**: Save and organize custom commands
- **Interactive Wizards**: Step-by-step command builders
- **Pipeline Search**: Multi-command workflow discovery
- **Cross-Platform**: Support for Windows, macOS, and Linux
- **Alias System**: Custom command name setup

#### Search Capabilities
- **Basic Text Matching**: Command and description search
- **Category Filtering**: Commands organized by categories
- **Keyword Matching**: Tag-based command discovery

#### User Interface
- **CLI Framework**: Built with Cobra for consistent command structure
- **Formatted Output**: Clean, readable command suggestions
- **Help System**: Comprehensive help and usage information

#### Technical Foundation
- **Go Implementation**: Modern, fast, single-binary deployment
- **YAML Database**: Human-readable command storage
- **Minimal Dependencies**: Lightweight with few external requirements
- **Build System**: Cross-platform build scripts and automation

---

## Roadmap

### 🔮 Planned Features (v1.2.0)
- **Machine Learning Integration**: Command recommendation based on usage patterns
- **Plugin System**: Support for custom command sources and processors
- **Web Interface**: Browser-based command discovery
- **Team Sharing**: Shared command databases for teams
- **Shell Integration**: Deeper integration with bash, zsh, PowerShell

### 🎯 Long-term Goals
- **Multi-language Support**: Internationalization for global users
- **Cloud Sync**: Synchronize personal commands across devices
- **Mobile Apps**: Native mobile applications
- **IDE Plugins**: VS Code, IntelliJ, and other IDE integrations
- **Community Hub**: Central repository for sharing command collections

---

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:
- 🐛 Bug reports and feature requests
- 🔧 Code contributions and pull requests  
- 📚 Documentation improvements
- 🧪 Testing and quality assurance

## Acknowledgments

- **[TLDR Pages](https://github.com/tldr-pages/tldr)** - Primary command database source
- **[Cheat/Cheatsheets](https://github.com/cheat/cheatsheets)** - Additional curated command examples  
- **[Cobra](https://github.com/spf13/cobra)** - Excellent CLI framework
- **Go Community** - Amazing ecosystem and tooling
- **Contributors** - Everyone who helped make WTF better!

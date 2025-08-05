# 🔍 WTF (What's The Function)

<div align="center">

![Version](https://img.shields.io/badge/version-1.2.0-blue)
![Go Version](https://img.shields.io/badge/go-1.24+-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/Vedant9500/WTF)](https://goreportcard.com/report/github.com/Vedant9500/WTF)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

*A powerful CLI tool to discover shell commands using natural language*

**When you can't remember a command, you think "What's The Function I need?" - that's WTF! 😄**

</div>

## ✨ Features

🔍 **Advanced Natural Language Search** - Find commands by describing what you want to do in plain English  
🧠 **Intent Detection** - Understands your intent (create, search, compress, install, etc.) for better results  
📚 **Massive Command Database** - 3,268+ commands sourced from TLDR pages and community contributions  
🎯 **Context-Aware Suggestions** - Smart recommendations based on your current directory and project type  
⚡ **Fuzzy Search & Typo Tolerance** - Finds commands even with spelling mistakes  
�️ ***Platform Filtering** - Filter commands by platform (Linux, macOS, Windows, cross-platform)  
� **Searcha History & Analytics** - Tracks your searches to improve recommendations  
📝 **Personal Command Notebook** - Save and organize your custom commands  
🎯 **Interactive Command Builder** - Step-by-step wizards for complex commands  
🔗 **Pipeline Search** - Specialized search for multi-command workflows  
⚡ **Lightning Fast** - ~200ms search performance with advanced scoring and caching  
�  **Cross-Platform** - Works on Windows, macOS, and Linux  
🎨 **Beautiful Output** - Clean, formatted command suggestions with relevance scores

---

## 🚀 Quick Start

### Installation

#### Option 1: Download Binary (Recommended)
```bash
# Download from releases page (coming soon)
# Extract and run: ./wtf "your query"
```

#### Option 2: Build from Source
```bash
git clone https://github.com/Vedant9500/WTF.git
cd WTF

# On Windows
build.bat build

# On Linux/Mac (with make)
make build

# Alternative (any platform)
go build -o wtf ./cmd/wtf
```

### Basic Usage

```bash
# Search for commands
wtf "compress files"
wtf "find large files"
wtf "git commit changes"

# Set up your preferred command name
wtf setup hey
hey "docker commands"
```

### 📚 Database Source

WTF leverages excellent community-driven projects as its command database sources:

**Primary Sources:**
- **[TLDR Pages](https://github.com/tldr-pages/tldr)** - Simplified, example-focused help pages for command-line tools
- **[Cheat/Cheatsheets](https://github.com/cheat/cheatsheets)** - Community-maintained cheatsheets for various tools and technologies

**Database Features:**
- **3,845+ Commands**: Comprehensive coverage of common CLI tools
- **Community Curated**: Maintained by developers worldwide
- **Example-Focused**: Practical usage patterns, not just syntax
- **Always Updated**: Fresh content from both communities
- **Multi-Source**: Best commands from multiple trusted sources

---

## 📚 Complete Feature Guide

### 🔍 Core Search

```bash
# Natural language search with advanced NLP
wtf "compress files"                 # Finds tar, zip, gzip commands
wtf "create directory"               # Prioritizes mkdir over other tools
wtf "download file"                  # Returns wget, curl commands
wtf "git commands"                   # Git-specific operations

# Advanced search options
wtf search "docker" --limit 10       # More results
wtf search "process" --verbose       # Show relevance scores and NLP analysis
wtf search "commands" --database custom.yml # Custom database

# Platform-specific searches
wtf "list files" --platform linux    # Linux-specific commands + cross-platform
wtf "compress files" --platform windows,macos # Multiple platforms
wtf "process management" --all-platforms # Override filtering, show all
wtf "system tools" --platform linux --no-cross-platform # Linux only

# Fuzzy search handles typos
wtf "comprss files"                  # Still finds compression commands
wtf "mkdir direectory"               # Still finds directory commands
wtf "gti comit changez"              # Finds git commit commands
```

### 🧠 Advanced Natural Language Processing

WTF now includes sophisticated NLP features for better command understanding:

**Intent Detection**:
- `create` → Prioritizes mkdir, touch, make commands
- `compress` → Focuses on tar, zip, gzip tools  
- `search` → Emphasizes grep, find, locate commands
- `download` → Highlights wget, curl, fetch tools
- `install` → Boosts package managers (apt, pip, npm)

**Query Processing**:
- **Action Recognition**: Identifies verbs like "compress", "extract", "download"
- **Target Detection**: Recognizes objects like "file", "directory", "package"
- **Synonym Expansion**: "folder" → "directory", "get" → "download"
- **Stop Word Removal**: Filters out "the", "and", "with" for cleaner matching

**Example NLP in action**:
```bash
wtf "I want to compress some files into an archive"
# Detects: Intent=compress, Action=compress, Target=files
# Returns: tar, zip, gzip commands with high relevance scores

wtf "help me create a new directory please"  
# Detects: Intent=create, Action=create, Target=directory
# Returns: mkdir commands prioritized over other creation tools

```

### 🧠 Context-Aware Search

WTF automatically detects your environment and prioritizes relevant commands:

```bash
# In a Git repository
wtf "commit"          # Prioritizes git commands

# Directory with Dockerfile  
wtf "build"           # Prioritizes docker commands

# Directory with package.json
wtf "install"         # Prioritizes npm commands

# Directory with Makefile
wtf "build"           # Prioritizes make commands

# Python project (requirements.txt, .py files)
wtf "install"         # Prioritizes pip commands
```

**Context Detection Features**:
- **15+ Project Types**: Git, Docker, Node.js, Python, Go, Rust, Java, and more
- **File Pattern Recognition**: Detects project files like package.json, Dockerfile, go.mod
- **Smart Boosts**: Gives relevant commands higher priority scores
- **Multi-Context Support**: Handles projects with multiple technologies

### 🖥️ Platform-Specific Search

WTF now supports filtering commands by platform, perfect for developers working across multiple operating systems:

```bash
# Filter by specific platform
wtf "list files" --platform linux
wtf "compress files" --platform windows
wtf "process management" --platform macos

# Multiple platforms
wtf "text processing" --platform linux,macos
wtf "network tools" --platform windows,linux

# Cross-platform behavior
wtf "git commands" --platform linux              # Linux + cross-platform commands
wtf "docker tools" --platform linux --no-cross-platform  # Linux only
wtf "system tools" --all-platforms               # All platforms (override filtering)

# Verbose output shows platform filtering
wtf "compression" --platform linux --verbose
# Platform filter: [linux] + cross-platform
# Shows which platforms are being searched
```

**Platform Filtering Features**:
- **Supported Platforms**: `linux`, `macos`, `windows`, `cross-platform`
- **Smart Defaults**: Cross-platform commands included by default
- **Multiple Selection**: Comma-separated platform lists
- **Override Options**: `--all-platforms` to disable filtering entirely
- **Exclusion Control**: `--no-cross-platform` to exclude cross-platform commands
- **Performance**: Platform filtering with full caching support

**Use Cases**:
- **Learning**: Discover Linux commands while on Windows
- **Documentation**: Find platform-specific alternatives
- **Migration**: Compare commands across different systems
- **Development**: Work with multi-platform deployment scripts

### 🎯 Interactive Command Wizards

Build complex commands step-by-step with interactive wizards:

```bash
# Interactive tar archive builder
wtf wizard tar
→ What do you want to do? [c]reate/e[x]tract: c
→ Archive name: backup.tar.gz
→ Files to archive: /home/user/documents
→ Result: tar -czf backup.tar.gz /home/user/documents

# Interactive find command builder  
wtf wizard find
→ Starting directory: .
→ File name pattern: *.log
→ Result: find . -name "*.log"

# Interactive ffmpeg converter
wtf wizard ffmpeg
→ Input file: video.mp4
→ Output format: mp3
→ Result: ffmpeg -i video.mp4 output.mp3
```

### 🔗 Pipeline Search

Find and visualize multi-step command workflows:

```bash
wtf pipeline "text processing"
wtf pipeline "log analysis" 
wtf pipeline "find and replace"

# Example output:
📋 Found pipeline command:
find . -name "*.txt" │ xargs grep "error" │ head -10
📝 find text files and show first 10 errors
🔗 Pipeline steps:
   1. find . -name "*.txt"
   2. xargs grep "error" 
   3. head -10
```

### � Search History & Analytics

WTF tracks your search patterns to provide better recommendations:

```bash
# View search history
wtf history

# Recent searches:
# 1. "compress files" → tar -czf (3 times today)
# 2. "git commit" → git commit -m (2 times today)  
# 3. "docker build" → docker build . (1 time today)

# Search analytics
wtf history --stats

# 📊 Search Statistics:
# Total searches: 47
# Most searched: "git commands" (8 times)
# Success rate: 94% (commands found vs not found)
# Average results per search: 4.2
```

**History Benefits**:
- **Personalized Results**: Frequently used commands get slight priority boosts
- **Usage Patterns**: Understand your command habits
- **Success Tracking**: See which searches work best
- **Auto-Cleanup**: Old history automatically managed

### 📝 Personal Command Notebook

Save and organize your custom commands:

```bash
# Save a regular command
wtf save
→ Command: docker ps -a --format "table {{.Names}}\t{{.Status}}"
→ Description: Show docker containers with custom format
→ Keywords: docker, containers, format
→ Saved to personal notebook!

# Save a pipeline workflow
wtf save-pipeline  
→ Command: find . -name "*.log" | grep -v "debug" | tail -20
→ Description: Get recent non-debug log entries
→ Keywords: logs, debug, recent
→ Saved to personal notebook!

# Your commands appear in all searches
wtf "docker containers"  # Shows both official and your custom commands
```

### 🎨 Beautiful Output

WTF provides clean, formatted output with advanced relevance scoring:

```
🔍 Searching for: compress files

📋 Found 5 matching command(s):

1. tar -czf archive.tar.gz folder/
   📝 compress a folder into a tar.gz archive
   📂 Category: compression
   🏷️  Keywords: tar, compress, archive
   📊 Relevance Score: 127.3

2. zip -r archive.zip folder/
   📝 compress folder into a zip file  
   📂 Category: compression
   🏷️  Keywords: zip, compress, archive
   📊 Relevance Score: 98.7

3. gzip file.txt
   📝 compress a single file with gzip
   📂 Category: compression  
   🏷️  Keywords: gzip, compress, file
   📊 Relevance Score: 85.2
```

**Verbose Mode** shows NLP analysis:
```bash
wtf search "compress files" --verbose

🧠 NLP Analysis:
   Intent: compress
   Actions: [compress]  
   Targets: [files]
   Enhanced Keywords: [compress, archive, files, tar, zip]
   
📊 Scoring Details:
   Command Match: +15.0 (exact match bonus)
   Domain Specific: +12.0 (compression domain)
   Intent Boost: ×2.5 (compression intent)
   Category Boost: ×1.5 (compression category)
   Platform Filter: [all platforms]
```

**Platform Filtering in Verbose Mode**:
```bash
wtf "system tools" --platform linux --verbose

🖥️ Platform Analysis:
   Filter: [linux] + cross-platform
   Excluded: windows, macos (platform-specific)
   Included: 1,247 commands (38% of database)
   
📊 Results by Platform:
   Linux-specific: 3 commands
   Cross-platform: 2 commands
```

---

## ⚙️ Setup & Configuration

### Custom Command Names

Set up WTF with any command name you prefer:

```bash
# One-command setup (creates alias/script automatically)
wtf setup hey         # Creates 'hey' command
wtf setup miko        # Creates 'miko' command
wtf setup cmd         # Creates 'cmd' command

# Manual setup options:
# Windows (PowerShell)
Set-Alias hey wtf

# Windows (CMD)  
doskey hey=wtf.exe $*

# Linux/Mac
alias hey='wtf'
echo "alias hey='wtf'" >> ~/.bashrc
```

### Database Configuration

WTF comes with 3,845+ curated commands sourced from TLDR pages and supports custom databases:

```bash
# Use custom database
wtf --database /path/to/custom.yml

# Database locations:
# Default: assets/commands.yml (3,845+ commands from TLDR)
# Personal: ~/.config/wtf/personal.yml (auto-created)
# Custom: any YAML file following the schema

# Database stats
wtf stats
# 📊 Database Statistics:
# Total commands: 3,845
# Categories: 12 (compression, system, networking, etc.)
# Platforms: Linux, macOS, Windows
# Average keywords per command: 4.2
```

---

## 🏗️ Development & Building

### Building from Source

```bash
# Clone repository
git clone https://github.com/your-username/WTF.git
cd WTF

# Install dependencies
go mod tidy

# Build for current platform
go build -o wtf ./cmd/wtf

# Or use build scripts:
# Windows
build.bat build

# Linux/Mac (with make)
make build
```

### Cross-Platform Building

```bash
# Build for all platforms
build.bat build-all     # Windows
make build-all          # Linux/Mac

# Creates binaries for:
# - Linux (amd64, arm64)
# - macOS (amd64, arm64) 
# - Windows (amd64)
```

### Testing

```bash
# Run tests
go test ./...

# With coverage
build.bat test          # Windows  
make test-coverage      # Linux/Mac (generates coverage.html)

# Run benchmarks
make benchmark
```

---

## 📊 Performance

WTF is optimized for speed with advanced algorithms:

- **Search Performance**: ~200ms average response time (optimized for accuracy)
- **NLP Processing**: < 50ms for intent detection and query analysis  
- **Database Size**: 3,268+ commands, ~2.8MB total
- **Memory Usage**: < 20MB RAM (includes enhanced search algorithms)
- **Binary Size**: < 25MB (statically linked with all features)
- **Cold Start**: < 150ms first run
- **Fuzzy Search**: Advanced typo correction with Levenshtein distance
- **Platform Filtering**: Instant filtering with full caching support

**Optimization Features**:
- **Hybrid Search Algorithm**: Combines exact, fuzzy, and semantic matching
- **Smart Scoring**: Multi-factor relevance calculation with domain-specific boosts
- **Efficient NLP**: Lightweight intent detection without external dependencies
- **Memory Management**: Smart caching and cleanup for long-running sessions

---

## 🗄️ Database

### Built-in Database
- **3,268+ curated commands** from [TLDR Pages](https://github.com/tldr-pages/tldr) and community contributions
- **Categories**: compression, system, networking, development, version-control, text-processing, and more
- **Multi-Platform**: Commands for Linux, macOS, Windows, and cross-platform tools
- **Platform Filtering**: Advanced filtering by platform with smart cross-platform handling
- **Enhanced Coverage**: Essential commands like `cal`, `wc`, `uniq`, `tr`, `yq` included
- **Regular updates** with new commands and improvements
- **Community Driven**: Maintained by multiple open-source communities worldwide

### Personal Database
- **Location**: `~/.config/wtf/personal.yml`
- **Auto-created** when you save first command
- **Merged** with main database in search results
- **Full CRUD** operations via CLI
- **Search Integration**: Personal commands appear in all searches with proper scoring

### Advanced Search Features
- **Intent-Aware Scoring**: Commands scored based on detected user intent
- **Domain-Specific Matching**: Special relevance for command categories
- **Fuzzy Matching**: Handles typos and partial matches
- **Context Boosting**: Project-aware command prioritization

### Command Schema
```yaml
commands:
  - command: "docker ps -a"
    description: "list all docker containers"
    keywords: ["docker", "containers", "ps"]
    category: "development"
    pipeline: false
```

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes and add tests
4. **Test** thoroughly: `make test`
5. **Commit** with clear messages: `git commit -m "Add amazing feature"`
6. **Push** to your branch: `git push origin feature/amazing-feature`
7. **Open** a Pull Request

### Areas for Contribution
- 🗄️ **Database**: Add more commands and categories
- 🌐 **Localization**: Support for multiple languages  
- 🎨 **Themes**: Custom color schemes and output formats
- 🔌 **Integrations**: IDE plugins, shell integrations
- 📱 **Platforms**: Mobile apps, web interface

---


## 🙏 Acknowledgments

- **[TLDR Pages](https://github.com/tldr-pages/tldr)** - Primary command database source (3,845+ commands)
- **[Cheat/Cheatsheets](https://github.com/cheat/cheatsheets)** - Additional curated command examples and usage patterns
- **[Cobra](https://github.com/spf13/cobra)** - Excellent CLI framework
- **Go Community** - Amazing ecosystem and tools

---

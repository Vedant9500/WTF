# 🔍 WTF (What's The Function)

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go Version](https://img.shields.io/badge/go-1.24+-green)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

*A powerful CLI tool to discover shell commands using natural language*

**When you can't remember a command, you think "What's The Function I need?" - that's WTF! 😄**

</div>

## ✨ Features

🔍 **Natural Language Search** - Find commands by describing what you want to do  
🧠 **Context-Aware Suggestions** - Smart recommendations based on your current directory  
📝 **Personal Command Notebook** - Save and organize your custom commands  
🎯 **Interactive Command Builder** - Step-by-step wizards for complex commands  
🔗 **Pipeline Search** - Specialized search for multi-command workflows  
⚡ **Lightning Fast** - Sub-50ms search performance  
🌍 **Cross-Platform** - Works on Windows, macOS, and Linux  
🎨 **Beautiful Output** - Clean, formatted command suggestions with examples

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

---

## 📚 Complete Feature Guide

### 🔍 Core Search

```bash
# Basic search (default behavior)
wtf "list files"
wtf search "process monitoring" 

# Advanced options
wtf "docker" --limit 10              # More results
wtf "git" --verbose                  # Show scoring details
wtf "commands" --database custom.yml # Custom database
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
```

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

WTF provides clean, formatted output that's easy to scan:

```
🔍 Searching for: compress files

📋 Found 5 matching command(s):

1. tar -czf archive.tar.gz folder/
   📝 compress a folder into a tar.gz archive
   📂 Category: filesystem
   🏷️  Keywords: tar, compress, archive

2. zip -r archive.zip folder/
   📝 compress folder into a zip file
   📂 Category: filesystem
   🏷️  Keywords: zip, compress, archive
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

WTF comes with 1,200+ curated commands and supports custom databases:

```bash
# Use custom database
wtf --database /path/to/custom.yml

# Database locations:
# Default: embedded in binary
# Personal: ~/.config/cmd-finder/personal.yml (auto-created)
# Custom: any YAML file following the schema
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

WTF is optimized for speed:

- **Search Performance**: < 50ms average response time
- **Database Size**: 1,200+ commands, ~1MB total
- **Memory Usage**: < 10MB RAM
- **Binary Size**: < 15MB (statically linked)
- **Cold Start**: < 100ms first run

---

## 🗄️ Database

### Built-in Database
- **1,200+ curated commands** from [cheat/cheatsheets](https://github.com/cheat/cheatsheets)
- **Categories**: filesystem, version-control, development, system, networking
- **Regular updates** with new commands and improvements

### Personal Database
- **Location**: `~/.config/cmd-finder/personal.yml`
- **Auto-created** when you save first command
- **Merged** with main database in search results
- **Full CRUD** operations via CLI

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

- **[cheat/cheatsheets](https://github.com/cheat/cheatsheets)** - Primary command database source
- **[Cobra](https://github.com/spf13/cobra)** - Excellent CLI framework
- **Go Community** - Amazing ecosystem and tools

---

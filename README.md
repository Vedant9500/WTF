# WTF (What's The Function) 

A CLI tool to find shell commands using natural language queries.

## Overview

WTF helps you discover shell commands by searching through a curated database of common command-line tools and their usage examples. Simply describe what you want to do in natural language, and WTF will suggest relevant commands.

**Why "WTF"?** When you can't remember a command, you think "What's The Function I need?" - that's exactly what this tool helps you find! 😄

## Installation

### From Source

```bash
git clone <repository-url>
cd WTF
make build
```

### Direct Build
```bash
go build -o wtf
# or on Windows:
go build -o wtf.exe
```

## Usage

### Basic Search

```bash
# Search for commands (multiple ways)
./wtf "compress a directory"
./wtf search "find files by name"
./wtf "git commit changes"
```

### Advanced Options

```bash
# Limit number of results
./wtf "docker commands" --limit 3

# Verbose output with keywords and scores
./wtf "tar compress" --verbose

# Use custom database file
./wtf "git" --database /path/to/custom.yml

# Get help
./wtf --help
./wtf search --help

# Check version
./wtf --version
```

### Setting up Custom Command Aliases (FR2)

WTF makes it super easy to use any command name you prefer:

#### 🚀 **One-Command Setup (All Platforms)**

```bash
# Simple setup - WTF handles everything automatically
wtf setup hey        # Creates 'hey' command
wtf setup miko       # Creates 'miko' command  
wtf setup cmd        # Creates 'cmd' command

# Then use your custom command:
hey "compress files"
miko "git commands"
```

#### 🪟 **Windows - Instant Setup**

```cmd
# For current session (super simple):
doskey hey=wtf.exe $*

# Now use immediately:
hey "find large files"
```

#### 🐧 **Linux/Mac - Classic Aliases**

```bash
# Quick alias:
alias hey='wtf'

# Make permanent:
echo "alias hey='wtf'" >> ~/.bashrc
```

## Database

The tool uses a curated database of 1,200+ commands sourced from the [cheat/cheatsheets](https://github.com/cheat/cheatsheets) repository, covering:

- Version control (git, svn)
- File operations (tar, zip, find)
- Text processing (grep, awk, sed)
- System administration (systemctl, ssh)
- Development tools (npm, pip, docker)
- And much more...

## Development Status

**Phase 1 (MVP) - ✅ COMPLETED**

- ✅ Command database with 1,218+ entries
- ✅ Robust CLI structure with Cobra framework
- ✅ Advanced search functionality with relevance scoring
- ✅ Configuration system with multiple database support
- ✅ Comprehensive error handling
- ✅ Command-line flags (--verbose, --limit, --database)
- ✅ Version management
- ✅ Test coverage for core components
- ✅ Cross-platform build system

**Coming Soon (Phase 2):**
- 🚧 Context-aware suggestions (git repos, dockerfiles)
- 🚧 Personal command notebook (`save` command)
- 🚧 Interactive command builder
- 🚧 Fuzzy search for typos

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

[Add your license here]
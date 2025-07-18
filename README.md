# cmd-finder

A CLI tool to find shell commands using natural language queries.

## Overview

cmd-finder helps you discover shell commands by searching through a curated database of common command-line tools and their usage examples. Simply describe what you want to do in natural language, and cmd-finder will suggest relevant commands.

## Installation

### From Source

```bash
git clone <repository-url>
cd cmd-finder
make build
```

### Direct Build
```bash
go build -o cmd-finder
```

## Usage

### Basic Search

```bash
# Search for commands (multiple ways)
./cmd-finder "compress a directory"
./cmd-finder search "find files by name"
./cmd-finder "git commit changes"
```

### Advanced Options

```bash
# Limit number of results
./cmd-finder "docker commands" --limit 3

# Verbose output with keywords and scores
./cmd-finder "tar compress" --verbose

# Use custom database file
./cmd-finder "git" --database /path/to/custom.yml

# Get help
./cmd-finder --help
./cmd-finder search --help

# Check version
./cmd-finder --version
```

### Setting up an Alias

To use a custom command like `hey` instead of `cmd-finder`, add this to your shell configuration:

#### Bash/Zsh (~/.bashrc or ~/.zshrc)
```bash
alias hey='/path/to/cmd-finder'
```

#### Fish (~/.config/fish/config.fish)
```fish
alias hey='/path/to/cmd-finder'
```

Then you can use:
```bash
hey search "compress files"
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
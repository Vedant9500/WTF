# cmd-finder

A CLI tool to find shell commands using natural language queries.

## Overview

cmd-finder helps you discover shell commands by searching through a curated database of common command-line tools and their usage examples. Simply describe what you want to do in natural language, and cmd-finder will suggest relevant commands.

## Installation

### From Source

```bash
git clone <repository-url>
cd cmd-finder
go build -o cmd-finder
```

## Usage

### Basic Search

```bash
# Search for commands
./cmd-finder search "compress a directory"
./cmd-finder search "find files by name"
./cmd-finder search "git commit changes"
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

This project is currently in development. Implemented features:

- ✅ Command database with 1,200+ entries
- ✅ Basic CLI structure
- 🚧 Search functionality (coming soon)
- 🚧 Context-aware suggestions
- 🚧 Interactive command builder
- 🚧 Personal command notebook

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

[Add your license here]
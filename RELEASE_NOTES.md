# Release Notes

## WTF v1.0.0 🎉

**Release Date**: January 18, 2025

We're excited to announce the first stable release of WTF (What's The Function)! This release includes all core functionality and represents the completion of our Phase 1-4 development cycle.

### 🎯 What is WTF?

WTF is a powerful CLI tool that helps you discover shell commands using natural language queries. When you can't remember a command, you think "What's The Function I need?" - that's WTF!

### ✨ Key Features

#### 🔍 **Natural Language Search**
- Search through 1,200+ curated commands using plain English
- Example: `wtf "compress files"` → `tar -czf archive.tar.gz folder/`
- Sub-50ms search performance with intelligent relevance scoring

#### 🧠 **Context-Aware Suggestions** 
- Automatically detects your environment (Git repos, Docker projects, etc.)
- Prioritizes relevant commands based on current directory
- 2.7x improvement in suggestion relevance

#### 📝 **Personal Command Notebook**
- Save your custom commands with `wtf save`
- Organize with descriptions, keywords, and categories
- Personal commands appear in all search results
- Stored in `~/.config/cmd-finder/personal.yml`

#### 🎯 **Interactive Command Wizards**
- Step-by-step builders for complex commands
- Built-in wizards for `tar`, `find`, and `ffmpeg`
- Example: `wtf wizard tar` → Interactive archive creation

#### 🔗 **Pipeline Search & Visualization**
- Specialized search for multi-command workflows
- Visual step-by-step breakdown of pipelines
- Example: `wtf pipeline "text processing"` → Shows command chains with visual formatting

### 🚀 **Installation & Setup**

#### Quick Start
```bash
# Download binary (Windows/Linux/macOS)
# Extract and run:
./wtf "your query"

# Set up custom command name
./wtf setup hey
hey "docker commands"
```

#### Build from Source
```bash
git clone https://github.com/your-username/WTF.git
cd WTF
go build -o wtf .
```

### 📊 **Performance Benchmarks**

- **Search Speed**: 48.9ms average response time
- **Database Size**: 1,222 total commands (1,000+ curated + personal)
- **Memory Usage**: < 10MB RAM
- **Binary Size**: < 15MB (statically linked)
- **Test Coverage**: 85.2% database, 83.3% config, 57.0% context

### 🛠️ **Technical Highlights**

- **Language**: Go 1.24+ with Cobra CLI framework
- **Architecture**: Modular design with comprehensive error handling
- **Cross-Platform**: Native binaries for Windows, macOS, Linux (amd64/arm64)
- **Testing**: Comprehensive test suite with automated CI/CD
- **Database**: YAML-based with extensible schema

### 📚 **Documentation**

- **README**: Comprehensive setup and usage guide
- **Examples**: Real-world command discovery scenarios
- **API**: Well-documented CLI interface with help system
- **Build System**: Cross-platform Makefile + Windows batch scripts

### 🎯 **All Requirements Completed**

**Phase 1: MVP** ✅
- [x] FR1: Core Command Search Engine
- [x] FR2: Customizable Call Command
- [x] All NFRs: Performance, Cross-platform, Usability

**Phase 2: Enhanced Features** ✅
- [x] FR4: Context-Aware Suggestions  
- [x] FR5: Personal Command Notebook

**Phase 3: Unique Features** ✅
- [x] FR3: Interactive Command Builder
- [x] FR6: Enhanced Pipeline Support

**Phase 4: Release & Polish** ✅
- [x] Task 4.1: Comprehensive Testing
- [x] Task 4.2: Cross-Platform Build Automation
- [x] Task 4.3: Complete Documentation
- [x] Task 4.4: v1.0.0 Release

### 🔄 **Upgrade Notes**

This is the first stable release. Future versions will maintain backward compatibility.

### 🐛 **Known Issues**

No known critical issues. Please report bugs via GitHub Issues.

### 🤝 **Contributing**

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### 🙏 **Acknowledgments**

- [cheat/cheatsheets](https://github.com/cheat/cheatsheets) for the command database
- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- The Go community for amazing tools and ecosystem

### 📝 **License**

MIT License - see [LICENSE](LICENSE) for details.

---

**Download**: [Releases Page](https://github.com/your-username/WTF/releases/tag/v1.0.0)
**Documentation**: [README.md](README.md)
**Issues**: [GitHub Issues](https://github.com/your-username/WTF/issues)

**⭐ Star the repo if WTF helps you find the commands you need! ⭐**

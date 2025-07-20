# Contributing to WTF

Thank you for your interest in contributing to WTF! We welcome contributions from the community.

## 🚀 How to Contribute

### Reporting Issues
- Search existing issues before creating a new one
- Use clear, descriptive titles
- Include reproduction steps and environment details
- Add relevant labels when possible

### Feature Requests
- Check the [Feature Roadmap](docs/FEATURES.md) first
- Explain the use case and expected benefit
- Consider the impact on existing functionality

### Code Contributions
1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following our guidelines below
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Commit your changes**: `git commit -m 'Add amazing feature'`
7. **Push to the branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

## 📋 Development Guidelines

### Code Style
- Follow standard Go conventions (`go fmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and concise

### Testing
- Write tests for new features and bug fixes
- Ensure all tests pass: `go test ./...`
- Aim for good test coverage of critical paths
- Include both unit tests and integration tests

### Documentation
- Update README.md for user-facing changes
- Add godoc comments for public APIs
- Update relevant documentation in `/docs`
- Include examples where helpful

### Commit Messages
- Use clear, descriptive commit messages
- Start with a verb in present tense
- Reference issues when applicable
- Keep the first line under 72 characters

### Pull Request Guidelines
- Link to relevant issues
- Describe what changes were made and why
- Include screenshots for UI changes
- Ensure CI checks pass
- Be responsive to feedback

## 🏗️ Development Setup

### Prerequisites
- Go 1.21 or later
- Git

### Local Development
```bash
# Clone the repository
git clone https://github.com/Vedant9500/WTF.git
cd WTF

# Install dependencies
go mod download

# Build the project
go build -o wtf ./cmd/wtf

# Run tests
go test ./...

# Run with development flags
./wtf "your test query" --verbose
```

### Project Structure
```
WTF/
├── cmd/wtf/           # Main application entry point
├── internal/          # Private application code
│   ├── cli/           # CLI commands and interface
│   ├── database/      # Command database and search
│   ├── nlp/           # Natural language processing
│   ├── context/       # Project context detection
│   ├── history/       # Search history management
│   └── ...
├── assets/            # Static assets (command database)
├── docs/              # Documentation
└── scripts/           # Build and development scripts
```

## 🤝 Community

- Be respectful and inclusive
- Follow our [Code of Conduct](CODE_OF_CONDUCT.md)
- Help others learn and grow
- Share knowledge and best practices

## 📞 Getting Help

- Check existing documentation
- Search through issues
- Ask questions in discussions
- Reach out to maintainers

## 🏷️ Release Process

1. Updates go through feature branches
2. Changes are reviewed via Pull Requests
3. Releases follow semantic versioning
4. Changelogs are maintained
5. Releases are tagged and published

Thank you for contributing to WTF! 🎉

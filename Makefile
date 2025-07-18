# Build variables
BINARY_NAME=wtf
BUILD_DIR=build
VERSION=1.0.0-dev
GIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X cmd-finder/internal/version.Version=$(VERSION) -X cmd-finder/internal/version.GitHash=$(GIT_HASH) -X cmd-finder/internal/version.Build=$(BUILD_TIME)"

# Default target
.PHONY: all
all: test build

# Build the application
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build the application with optimizations for release
.PHONY: build-release
build-release:
	go build $(LDFLAGS) -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Run the application
.PHONY: run
run:
	go run main.go

# Test the application
.PHONY: test
test:
	go test ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Development helpers
.PHONY: dev-search
dev-search:
	go run main.go "$(QUERY)"

# Release preparation
.PHONY: prepare-release
prepare-release: clean test build-all
	@echo "Release artifacts ready in $(BUILD_DIR)/"

# Help
.PHONY: help
help:
	@echo "WTF (What's The Function) - Build System"
	@echo "========================================"
	@echo ""
	@echo "Development:"
	@echo "  build           - Build for current platform"
	@echo "  build-release   - Build optimized release version"
	@echo "  run             - Run the application"
	@echo "  test            - Run tests"
	@echo "  dev-search      - Quick search (use: make dev-search QUERY='your query')"
	@echo ""
	@echo "Release:"
	@echo "  build-all       - Build for all platforms (Linux, macOS, Windows)"
	@echo "  prepare-release - Clean, test, and build for all platforms"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  help            - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make dev-search QUERY='compress files'"
	@echo "  make prepare-release"
	@echo ""
	@echo "After building, users can run: wtf setup hey"
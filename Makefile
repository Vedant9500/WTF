# Build variables
BINARY_NAME=cmd-finder
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
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

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
	go run main.go search "$(QUERY)"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  run        - Run the application"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Install dependencies"
	@echo "  dev-search - Quick search (use: make dev-search QUERY='your query')"
	@echo "  help       - Show this help"
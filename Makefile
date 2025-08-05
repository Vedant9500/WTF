# Build variables
BINARY_NAME=wtf
BUILD_DIR=build
VERSION=1.2.0
GIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/Vedant9500/WTF/internal/version.Version=$(VERSION) -X github.com/Vedant9500/WTF/internal/version.GitHash=$(GIT_HASH) -X github.com/Vedant9500/WTF/internal/version.Build=$(BUILD_TIME)"
MAIN_PATH=./cmd/wtf

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building $(BINARY_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Cross-platform builds completed!"

# Default target
.PHONY: all
all: test build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build the application with optimizations for release
.PHONY: build-release
build-release:
	@echo "Building optimized release $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)/cmd/wtf

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building $(BINARY_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/wtf
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/wtf
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/wtf
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/wtf
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/wtf
	@echo "Cross-platform builds completed!"

# Create release packages
.PHONY: release
release: clean test build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/releases
	cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release packages created in $(BUILD_DIR)/releases/"

# Run the application
.PHONY: run
run:
	go run main.go

# Test the application
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test with verbose output
.PHONY: test-verbose
test-verbose:
	@echo "Running tests with verbose output..."
	go test ./... -v -cover

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	go test ./... -bench=. -benchmem

# Install to local Go bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install $(LDFLAGS) $(MAIN_PATH)

# Install to system (requires sudo)
.PHONY: install-system
install-system: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Creating system-wide command database directory..."
	sudo mkdir -p /usr/local/share/wtf
	sudo install -Dm644 assets/commands.yml /usr/local/share/wtf/commands.yml
	@echo "$(BINARY_NAME) installed successfully! You can now run 'wtf' from anywhere."

# Uninstall from system
.PHONY: uninstall-system
uninstall-system:
	@echo "Removing $(BINARY_NAME) from system..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -rf /usr/local/share/wtf
	@echo "$(BINARY_NAME) uninstalled successfully."

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

# Check version and build info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Git Hash: $(GIT_HASH)"
	@echo "Build Time: $(BUILD_TIME)"

# Help
.PHONY: help
help:
	@echo "WTF (What's The Function) - Build System v$(VERSION)"
	@echo "==================================================="
	@echo ""
	@echo "Development:"
	@echo "  build           - Build for current platform"
	@echo "  build-release   - Build optimized release version"
	@echo "  run             - Run the application"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-verbose    - Run tests with verbose output"
	@echo "  benchmark       - Run performance benchmarks"
	@echo "  dev-search      - Quick search (use: make dev-search QUERY='your query')"
	@echo ""
	@echo "Release:"
	@echo "  build-all       - Build for all platforms (Linux, macOS, Windows)"
	@echo "  release         - Create release packages for all platforms"
	@echo "  prepare-release - Clean, test, and build for all platforms"
	@echo "  install         - Install to local Go bin"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  version         - Show version information"
	@echo "  help            - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test-coverage"
	@echo "  make dev-search QUERY='compress files'"
	@echo "  make release"
	@echo ""
	@echo "After building, users can run: wtf setup hey"

# Release target
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v1.1.0"; \
		exit 1; \
	fi
	@chmod +x scripts/release.sh
	@scripts/release.sh $(VERSION)
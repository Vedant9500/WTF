#!/bin/bash

# Pre-commit hook for code quality checks
set -e

echo "Running pre-commit checks..."

# Format code
echo "Formatting code..."
go fmt ./...

# Run linter
echo "Running linter..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
else
    echo "Warning: golangci-lint not found. Please install it for full linting."
fi

# Run tests
echo "Running tests..."
go test ./... -short

echo "Pre-commit checks passed!"
---
allowed-tools: all
description: Generate or update Makefile with Go development targets
---

# Makefile Generator for Go Projects

Generate a comprehensive Makefile for Go projects with essential development targets including `make lint`, `make test`, and `make test-changed`.

## What this command does:

1. **Detects existing Makefile**: Checks if a Makefile already exists in the current directory
2. **Generates or updates**: Creates a new Makefile or updates existing one with Go-specific targets
3. **Includes essential targets**:
   - `make lint` - Run golangci-lint with fallback detection
   - `make test` - Run all tests with verbose output
   - `make fmt` - Format code with go fmt
   - `make vet` - Run go vet static analysis
   - `make build` - Build the application
   - `make clean` - Clean build artifacts
   - `make check` - Combined fmt, vet, and test

## Smart Features:

- **Project detection**: Automatically detects main package path
- **Git integration**: `test-changed` target intelligently finds changed Go files
- **Tool availability checks**: Gracefully handles missing tools (golangci-lint)
- **Cross-platform compatibility**: Works on Linux, macOS
- **Preservation**: Keeps existing custom targets when updating

## Usage:

Run this command in any Go project directory. It will:
- Create a new Makefile if none exists
- Update existing Makefile with missing Go targets
- Preserve any existing custom targets
- Use appropriate binary names and paths for your project

The generated Makefile follows Go best practices and integrates seamlessly with modern Go development workflows.

## Makefile Template

```makefile
# Variables
BINARY_NAME={{.BinaryName}}
MAIN_PATH={{.MainPath}}
BUILD_DIR=bin
GOARCH=$(shell go env GOARCH)

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-$(GOARCH) $(MAIN_PATH)
	GOOS=darwin GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-$(GOARCH) $(MAIN_PATH)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run go fmt
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint (requires golangci-lint to be installed)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Quick check (fmt, vet, test)
.PHONY: check
check: fmt vet test

# Development workflow (clean, check, build)
.PHONY: dev
dev: clean check build

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build (default)"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run tests"
	@echo "  test-changed - Run tests only for git changed files"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  deps         - Install and tidy dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run the application"
	@echo "  check        - Quick check (fmt, vet, test)"
	@echo "  dev          - Development workflow (clean, check, build)"
	@echo "  help         - Show this help message"
```

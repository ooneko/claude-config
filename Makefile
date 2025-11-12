# Variables
BINARY_NAME=claude-config
MAIN_PATH=./cmd/claude-config
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

# Install the binary to ~/go/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	@GOBIN=$$HOME/go/bin go install $(MAIN_PATH)
	@if ! echo "$$PATH" | grep -q "$$HOME/go/bin"; then \
		echo ""; \
		echo "⚠️  ~/go/bin is not in your PATH!"; \
		echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, ~/.profile):"; \
		echo ""; \
		echo "export PATH=\"\$$HOME/go/bin:\$$PATH\""; \
		echo ""; \
		echo "Then run: source ~/.bashrc (or your shell profile)"; \
		echo ""; \
	else \
		echo "✅ $(BINARY_NAME) installed successfully to ~/go/bin"; \
	fi

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
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  check        - Quick check (fmt, vet, test)"
	@echo "  dev          - Development workflow (clean, check, build)"
	@echo "  help         - Show this help message"
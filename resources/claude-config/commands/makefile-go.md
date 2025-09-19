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
   - `make test-changed` - Run tests only for git-changed Go files
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
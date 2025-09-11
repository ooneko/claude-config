# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is `claude-config`, a unified configuration management tool for Claude Code. It provides configuration management, proxy setup, hooks system control, NTFY notifications, DeepSeek API integration, and file operations.

## Architecture

### Core Structure
```
cmd/claude-config/     # CLI application entrypoint and command implementations
internal/              # Private packages
  ├── config/         # Configuration management
  ├── proxy/          # HTTP/HTTPS proxy management  
  ├── check/          # Hooks system validation
  ├── deepseek/       # DeepSeek API integration
  ├── file/           # File operations and merging
  ├── install/        # Installation and resource management
  └── claude/         # Core interfaces and types
resources/             # Template files, hooks, agents, and default configurations
```

### Key Managers
The application uses a manager pattern with these core components:
- `ConfigManager` - Handles Claude configuration settings
- `ProxyManager` - Manages HTTP/HTTPS proxy configurations
- `CheckManager` - Controls hooks system validation
- `DeepSeekManager` - Manages DeepSeek API integration

All managers are initialized in `main.go:init()` and operate on `~/.claude` directory.

## Development Commands

### Build and Test
```bash
# Build the application
go build ./cmd/claude-config

# Run all tests (note: some tests currently failing due to type issues)
go test ./...

# Run tests for specific package
go test ./internal/config
go test ./internal/proxy
go test ./internal/file
```

### Legacy Python Support
The project maintains Python compatibility through a Makefile:
```bash
# Install shell aliases for Python version
make install

# Run Python tests
make test

# Clean temporary files  
make clean
```

## CLI Commands Structure

The CLI is built with Cobra and provides these main commands:
- `status` - Show current configuration status
- `proxy` - Configure HTTP/HTTPS proxy settings
- `check` - Manage hooks system validation
- `deepseek` - DeepSeek API configuration
- `notify` - NTFY notification management
- `install` - Install configuration files and resources
- `backup` - Backup and restore configurations

## Resource Management

The `resources/` directory contains:
- **agents/** - Claude Code agent definitions (code-reviewer, golang-pro, etc.)
- **commands/** - Custom Claude commands (archviz, check, commit, etc.)
- **hooks/** - Shell hook scripts for linting, testing, and notifications
- **output-styles/** - Output formatting configurations
- **settings.json** - Default Claude settings
- **CLAUDE.md.template** - Template for project-specific Claude configurations

Resources are managed through the install system and can be deployed to user's `~/.claude` directory.

## Testing Strategy

- Use table-driven tests for complex logic
- Focus on testing manager interfaces and file operations
- Current test coverage includes file merging, configuration management, and proxy settings
- Some tests are currently failing due to type mismatches that need resolution

## Key Patterns

### Error Handling
Use simple error wrapping:
```go
return fmt.Errorf("operation failed: %w", err)
```

### Configuration Paths
All managers work with `~/.claude` as base directory, passed during initialization.

### File Operations
Use the `file` package for atomic operations, merging configurations, and handling Claude-specific file formats.
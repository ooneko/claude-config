# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
`claude-config` is a Go-based unified configuration management tool for Claude Code. It provides configuration management, proxy setup, hooks system validation, DeepSeek API integration, NTFY notifications, and resource installation through a manager-based architecture.

## Development Commands

### Build and Test
```bash
# Build the application
go build ./cmd/claude-config
# Or using make
make build

# Install to ~/go/bin
make install

# Run all tests
go test ./...
# Or using make
make test

# Run tests for specific packages
go test ./internal/config
go test ./internal/proxy
go test ./internal/file

# Run tests only for git-changed files (useful during development)
make test-changed

# Run tests with coverage
make test-coverage
```

### Code Quality and Development
```bash
# Format, lint, and test (complete workflow)
make check

# Format code
make fmt

# Run static analysis
make vet

# Run linter (requires golangci-lint)
make lint

# Development workflow: clean, check, build
make dev
```

### Cross-platform Building
```bash
# Build for multiple platforms
make build-all

# Manual cross-compilation examples
GOOS=linux GOARCH=amd64 go build ./cmd/claude-config
GOOS=darwin GOARCH=amd64 go build ./cmd/claude-config
GOOS=windows GOARCH=amd64 go build ./cmd/claude-config
```

## Architecture

### Manager Pattern
The application uses a manager-based architecture initialized in `main.go:init()`:

```go
// Key managers (all operate on ~/.claude directory)
configMgr     claude.ConfigManager      // Configuration management
proxyMgr      claude.ProxyManager       // HTTP/HTTPS proxy management
checkMgr      *check.Manager            // Hooks system validation
aiProviderMgr claude.AIProviderManager  // AI model provider configuration
```

All managers implement interfaces defined in `internal/claude/interfaces.go` and work with context-based operations.

### Core Package Structure
```
cmd/claude-config/     # CLI entrypoint and command implementations
├── main.go           # Manager initialization and application entry
├── commands.go       # Cobra command structure and routing
├── status.go         # Status command - shows configuration state
├── proxy.go          # Proxy management command
├── check.go          # Hooks system management
├── aiprovider.go     # AI model provider configuration
├── install.go        # Resource installation command
├── backup.go         # Backup and restore functionality
├── start.go          # Launch Claude Code with configured AI providers
├── notify.go         # Notification system (NTFY integration)
└── utils.go          # Utility functions

internal/             # Private packages following Go conventions
├── claude/          # Core interfaces and shared types
├── config/          # Settings.json management and configuration
├── proxy/           # HTTP/HTTPS proxy configuration management
├── check/           # Hooks system validation and control
├── aiprovider/      # AI model provider configuration (unified interface)
├── provider/        # Provider-related utilities and env mapping
├── file/            # File operations, merging, and atomic operations
├── install/         # Resource installation with embedded files

resources/           # Embedded resources using go:embed
└── claude-config/   # Resources deployed to ~/.claude
    ├── agents/      # Claude Code agent definitions
    ├── commands/    # Custom Claude commands
    ├── hooks/       # Shell hook scripts for development workflows
    └── settings.json # Default Claude settings template
```

### Resource Management System
The resource system uses Go's `embed` package to bundle files at compile time. Resources are installed atomically to `~/.claude` with backup and conflict resolution:

- **Embedded Resources**: Files bundled into binary using `//go:embed`
- **Template Processing**: Dynamic content generation with customization
- **Configuration Merging**: Intelligent merging of settings.json with conflict resolution
- **Atomic Operations**: All installations are atomic with rollback capability

### Configuration Directory Structure
All operations target `~/.claude` as the base configuration directory:
```
~/.claude/
├── settings.json              # Main Claude settings (merged intelligently)
├── claude_config.toml         # Tool-specific configuration
├── agents/                    # Custom agent definitions
├── commands/                  # Custom commands
├── hooks/                     # Development hooks (linting, testing)
└── output-styles/             # Output formatting styles
```

## CLI Command Architecture

Built with Cobra framework, commands follow this pattern:
- Each command has dedicated implementation file in `cmd/claude-config/`
- Commands are registered in `initCommands()` function in `commands.go`
- All commands operate through manager interfaces for testability
- Interactive commands use consistent prompting patterns

### Available Commands
- `status` - Comprehensive configuration status across all managers
- `proxy` - Interactive proxy configuration with connectivity validation
- `check` - Hooks system management with language-specific validation
- `aiprovider` - AI model provider configuration (supports DeepSeek, GLM, Doubao, Kimi)
- `start` - Launch Claude Code with configured AI providers
- `install` - Resource installation with selective component installation (supports --force flag)
- `backup` - Configuration backup and restore with versioning
- `notify` - Notification system configuration (NTFY integration)

## Testing Architecture

### Test Strategy
- **Manager Interface Testing**: Focus on testing through interfaces defined in `internal/claude/`
- **Table-Driven Tests**: Used extensively for complex logic validation
- **Context-Based Testing**: All managers use context for cancellation and timeouts
- **File Operation Testing**: Comprehensive testing of atomic operations and merging logic

### Test Organization
```bash
# Test packages individually (faster during development)
go test ./internal/config      # Configuration management tests
go test ./internal/proxy       # Proxy functionality tests
go test ./internal/file        # File operations and merging tests
go test ./internal/install     # Resource installation tests
go test ./internal/deepseek    # DeepSeek API integration tests
go test ./internal/aiprovider  # AI provider configuration tests
go test ./internal/provider    # Provider-related utilities tests
```

## AI Provider Architecture

The application implements a unified AI provider architecture with support for multiple AI model providers:

### Supported Providers
- **DeepSeek** - DeepSeek AI models
- **GLM** - ZhiPu AI models (GLM series)
- **Doubao** - ByteDance Doubao models
- **Kimi** - Moonshot AI Kimi models

### Key Features
- Centralized configuration for all AI providers
- Consistent API interface across different providers
- Automatic provider selection based on configuration
- Support for provider-specific settings (API endpoints, models, etc.)
- Easy extension for adding new AI providers
- Environment variable mapping utilities in `internal/provider/` for flexible configuration

## Key Development Patterns

### Manager Initialization
All managers are initialized in `main.go:init()` with the Claude directory path:
```go
claudeDir := filepath.Join(homeDir, ".claude")
configMgr = config.NewManager(claudeDir)
```

### Context Usage
All manager operations use context for proper cancellation and timeout handling:
```go
func (m *Manager) Operation(ctx context.Context) error {
    // Implementation with context awareness
}
```

### Interface-Based Design
Core functionality is defined through interfaces in `internal/claude/interfaces.go`, enabling:
- Easy testing with mock implementations  
- Clear separation of concerns
- Consistent API across all managers

### Error Handling
Use simple error wrapping with context:
```go
return fmt.Errorf("operation failed: %w", err)
```
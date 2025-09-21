# claude-config

[中文文档](README_CN.md) | [English Documentation](README.md)

A modern, unified configuration management tool for Claude Code written in Go. Provides comprehensive management for Claude Code settings, proxy setup, validation system control, AI provider integration, NTFY notifications, and resource installation.

## Features

- **Configuration Management** - Manage Claude Code settings and configurations
- **Proxy Setup** - Configure HTTP/HTTPS proxy settings with validation
- **Validation System** - Advanced development workflow validation and management
- **AI Provider Integration** - Multi-provider AI API configuration and management (DeepSeek, Kimi, GLM4.5)
- **NTFY Notifications** - Configure notification systems for development workflows
- **Resource Management** - Install and manage agents, commands, hooks, and templates
- **Backup & Restore** - Complete configuration backup and restoration system

## Installation

### Direct Install from GitHub (Easiest)
```bash
# Install directly from GitHub (requires Go 1.21+)
go install github.com/ooneko/claude-config/cmd/claude-config@latest
```

After installation, you can run the tool from anywhere:
```bash
claude-config --help
claude-config status
```

**Note**: Ensure `~/go/bin` is in your PATH. If not, add this line to your shell profile:
```bash
export PATH="$HOME/go/bin:$PATH"
```

### Local Install with Make (Recommended for Development)
```bash
# Clone the repository
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# Install to ~/go/bin (adds to PATH automatically)
make install
```

The `make install` command will install the binary to `~/go/bin` and provide PATH setup instructions if needed.

### Build from Source
```bash
# Clone the repository
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# Build the binary locally
go build ./cmd/claude-config

# Run the tool
./claude-config --help
```

### System Requirements
- Go 1.21 or later
- Access to `~/.claude` directory (Claude Code configuration directory)

## Usage

### Main Commands

```bash
# Show current configuration status
claude-config status

# Configure proxy settings (interactive)
claude-config proxy

# Manage validation system
claude-config check

# Configure NTFY notifications
claude-config notify

# Configure AI provider integration
claude-config ai

# Setup NTFY notifications
claude-config notify

# Manage validation system
claude-config check

# Install resources (agents, commands, hooks, templates)
claude-config install

# Backup and restore configurations
claude-config backup
```

**Note**: If you built from source instead of using `make install`, prefix commands with `./` (e.g., `./claude-config status`).

### Command Examples

```bash
# Check current status of all configurations
claude-config status

# Interactive proxy configuration with validation
claude-config proxy
# Sets up HTTP_PROXY and HTTPS_PROXY environment variables
# Validates proxy connectivity

# Install all available resources to ~/.claude
claude-config install
# Installs: agents, commands, hooks, output-styles, settings

# Configure AI provider (DeepSeek, Kimi, GLM4.5)
claude-config ai on deepseek
# Interactive setup and connection testing

# Enable validation system
claude-config check on
# Configures validation with language-specific linting and testing
```

## Command Reference

### Main Commands

#### `status` - Show Configuration Status
```bash
claude-config status
```
Displays comprehensive status of all configurations including proxy, AI providers, validation system, and notifications.

#### `proxy` - Proxy Management
```bash
# Interactive proxy configuration
claude-config proxy on

# Disable proxy
claude-config proxy off

# Toggle proxy status
claude-config proxy toggle
```
Manages HTTP/HTTPS proxy settings with validation and connectivity testing.

#### `ai` - AI Provider Management
```bash
# Enable specific AI provider (with API key prompt if needed)
claude-config ai on deepseek
claude-config ai on kimi
claude-config ai on zhipu

# Disable all AI providers
claude-config ai off

# Reset specific provider (removes API key)
claude-config ai reset deepseek

# List all supported providers
claude-config ai list

# Show current AI provider status
claude-config ai
```
Supports multiple AI providers: DeepSeek, Kimi (Moonshot), and GLM4.5 (ZhipuAI).

#### `check` - Validation System Management
```bash
# Enable validation system
claude-config check on

# Disable validation system
claude-config check off
```
Manages development validation for linting, testing, and code quality checks.

#### `notify` - NTFY Notifications
```bash
# Enable NTFY notifications
claude-config notify on

# Disable NTFY notifications
claude-config notify off
```
Configures NTFY notification system for development workflows.

#### `install` - Resource Installation
```bash
# Install all resources
claude-config install

# Install with force flag (overwrite existing)
claude-config install --force
```
Installs agents, commands, validation hooks, output-styles, and settings to `~/.claude`.

#### `backup` - Configuration Backup
```bash
claude-config backup
```
Creates backups and restores Claude Code configurations.

## Project Structure

```
claude-config/
├── cmd/claude-config/          # CLI application and command implementations
│   ├── main.go                # Application entrypoint
│   ├── commands.go            # Command structure and initialization
│   ├── status.go              # Status command implementation
│   ├── proxy.go               # Proxy management command
│   ├── check.go               # Validation system management
│   ├── aiprovider.go          # AI provider integration
│   ├── notify.go              # NTFY notifications setup
│   ├── install.go             # Resource installation command
│   └── backup.go              # Backup and restore functionality
├── internal/                   # Private packages (Go internal convention)
│   ├── config/                # Configuration file management
│   ├── proxy/                 # HTTP/HTTPS proxy management
│   ├── check/                 # Validation system
│   ├── aiprovider/            # AI provider client and configuration
│   ├── file/                  # File operations and merging utilities
│   ├── install/               # Resource installation and management
│   └── claude/                # Core interfaces and shared types
└── resources/                  # Embedded resources and templates
    └── claude-config/         # Resource files for installation
        ├── agents/            # Claude Code agent definitions
        ├── commands/          # Custom Claude commands
        ├── hooks/             # Shell hook scripts
        ├── output-styles/     # Output formatting configurations
        ├── settings.json      # Default Claude settings
        └── CLAUDE.md.template # Template for project configurations
```

## Development

### Build and Test

```bash
# Build the application
go build ./cmd/claude-config

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/config
go test ./internal/proxy
go test ./internal/file

# Test with race detection
go test -race ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build ./cmd/claude-config
GOOS=darwin GOARCH=amd64 go build ./cmd/claude-config
GOOS=windows GOARCH=amd64 go build ./cmd/claude-config
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Install and run golangci-lint
golangci-lint run
```

## Architecture

### Manager Pattern
The application uses a manager-based architecture with these core components:

- **ConfigManager** (`internal/config`) - Handles Claude configuration settings
- **ProxyManager** (`internal/proxy`) - Manages HTTP/HTTPS proxy configurations
- **CheckManager** (`internal/check`) - Controls validation system
- **AIProviderManager** (`internal/aiprovider`) - Manages multi-provider AI API integration

All managers are initialized in `main.go:init()` and operate on the `~/.claude` directory.

### Resource System
The resource system (`internal/install`) provides:
- Embedded resource files using Go embed
- Template processing and customization
- Atomic file operations with backup
- Configuration merging with conflict resolution

### Configuration Directory
All operations work with `~/.claude` as the base configuration directory:

```
~/.claude/
├── settings.json              # Main Claude settings
├── claude_config.toml         # Tool-specific configuration
├── agents/                    # Custom agent definitions
├── commands/                  # Custom commands
├── hooks/                     # Development validation hooks
└── output-styles/             # Output formatting styles
```

## Contributing

### Development Guidelines
- Follow standard Go project structure
- Use meaningful package and function names
- Write tests for all new functionality
- Ensure all tests pass before submitting PRs
- Use Go modules for dependency management

### Adding New Commands
1. Create command implementation in `cmd/claude-config/`
2. Add command to `initCommands()` in `commands.go`
3. Create corresponding manager in `internal/` if needed
4. Add tests for the new functionality
5. Update documentation

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Note**: This tool manages your Claude Code configuration in `~/.claude`. Always backup your configurations before making changes.
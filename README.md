# claude-config

[中文文档](README_CN.md) | [English Documentation](README.md)

A unified configuration management tool for Claude Code that provides configuration management, proxy setup, hooks system control, NTFY notifications, DeepSeek API integration, and file operations.

## Features

- **Configuration Management** - Manage Claude Code settings and configurations
- **Proxy Setup** - Configure HTTP/HTTPS proxy settings
- **Hooks System** - Control validation hooks and automated checks
- **DeepSeek API Integration** - Manage DeepSeek API configurations
- **NTFY Notifications** - Configure notification systems
- **File Operations** - Handle file merging and operations
- **Resource Management** - Install and manage agents, commands, and templates

## Installation

### Go Installation
```bash
go build ./cmd/claude-config
```

### Legacy Python Support
```bash
make install
```

## Usage

### Main Commands

```bash
# Show current configuration status
./claude-config status

# Configure proxy settings
./claude-config proxy

# Manage hooks system
./claude-config check

# Configure DeepSeek API
./claude-config deepseek

# Manage notifications
./claude-config notify

# Install resources and configurations
./claude-config install

# Backup and restore configurations
./claude-config backup
```

## Project Structure

```
cmd/claude-config/     # CLI application entrypoint
internal/              # Private packages
  ├── config/         # Configuration management
  ├── proxy/          # HTTP/HTTPS proxy management  
  ├── check/          # Hooks system validation
  ├── deepseek/       # DeepSeek API integration
  ├── file/           # File operations and merging
  ├── install/        # Installation and resource management
  └── claude/         # Core interfaces and types
resources/             # Template files, hooks, agents, and configurations
  ├── agents/         # Claude Code agent definitions
  ├── commands/       # Custom Claude commands
  ├── hooks/          # Shell hook scripts
  ├── output-styles/  # Output formatting configurations
  └── settings.json   # Default Claude settings
```

## Development

### Build and Test
```bash
# Build
go build ./cmd/claude-config

# Run tests
go test ./...

# Run specific package tests
go test ./internal/config
go test ./internal/proxy
go test ./internal/file
```

### Legacy Python Commands
```bash
make test    # Run Python tests
make clean   # Clean temporary files
```

## Configuration Directory

All operations work with the `~/.claude` directory as the base configuration location.

## License

[License information to be added]
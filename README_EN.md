# claude-config

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=flat-square)](https://github.com/ooneko/claude-config)

**[简体中文](README.md) | [English](README_EN.md)**

<p align="center">
  <img src="logo.png" alt="claude-config logo" width="300"/>
</p>

<p align="center">
  <strong>Modern Claude Configuration Management</strong>
</p>

A modern, unified configuration management tool for Claude Code written in Go. Making configuration management simple and efficient.

## ✨ Quick Start

Get started with claude-config in just 3 steps:

```bash
# 1️⃣ Install (the easiest way)
go install github.com/ooneko/claude-config/cmd/claude-config@latest

# 2️⃣ Install resources (one-click setup for all components)
claude-config install

# 3️⃣ Check status (confirm configuration is complete)
claude-config status
```

🎉 **Done!** Your Claude Code environment is now configured.

## 📖 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Usage Examples](#-usage-examples)
- [Command Reference](#-command-reference)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Features

### Core Functionality
- 🎯 **Configuration Management** - One-click management of Claude Code settings
- 🌐 **Proxy Setup** - Smart HTTP/HTTPS proxy configuration with connection validation
- ✅ **Validation System** - Advanced development workflow validation and code quality checks
- 🤖 **AI Provider Integration** - Support for DeepSeek, Kimi, GLM4.5, Doubao and more AI providers
- 🔔 **NTFY Notifications** - Configure real-time notifications for development workflows
- 📦 **Resource Management** - Install and manage agents, commands, hooks and development resources
- 💾 **Backup & Restore** - Complete configuration backup and one-click restoration system

### Why choose claude-config?
- ⚡ **Extremely Easy** - Complete configuration with a single command
- 🔧 **Smart Management** - Automatic detection and resolution of configuration conflicts
- 🛡️ **Secure & Reliable** - Atomic operations ensure configuration integrity
- 🌍 **Cross-Platform** - Support for Linux, macOS, Windows

## 📦 Installation

### 🚀 Method 1: Direct Install (Recommended for Beginners)

The simplest and fastest installation method, no repository cloning required:

```bash
# One-click install (requires Go 1.21+)
go install github.com/ooneko/claude-config/cmd/claude-config@latest
```

Ready to use immediately after installation:
```bash
claude-config --help
claude-config status
```

> 💡 **PATH Tip**: Make sure `~/go/bin` is in your PATH. If not, add:
> ```bash
> export PATH="$HOME/go/bin:$PATH"
> ```

### 🔧 Method 2: Local Build (Recommended for Developers)

Ideal for users who want to modify code or develop locally:

```bash
# 1. Clone repository
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# 2. Install to system
make install
```

`make install` will automatically handle PATH configuration.

### 🏗️ Method 3: Build from Source

Full control over the build process:

```bash
# 1. Get source code
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# 2. Build binary
go build ./cmd/claude-config

# 3. Test run
./claude-config --help
```

### 📋 System Requirements

- ✅ **Go 1.21** or later
- ✅ **Permissions**: Access to `~/.claude` directory (Claude Code configuration directory)
- ✅ **Systems**: Linux / macOS / Windows

---

## 💡 Usage Examples

### 🎯 Basic Workflow

```bash
# 1️⃣ First time: Install all resources
claude-config install

# 2️⃣ Check current configuration status
claude-config status

# 3️⃣ Configure proxy (if needed)
claude-config proxy on

# 4️⃣ Configure AI provider (e.g., DeepSeek)
claude-config ai on deepseek

# 5️⃣ Start Claude Code (optional)
claude-config start
```

### 🌐 Proxy Configuration Examples

```bash
# Interactive proxy configuration
claude-config proxy

# Quick enable proxy
claude-config proxy on

# Toggle proxy status
claude-config proxy toggle

# Disable proxy
claude-config proxy off
```

### 🤖 AI Provider Configuration Examples

```bash
# Configure DeepSeek
claude-config ai on deepseek

# Configure Kimi (Moonshot AI)
claude-config ai on kimi

# Configure Zhipu GLM
claude-config ai on glm

# Configure Doubao (ByteDance)
claude-config ai on doubao

# View current AI configuration
claude-config ai

# Reset specific provider
claude-config ai reset deepseek
```

### 🚀 Launch Claude Code Examples

```bash
# Launch native Claude Code (clears all AI configurations)
claude-config start

# Launch with configured DeepSeek
claude-config start deepseek

# Launch with configured Kimi, specify model
claude-config start kimi --model kimi-plus

# Launch with GLM, temporary API key
claude-config start glm --api-key sk-xxxxxxxx

# Launch with Doubao, specify model and API key
claude-config start doubao --model doubao-pro --api-key your-api-key
```

## 📚 Command Reference

### 🎯 Core Commands Overview

| Command | Function | Quick Example |
|---------|----------|---------------|
| `install` | Install all resources | `claude-config install` |
| `status` | View configuration status | `claude-config status` |
| `proxy` | Proxy configuration management | `claude-config proxy on` |
| `ai` | AI provider configuration | `claude-config ai on deepseek` |
| `check` | Validation system control | `claude-config check on` |
| `notify` | Notification system configuration | `claude-config notify on` |
| `start` | Launch Claude Code | `claude-config start` |
| `backup` | Backup and restore configuration | `claude-config backup` |

### 📋 Detailed Command Documentation

#### `claude-config install` - Resource Installation
One-click installation of all development resources to `~/.claude`:
```bash
# Install all resources (agents, commands, hooks, templates, etc.)
claude-config install

# Force overwrite installation (use with caution)
claude-config install --force
```

#### `claude-config status` - Configuration Status
View the current status of all configurations:
```bash
claude-config status
```
Example output:
```
✅ Configuration Files: Ready
🤖 AI Providers: DeepSeek (Connected)
🌐 Proxy Configuration: Disabled
✅ Validation System: Enabled
🔔 Notification System: Enabled
```

#### `claude-config proxy` - Proxy Management
Intelligent proxy configuration and validation:
```bash
# Interactive configuration
claude-config proxy

# Quick enable
claude-config proxy on

# Toggle status
claude-config proxy toggle

# Complete disable
claude-config proxy off
```

#### `claude-config ai` - AI Provider Management
Support for multiple AI service providers:
```bash
# Enable specific provider (will prompt for API key if needed)
claude-config ai on deepseek    # DeepSeek AI
claude-config ai on kimi        # Kimi (Moonshot AI)
claude-config ai on glm         # Zhipu GLM
claude-config ai on doubao      # Doubao (ByteDance)

# List all supported providers
claude-config ai list

# View current configuration
claude-config ai

# Disable all AI providers
claude-config ai off

# Reset specific provider (remove API key)
claude-config ai reset deepseek
```

#### `claude-config check` - Validation System
Control code quality checks:
```bash
# Enable development validation
claude-config check on

# Disable validation system
claude-config check off
```

#### `claude-config notify` - Notification System
Configure NTFY real-time notifications:
```bash
# Enable notifications
claude-config notify on

# Disable notifications
claude-config notify off
```

#### `claude-config start` - Launch Claude Code
Intelligent launch of Claude Code with multiple modes:
```bash
# Launch native Claude Code (clears all AI configurations)
claude-config start

# Launch with configured AI providers
claude-config start deepseek    # Use DeepSeek
claude-config start kimi        # Use Kimi
claude-config start glm         # Use GLM
claude-config start doubao      # Use Doubao

# Advanced options (temporary override configurations)
claude-config start kimi --model kimi-plus              # Specify model
claude-config start glm --api-key sk-xxxxxxxx           # Temporary API key
claude-config start doubao --model pro --key your-key   # Specify both model and key
```

**Features:**
- 🔄 **Smart Switching** - Launch native Claude without arguments, use specified AI with arguments
- 🔐 **Key Management** - Prioritizes stored keys, supports temporary override
- 🎯 **Model Selection** - Supports temporary specification of different models
- 🧹 **Configuration Cleanup** - Automatically clears existing configurations when launching native version

#### `claude-config backup` - Configuration Backup
Safe backup and restore:
```bash
# Create configuration backup
claude-config backup

# View restore options
claude-config backup --help
```

> 💡 **Tip**: If you built from source, prefix commands with `./` (e.g., `./claude-config status`)



## 🤝 Contributing

We welcome all forms of contributions! Whether it's bug reports, feature suggestions, or code contributions.

### 🚀 Getting Started

```bash
# 1. Fork and clone the repository
git clone https://github.com/your-username/claude-config.git
cd claude-config

# 2. Create a feature branch
git checkout -b feature/amazing-feature

# 3. Develop and test
make dev  # Run the complete development workflow

# 4. Commit your changes
git commit -m "feat: add amazing feature"

# 5. Push and create PR
git push origin feature/amazing-feature
```

### 📋 Development Guidelines

#### Code Quality
- ✅ **Follow Go Project Structure** - Use standard Go project layout
- ✅ **Meaningful Naming** - Function and variable names should clearly express intent
- ✅ **Write Tests** - New features must have corresponding test cases
- ✅ **Pass Checks** - Run `make check` before committing to ensure code quality

#### Commit Guidelines
- 🎯 **Clear Messages** - Commit messages should explain the purpose of changes
- 🔍 **Atomic Commits** - One commit should do one thing
- 📝 **Update Documentation** - Update relevant documentation for major changes

#### Adding New Features
1. **Implement Command** - Create command file in `cmd/claude-config/`
2. **Register Command** - Register in `initCommands()` in `commands.go`
3. **Create Manager** - Create corresponding manager in `internal/` if needed
4. **Write Tests** - Add comprehensive tests for new functionality
5. **Update Documentation** - Update README and related documentation

### 🛠️ Development Tools

```bash
# Development workflow (format, check, build)
make dev

# Run tests
make test

# Code coverage
make test-coverage

# Code quality checks
make check

# Build multi-platform binaries
make build-all
```

### 🐛 Reporting Issues

Found a bug? Please report it through:

- 📋 **Issue Templates** - Use the bug template in GitHub Issues
- 🔍 **Detailed Information** - Provide reproduction steps and environment details
- 📸 **Screenshots** - Include relevant screenshots if possible
- 💻 **Logs** - Attach relevant error logs

---

## 📄 License

This project is licensed under the [Apache License 2.0](LICENSE).

### 📋 License Summary

- ✅ **Commercial Use** - Can be used in commercial projects
- ✅ **Modification** - Can modify source code
- ✅ **Distribution** - Can distribute original or modified versions
- ✅ **Private Use** - Can use privately without open sourcing
- ⚠️ **Responsibility** - Must retain original author copyright notices
- ⚠️ **Patents** - Provides patent grant

---

## ⚠️ Important Notice

This tool manages your Claude Code configuration in `~/.claude`. Before making major changes, **we strongly recommend backing up your configuration**:

```bash
# Create configuration backup
claude-config backup
```

## 🙏 Acknowledgments

Thanks to all the developers and users who contribute to the claude-config project!

---

<div align="center">

**[⬆️ Back to Top](#claude-config)**

Made with ❤️ by the claude-config community

</div>
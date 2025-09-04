# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 📁 Project Overview

This is a Claude Code configuration management repository that provides a complete automation solution for development workflows. It integrates intelligent hooks, professional agents, custom commands, and dynamic configuration management.

## 🚀 Core Commands

### Configuration Management
- `./claude-config.sh` - 动态配置管理脚本
  - `claude-config status` - 显示当前配置状态 
  - `claude-config proxy on/off` - 代理开关管理
  - `claude-config hooks on/off` - Hooks 开关管理
  - `claude-config deepseek on/off` - DeepSeek API 配置管理
  - `claude-config backup` - 备份配置

### Development Tools
- `./copy_to_claude.py` - 智能文件复制工具，支持项目结构分析和安全检查

## 📋 Project Architecture

### Directory Structure
- `agents/` - 7个专业智能代理 (ai-engineer, backend-developer, code-reviewer, frontend-developer, golang-pro, product-manager, vue-expert)
- `commands/` - 6个自定义命令 (archviz, check, commit, next, prompt, ultrathink)  
- `hooks/` - 智能自动化钩子脚本系统
- `output-styles/` - 专业输出样式规范
- `settings.json` - Claude Code核心配置文件

### Hook System
The intelligent hook system automatically runs quality checks:
- **PostToolUse hooks**: 在代码编辑后自动运行
  - `smart-lint.sh` - 智能检测项目类型并运行相应 linter
  - `smart-test.sh` - 根据修改文件智能运行相关测试
- **Stop hooks**: 会话结束时运行通知脚本

### Configuration Management
- **Dynamic proxy management**: HTTP/HTTPS 代理一键开关 (127.0.0.1:7890)
- **DeepSeek API integration**: 安全的 API 密钥管理和自动模型配置
- **Hook system toggle**: 智能钩子系统开关管理

## 🛠️ Key Features

### Intelligent Project Detection
The `smart-lint.sh` automatically detects project types and runs appropriate tools:
- **Go projects**: `golangci-lint run`
- **Node.js projects**: `npm run lint` or `eslint`
- **Python projects**: `ruff check` or `flake8`
- **Tilt projects**: `tilt verify`

### Professional Agents
Each agent is specialized for specific domains:
- Use `ai-engineer` for AI system design and model implementation
- Use `backend-developer` for scalable API development
- Use `code-reviewer` for pragmatic code review
- Use `frontend-developer` for React/UI development
- Use `golang-pro` for Go development and concurrent programming
- Use `product-manager` for product strategy
- Use `vue-expert` for Vue 3 and Nuxt development

### Custom Commands
- `/archviz` - 生成交互式HTML架构图
- `/check` - 运行代码质量检查
- `/commit` - 智能提交管理
- `/next` - 任务规划助手
- `/prompt` - 项目特定提示生成器
- `/ultrathink` - 深度思考模式

## 🔧 Development Workflow

### Configuration Updates
When modifying `settings.json`:
1. Always backup first: `./claude-config.sh backup`
2. Use the configuration script rather than direct editing
3. Verify changes with: `./claude-config.sh status`

### Hook Development
When working on hook scripts in `hooks/`:
- All scripts must follow the common exit code pattern (0=success, 1=error, 2=issues found)
- Source `common-helpers.sh` for shared functionality
- Test hooks thoroughly as they run automatically

### Agent and Command Development
When modifying agents or commands:
- Follow the existing markdown structure with frontmatter
- Include clear descriptions and tool permissions
- Test functionality before committing changes

## 📊 Important Files

- `claude-config.sh` - Main configuration management script
- `settings.json` - Core Claude Code configuration
- `copy_to_claude.py` - Intelligent file copying utility
- `hooks/smart-lint.sh` - Intelligent linting system
- `hooks/smart-test.sh` - Intelligent testing system
- `hooks/common-helpers.sh` - Shared utilities for hooks

## 🚨 Critical Requirements

- **All hook issues are BLOCKING** - Every error must be fixed before proceeding
- **Configuration changes require backup** - Always backup before modifications
- **Python scripts require proper dependencies** - Ensure required packages are installed
- **Hooks must maintain exit code standards** - 0=success, 1=error, 2=issues found
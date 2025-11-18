# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 项目概述

`claude-config` 是一个基于 Go 的统一配置管理工具，专为 Claude Code 设计。它通过基于管理器的架构提供配置管理、代理设置、钩子系统验证、DeepSeek API 集成、NTFY 通知和资源安装功能。

## 开发命令

### 构建和测试
```bash
# 构建应用程序
go build ./cmd/claude-config
# 或者使用 make
make build

# 安装到 ~/go/bin
make install

# 运行所有测试
go test ./...
# 或者使用 make
make test

# 运行特定包的测试
go test ./internal/config
go test ./internal/proxy
go test ./internal/file

# 仅对 git 更改的文件运行测试（开发期间有用）
make test-changed

# 运行测试并生成覆盖率报告
make test-coverage
```

### 代码质量和开发
```bash
# 格式化、静态分析和测试（完整工作流）
make check

# 格式化代码
make fmt

# 运行静态分析
make vet

# 运行代码检查器（需要 golangci-lint）
make lint

# 开发工作流：清理、检查、构建
make dev
```

### 跨平台构建
```bash
# 为多个平台构建
make build-all

# 手动跨编译示例
GOOS=linux GOARCH=amd64 go build ./cmd/claude-config
GOOS=darwin GOARCH=amd64 go build ./cmd/claude-config
GOOS=windows GOARCH=amd64 go build ./cmd/claude-config
```

## 架构

### 管理器模式
应用程序使用基于管理器的架构，在 `main.go:init()` 中初始化：

```go
// 核心管理器（都操作 ~/.claude 目录）
configMgr     claude.ConfigManager      // 配置管理
proxyMgr      claude.ProxyManager       // HTTP/HTTPS 代理管理
checkMgr      *check.Manager            // 钩子系统验证
aiProviderMgr claude.AIProviderManager  // AI 模型提供商配置
```

所有管理器都实现 `internal/claude/interfaces.go` 中定义的接口，并支持基于上下文的操作。

### 核心包结构
```
cmd/claude-config/     # CLI 入口点和命令实现
├── main.go           # 管理器初始化和应用程序入口
├── commands.go       # Cobra 命令结构和路由
├── status.go         # 状态命令 - 显示配置状态
├── proxy.go          # 代理管理命令
├── check.go          # 钩子系统管理
├── aiprovider.go     # AI 模型提供商配置
├── install.go        # 资源安装命令
├── backup.go         # 备份和恢复功能
├── start.go          # 使用配置的 AI 提供商启动 Claude Code
├── notify.go         # 通知系统（NTFY 集成）
└── utils.go          # 实用函数

internal/             # 遵循 Go 约定的私有包
├── claude/          # 核心接口和共享类型
├── config/          # settings.json 管理和配置
├── proxy/           # HTTP/HTTPS 代理配置管理
├── check/           # 钩子系统验证和控制
├── aiprovider/      # AI 模型提供商配置（统一接口）
├── provider/        # 提供商相关工具和环境变量映射
├── file/            # 文件操作、合并和原子操作
├── install/         # 带嵌入文件的资源安装

resources/           # 使用 go:embed 的嵌入资源
└── claude-config/   # 部署到 ~/.claude 的资源
    ├── agents/      # Claude Code 代理定义
    ├── commands/    # 自定义 Claude 命令
    ├── hooks/       # 开发工作流的 Shell 钩子脚本
    └── settings.json # 默认 Claude 设置模板
```

### 资源管理系统
资源系统使用 Go 的 `embed` 包在编译时捆绑文件。资源被原子地安装到 `~/.claude`，具有备份和冲突解决功能：

- **嵌入资源**：使用 `//go:embed` 捆绑到二进制文件中的文件
- **模板处理**：支持自定义的动态内容生成
- **配置合并**：智能合并 settings.json 并解决冲突
- **原子操作**：所有安装都是原子的，具有回滚能力

### 配置目录结构
所有操作都以 `~/.claude` 作为基础配置目录：
```
~/.claude/
├── settings.json              # 主要 Claude 设置（智能合并）
├── claude_config.toml         # 工具特定配置
├── agents/                    # 自定义代理定义
├── commands/                  # 自定义命令
├── hooks/                     # 开发钩子（代码检查、测试）
└── output-styles/             # 输出格式化样式
```

## CLI 命令架构

使用 Cobra 框架构建，命令遵循以下模式：
- 每个命令在 `cmd/claude-config/` 中都有专用的实现文件
- 命令在 `commands.go` 的 `initCommands()` 函数中注册
- 所有命令都通过管理器接口操作以确保可测试性
- 交互式命令使用一致的提示模式

### 可用命令
- `status` - 跨所有管理器的全面配置状态
- `proxy` - 带连接性验证的交互式代理配置
- `check` - 钩子系统管理，支持语言特定验证
- `aiprovider` - AI 模型提供商配置（支持 DeepSeek、GLM、Doubao、Kimi）
- `start` - 使用配置的 AI 提供商启动 Claude Code
- `install` - 选择性组件安装的资源安装（支持 --force 标志）
- `backup` - 带版本控制的配置备份和恢复
- `notify` - 通知系统配置（NTFY 集成）

## 测试架构

### 测试策略
- **管理器接口测试**：专注于通过 `internal/claude/` 中定义的接口进行测试
- **表驱动测试**：广泛用于复杂逻辑验证
- **基于上下文的测试**：所有管理器都使用上下文进行取消和超时处理
- **文件操作测试**：全面测试原子操作和合并逻辑

### 测试组织
```bash
# 单独测试包（开发期间更快）
go test ./internal/config      # 配置管理测试
go test ./internal/proxy       # 代理功能测试
go test ./internal/file        # 文件操作和合并测试
go test ./internal/install     # 资源安装测试
go test ./internal/deepseek    # DeepSeek API 集成测试
go test ./internal/aiprovider  # AI 提供商配置测试
go test ./internal/provider    # 提供商相关工具测试
```

## AI 提供商架构

应用程序实现了统一的 AI 提供商架构，支持多个 AI 模型提供商：

### 支持的提供商
- **DeepSeek** - DeepSeek AI 模型
- **GLM** - 智谱 AI 模型（GLM 系列）
- **Doubao** - 字节跳动豆包模型
- **Kimi** - Moonshot AI Kimi 模型

### 关键特性
- 所有 AI 提供商的集中化配置
- 跨不同提供商的一致 API 接口
- 基于配置的自动提供商选择
- 支持提供商特定设置（API 端点、模型等）
- 易于添加新 AI 提供商的扩展
- `internal/provider/` 中的环境变量映射工具，支持灵活配置

## 关键开发模式

### 管理器初始化
所有管理器都在 `main.go:init()` 中使用 Claude 目录路径初始化：
```go
claudeDir := filepath.Join(homeDir, ".claude")
configMgr = config.NewManager(claudeDir)
```

### 上下文使用
所有管理器操作都使用上下文进行适当的取消和超时处理：
```go
func (m *Manager) Operation(ctx context.Context) error {
    // 具有上下文感知的实现
}
```

### 基于接口的设计
核心功能通过 `internal/claude/interfaces.go` 中的接口定义，支持：
- 使用模拟实现的轻松测试
- 关注点的清晰分离
- 所有管理器的一致 API

### 错误处理
使用简单的错误包装并提供上下文：
```go
return fmt.Errorf("操作失败: %w", err)
```
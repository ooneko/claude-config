# claude-config

[中文文档](README_CN.md) | [English Documentation](README.md)

一个用Go编写的现代化、统一的Claude Code配置管理工具。提供Claude Code设置、代理配置、验证系统控制、AI提供商集成、NTFY通知和资源安装的全面管理功能。

## 功能特性

- **配置管理** - 管理Claude Code设置和配置
- **代理设置** - 配置HTTP/HTTPS代理设置并进行验证
- **验证系统** - 高级开发工作流验证和管理
- **AI提供商集成** - 多提供商AI API配置和管理（DeepSeek、Kimi、GLM4.5、Doubao）
- **NTFY通知** - 为开发工作流配置通知系统
- **资源管理** - 安装和管理代理、命令、钩子和模板
- **备份与恢复** - 完整的配置备份和恢复系统

## 安装

### 从GitHub直接安装（最简单）
```bash
# 直接从GitHub安装（需要Go 1.21+）
go install github.com/ooneko/claude-config/cmd/claude-config@latest
```

安装完成后，你可以在任何地方运行该工具：
```bash
claude-config --help
claude-config status
```

**注意**：确保`~/go/bin`在你的PATH中。如果没有，请将此行添加到你的shell配置文件中：
```bash
export PATH="$HOME/go/bin:$PATH"
```

### 使用Make本地安装（推荐用于开发）
```bash
# 克隆仓库
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# 安装到~/go/bin（自动添加到PATH）
make install
```

`make install`命令将二进制文件安装到`~/go/bin`，如果需要会提供PATH设置说明。

### 从源码构建
```bash
# 克隆仓库
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# 本地构建二进制文件
go build ./cmd/claude-config

# 运行工具
./claude-config --help
```

### 系统要求
- Go 1.21或更高版本
- 访问`~/.claude`目录的权限（Claude Code配置目录）

## 使用方法

### 主要命令

```bash
# 安装资源（代理、命令、钩子、模板）
claude-config install

# 显示当前配置状态
claude-config status

# 配置代理设置（交互式）
claude-config proxy

# 管理验证系统
claude-config check

# 配置NTFY通知
claude-config notify

# 配置AI提供商集成
claude-config ai

# 备份和恢复配置
claude-config backup
```

**注意**：如果你是从源码构建而不是使用`make install`，请在命令前加上`./`（例如，`./claude-config status`）。

### 命令示例

```bash
# 安装所有可用资源到~/.claude
claude-config install
# 安装：代理、命令、钩子、输出样式、设置

# 检查所有配置的当前状态
claude-config status

# 交互式代理配置并进行验证
claude-config proxy
# 设置HTTP_PROXY和HTTPS_PROXY环境变量
# 验证代理连接性

# 配置AI提供商（DeepSeek、Kimi、GLM4.5、Doubao）
claude-config ai on deepseek
# 交互式设置和连接测试

# 启用验证系统
claude-config check on
# 配置验证，支持特定语言的代码检查和测试
```

## 命令参考

### 主要命令

#### `status` - 显示配置状态
```bash
claude-config status
```
显示所有配置的综合状态，包括代理、AI提供商、验证系统和通知。

#### `proxy` - 代理管理
```bash
# 交互式代理配置
claude-config proxy on

# 禁用代理
claude-config proxy off

# 切换代理状态
claude-config proxy toggle
```
管理HTTP/HTTPS代理设置并进行验证和连接测试。

#### `ai` - AI提供商管理
```bash
# 启用特定AI提供商（如需要会提示输入API密钥）
claude-config ai on deepseek
claude-config ai on kimi
claude-config ai on zhipu
claude-config ai on doubao

# 禁用所有AI提供商
claude-config ai off

# 重置特定提供商（删除API密钥）
claude-config ai reset deepseek

# 列出所有支持的提供商
claude-config ai list

# 显示当前AI提供商状态
claude-config ai
```
支持多个AI提供商：DeepSeek、Kimi（Moonshot）、GLM4.5（ZhipuAI）和Doubao（字节跳动）。

#### `check` - 验证系统管理
```bash
# 启用验证系统
claude-config check on

# 禁用验证系统
claude-config check off
```
管理代码检查、测试和代码质量检查的开发验证。

#### `notify` - NTFY通知
```bash
# 启用NTFY通知
claude-config notify on

# 禁用NTFY通知
claude-config notify off
```
为开发工作流配置NTFY通知系统。

#### `install` - 资源安装
```bash
# 安装所有资源
claude-config install

# 使用强制标志安装（覆盖现有文件）
claude-config install --force
```
将代理、命令、验证钩子、输出样式和设置安装到`~/.claude`。

#### `backup` - 配置备份
```bash
claude-config backup
```
创建备份并恢复Claude Code配置。

## 项目结构

```
claude-config/
├── cmd/claude-config/          # CLI应用程序和命令实现
│   ├── main.go                # 应用程序入口点
│   ├── commands.go            # 命令结构和初始化
│   ├── status.go              # 状态命令实现
│   ├── proxy.go               # 代理管理命令
│   ├── check.go               # 验证系统管理
│   ├── aiprovider.go          # AI提供商集成
│   ├── notify.go              # NTFY通知设置
│   ├── install.go             # 资源安装命令
│   └── backup.go              # 备份和恢复功能
├── internal/                   # 私有包（Go internal约定）
│   ├── config/                # 配置文件管理
│   ├── proxy/                 # HTTP/HTTPS代理管理
│   ├── check/                 # 验证系统
│   ├── aiprovider/            # AI提供商客户端和配置
│   ├── file/                  # 文件操作和合并工具
│   ├── install/               # 资源安装和管理
│   └── claude/                # 核心接口和共享类型
└── resources/                  # 嵌入的资源和模板
    └── claude-config/         # 安装用的资源文件
        ├── agents/            # Claude Code代理定义
        ├── commands/          # 自定义Claude命令
        ├── hooks/             # Shell钩子脚本
        ├── output-styles/     # 输出格式化配置
        ├── settings.json      # 默认Claude设置
        └── CLAUDE.md.template # 项目配置模板
```

## 开发

### 构建和测试

```bash
# 构建应用程序
go build ./cmd/claude-config

# 运行所有测试
go test ./...

# 运行详细输出的测试
go test -v ./...

# 运行特定包的测试
go test ./internal/config
go test ./internal/proxy
go test ./internal/file

# 运行带竞争检测的测试
go test -race ./...

# 为不同平台构建
GOOS=linux GOARCH=amd64 go build ./cmd/claude-config
GOOS=darwin GOARCH=amd64 go build ./cmd/claude-config
GOOS=windows GOARCH=amd64 go build ./cmd/claude-config
```

### 代码质量

```bash
# 格式化代码
go fmt ./...

# 运行静态分析
go vet ./...

# 安装并运行golangci-lint
golangci-lint run
```

## 架构

### 管理器模式
应用程序使用基于管理器的架构，包含以下核心组件：

- **ConfigManager** (`internal/config`) - 处理Claude配置设置
- **ProxyManager** (`internal/proxy`) - 管理HTTP/HTTPS代理配置
- **CheckManager** (`internal/check`) - 控制验证系统
- **AIProviderManager** (`internal/aiprovider`) - 管理多提供商AI API集成

所有管理器都在`main.go:init()`中初始化，并在`~/.claude`目录上运行。

### 资源系统
资源系统（`internal/install`）提供：
- 使用Go embed嵌入资源文件
- 模板处理和自定义
- 带备份的原子文件操作
- 具有冲突解决的配置合并

### 配置目录
所有操作都以`~/.claude`作为基础配置目录：

```
~/.claude/
├── settings.json              # 主要Claude设置
├── claude_config.toml         # 工具特定配置
├── agents/                    # 自定义代理定义
├── commands/                  # 自定义命令
├── hooks/                     # 开发验证钩子
└── output-styles/             # 输出格式样式
```

## 贡献

### 开发指南
- 遵循标准Go项目结构
- 使用有意义的包和函数名
- 为所有新功能编写测试
- 在提交PR之前确保所有测试通过
- 使用Go modules进行依赖管理

### 添加新命令
1. 在`cmd/claude-config/`中创建命令实现
2. 在`commands.go`的`initCommands()`中添加命令
3. 如需要，在`internal/`中创建对应的管理器
4. 为新功能添加测试
5. 更新文档

## 许可证

本项目采用Apache License 2.0许可证 - 详情请参阅[LICENSE](LICENSE)文件。

---

**注意**：此工具管理你在`~/.claude`中的Claude Code配置。在进行更改之前，请务必备份你的配置。
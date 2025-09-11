# claude-config

Claude Code 统一配置管理工具，提供配置管理、代理设置、钩子系统控制、NTFY 通知、DeepSeek API 集成和文件操作功能。

## 功能特性

- **配置管理** - 管理 Claude Code 设置和配置
- **代理设置** - 配置 HTTP/HTTPS 代理设置
- **钩子系统** - 控制验证钩子和自动化检查
- **DeepSeek API 集成** - 管理 DeepSeek API 配置
- **NTFY 通知** - 配置通知系统
- **文件操作** - 处理文件合并和操作
- **资源管理** - 安装和管理代理、命令和模板

## 安装

### Go 安装
```bash
go build ./cmd/claude-config
```

### 传统 Python 支持
```bash
make install
```

## 使用方法

### 主要命令

```bash
# 显示当前配置状态
./claude-config status

# 配置代理设置
./claude-config proxy

# 管理钩子系统
./claude-config check

# 配置 DeepSeek API
./claude-config deepseek

# 管理通知
./claude-config notify

# 安装资源和配置
./claude-config install

# 备份和恢复配置
./claude-config backup
```

## 项目结构

```
cmd/claude-config/     # CLI 应用程序入口点
internal/              # 私有包
  ├── config/         # 配置管理
  ├── proxy/          # HTTP/HTTPS 代理管理  
  ├── check/          # 钩子系统验证
  ├── deepseek/       # DeepSeek API 集成
  ├── file/           # 文件操作和合并
  ├── install/        # 安装和资源管理
  └── claude/         # 核心接口和类型
resources/             # 模板文件、钩子、代理和配置
  ├── agents/         # Claude Code 代理定义
  ├── commands/       # 自定义 Claude 命令
  ├── hooks/          # Shell 钩子脚本
  ├── output-styles/  # 输出格式配置
  └── settings.json   # 默认 Claude 设置
```

## 开发

### 构建和测试
```bash
# 构建
go build ./cmd/claude-config

# 运行测试
go test ./...

# 运行特定包的测试
go test ./internal/config
go test ./internal/proxy
go test ./internal/file
```

### 传统 Python 命令
```bash
make test    # 运行 Python 测试
make clean   # 清理临时文件
```

## 配置目录

所有操作都以 `~/.claude` 目录作为基础配置位置。

## 许可证

[待添加许可证信息]
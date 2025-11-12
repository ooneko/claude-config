# claude-config

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=flat-square)](https://github.com/ooneko/claude-config)

**[简体中文](README.md) | [English](README_EN.md)**

<p align="center">
  <img src="logo.png" alt="claude-config logo" width="300"/>
</p>

<p align="center">
  <strong>现代化 Claude 配置管理工具</strong>
</p>

一个用 Go 编写的现代化、统一的 Claude Code 配置管理工具，让配置管理变得简单高效。

## ✨ 快速开始

只需 3 步，即可开始使用 claude-config：

```bash
# 1️⃣ 安装（最简单的方式）
go install github.com/ooneko/claude-config/cmd/claude-config@latest

# 2️⃣ 安装资源（一键配置所有组件）
claude-config install

# 3️⃣ 查看状态（确认配置完成）
claude-config status
```

🎉 **完成！** 你的 Claude Code 环境已经配置完毕。

## 📖 目录

- [功能特性](#-功能特性)
- [安装指南](#-安装指南)
- [使用示例](#-使用示例)
- [命令参考](#-命令参考)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

## 🚀 功能特性

### 核心功能
- 🎯 **配置管理** - 一键管理 Claude Code 设置和配置
- 🌐 **代理设置** - 智能配置 HTTP/HTTPS 代理并进行连接验证
- ✅ **验证系统** - 高级开发工作流验证和代码质量检查
- 🤖 **AI提供商集成** - 支持 DeepSeek、Kimi、GLM4.5、Doubao 等多家 AI
- 🔔 **NTFY通知** - 为开发工作流配置实时通知系统
- 📦 **资源管理** - 安装和管理代理、命令、钩子等开发资源
- 💾 **备份与恢复** - 完整的配置备份和一键恢复系统

### 为什么选择 claude-config？
- ⚡ **极简易用** - 一条命令完成所有配置
- 🔧 **智能管理** - 自动检测和解决配置冲突
- 🛡️ **安全可靠** - 原子操作，确保配置完整性
- 🌍 **跨平台** - 支持 Linux、macOS、Windows

## 📦 安装指南

### 🚀 方式一：直接安装

最简单快速的安装方式，无需克隆仓库：

```bash
# 一键安装（需要 Go 1.21+）
go install github.com/ooneko/claude-config/cmd/claude-config@latest
```

安装完成后，立即可用：
```bash
claude-config --help
claude-config status
```

> 💡 **PATH 提示**：确保 `~/go/bin` 在你的 PATH 中。如果没有，请添加：
> ```bash
> export PATH="$HOME/go/bin:$PATH"
> ```

### 🔧 方式二：本地构建

适合想要修改代码或本地开发的用户：

```bash
# 1. 克隆仓库
git clone https://github.com/ooneko/claude-config.git
cd claude-config

# 2. 安装到系统
make install
```

`make install` 会自动处理 PATH 配置。

### 📋 系统要求

- ✅ **Go 1.21** 或更高版本
- ✅ **权限**：访问 `~/.claude` 目录（Claude Code 配置目录）
- ✅ **系统**：Linux / macOS / Windows

---

## 💡 使用示例

### 🎯 基础工作流

```bash
# 1️⃣ 首次使用：安装所有资源
claude-config install

# 2️⃣ 检查当前配置状态
claude-config status

# 3️⃣ 配置代理（如果需要）
claude-config proxy on

# 4️⃣ 配置 AI 提供商（例如 DeepSeek）
claude-config ai on deepseek

# 5️⃣ 启动 Claude Code （glm 模型）
claude-config start glm
```

### 🌐 代理配置示例

```bash
# 交互式配置代理
claude-config proxy

# 快速开启代理
claude-config proxy on

# 临时切换代理状态
claude-config proxy toggle

# 关闭代理
claude-config proxy off
```

### 🤖 AI 提供商配置示例

```bash
# 配置 DeepSeek
claude-config ai on deepseek

# 配置 Kimi（月之暗面）
claude-config ai on kimi

# 配置智谱 GLM
claude-config ai on glm

# 配置豆包（字节跳动）
claude-config ai on doubao

# 查看当前 AI 配置
claude-config ai

# 重置特定提供商
claude-config ai reset deepseek
```

### 🚀 启动 Claude Code 示例

```bash
# 启动原生 Claude Code（清理所有 AI 配置）
claude-config start

# 使用已配置的 DeepSeek 启动
claude-config start deepseek

# 使用已配置的 Kimi 启动，指定模型
claude-config start kimi --model kimi-plus

# 使用 GLM 启动，临时指定 API 密钥
claude-config start glm --api-key sk-xxxxxxxx

# 使用豆包启动，指定模型和 API 密钥
claude-config start doubao --model doubao-pro --api-key your-api-key
```

## 📚 命令参考

### 🎯 核心命令一览

| 命令 | 功能 | 快速示例 |
|------|------|----------|
| `install` | 安装所有资源 | `claude-config install` |
| `status` | 查看配置状态 | `claude-config status` |
| `proxy` | 代理配置管理 | `claude-config proxy on` |
| `ai` | AI提供商配置 | `claude-config ai on deepseek` |
| `check` | 验证系统控制 | `claude-config check on` |
| `notify` | 通知系统配置 | `claude-config notify on` |
| `start` | 启动Claude Code | `claude-config start` |
| `backup` | 备份恢复配置 | `claude-config backup` |

### 📋 详细命令说明

#### `claude-config install` - 资源安装
一键安装所有开发资源到 `~/.claude`：
```bash
# 安装所有资源（代理、命令、钩子、模板等）
claude-config install

# 强制覆盖安装（慎用）
claude-config install --force
```

#### `claude-config status` - 配置状态
查看当前所有配置的状态：
```bash
claude-config status
```
输出示例：
```
✅ 配置文件: 已就绪
🤖 AI提供商: DeepSeek (已连接)
🌐 代理配置: 未启用
✅ 验证系统: 已启用
🔔 通知系统: 已启用
```

#### `claude-config proxy` - 代理管理
智能代理配置和验证：
```bash
# 交互式配置
claude-config proxy

# 快速启用
claude-config proxy on

# 切换状态
claude-config proxy toggle

# 完全禁用
claude-config proxy off
```

#### `claude-config ai` - AI提供商管理
支持多家 AI 服务商：
```bash
# 启用特定提供商（会提示输入API密钥）
claude-config ai on deepseek    # DeepSeek AI
claude-config ai on kimi        # Kimi (月之暗面)
claude-config ai on glm         # 智谱 GLM
claude-config ai on doubao      # 豆包 (字节跳动)

# 查看所有支持的提供商
claude-config ai list

# 查看当前配置
claude-config ai

# 禁用所有AI提供商
claude-config ai off

# 重置特定提供商（删除密钥）
claude-config ai reset deepseek
```

#### `claude-config check` - 验证系统
控制代码质量检查：
```bash
# 启用开发验证
claude-config check on

# 禁用验证系统
claude-config check off
```

#### `claude-config notify` - 通知系统
配置 NTFY 实时通知：
```bash
# 启用通知
claude-config notify on

# 禁用通知
claude-config notify off
```

#### `claude-config start` - 启动 Claude Code
智能启动 Claude Code，支持多种模式：
```bash
# 启动原生 Claude Code（清理所有 AI 配置）
claude-config start

# 使用已配置的 AI 提供商启动
claude-config start deepseek    # 使用 DeepSeek
claude-config start kimi        # 使用 Kimi
claude-config start glm         # 使用 GLM
claude-config start doubao      # 使用豆包

# 高级选项（临时覆盖配置）
claude-config start kimi --model kimi-plus              # 指定模型
claude-config start glm --api-key sk-xxxxxxxx           # 临时 API 密钥
claude-config start doubao --model pro --key your-key   # 同时指定模型和密钥
```

**特性：**
- 🔄 **智能切换** - 无参数时启动原生 Claude，有参数时使用指定 AI
- 🔐 **密钥管理** - 优先使用存储的密钥，支持临时覆盖
- 🎯 **模型选择** - 支持临时指定不同模型
- 🧹 **配置清理** - 启动原生版本时自动清理现有配置

#### `claude-config backup` - 配置备份
安全备份和恢复：
```bash
# 创建配置备份
claude-config backup

# 查看恢复选项
claude-config backup --help
```

> 💡 **提示**：如果是从源码构建，请在命令前加上 `./`（例如 `./claude-config status`）



## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是 bug 报告、功能建议还是代码贡献。

### 🚀 快速开始

```bash
# 1. Fork 并克隆仓库
git clone https://github.com/your-username/claude-config.git
cd claude-config

# 2. 创建功能分支
git checkout -b feature/amazing-feature

# 3. 开发和测试
make dev  # 运行完整的开发工作流

# 4. 提交更改
git commit -m "feat: add amazing feature"

# 5. 推送并创建 PR
git push origin feature/amazing-feature
```

### 📋 开发规范

#### 代码质量
- ✅ **遵循 Go 项目结构** - 使用标准的 Go 项目布局
- ✅ **有意义的命名** - 函数和变量名要清晰表达意图
- ✅ **编写测试** - 新功能必须有对应的测试用例
- ✅ **通过检查** - 提交前运行 `make check` 确保代码质量

#### 提交规范
- 🎯 **清晰的信息** - 提交信息要说明更改的目的
- 🔍 **原子提交** - 一次提交只做一件事
- 📝 **更新文档** - 重大更改要更新相关文档

#### 添加新功能
1. **实现命令** - 在 `cmd/claude-config/` 中创建命令文件
2. **注册命令** - 在 `commands.go` 的 `initCommands()` 中注册
3. **创建管理器** - 需要时在 `internal/` 中创建对应管理器
4. **编写测试** - 为新功能添加全面的测试
5. **更新文档** - 更新 README 和相关文档

### 🛠️ 开发工具

```bash
# 开发工作流（格式化、检查、构建）
make dev

# 运行测试
make test

# 代码覆盖率
make test-coverage

# 代码质量检查
make check

# 构建多平台二进制
make build-all
```

### 🐛 报告问题

发现 bug？请通过以下方式报告：

- 📋 **问题模板** - 使用 GitHub Issues 的 bug 模板
- 🔍 **详细信息** - 提供复现步骤和环境信息
- 📸 **截图** - 如可能，提供相关截图
- 💻 **日志** - 附上相关的错误日志

---

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

### 📋 许可证摘要

- ✅ **商业使用** - 可以用于商业项目
- ✅ **修改** - 可以修改源代码
- ✅ **分发** - 可以分发原版或修改版
- ✅ **私用** - 可以私人使用而不开源
- ⚠️ **责任** - 需要保留原作者的版权声明
- ⚠️ **专利** - 提供专利授权

---

## ⚠️ 重要提示

此工具管理你在 `~/.claude` 中的 Claude Code 配置。在进行重大更改之前，**强烈建议备份你的配置**：

```bash
# 创建配置备份
claude-config backup
```

## 🙏 致谢

感谢所有为 claude-config 项目做出贡献的开发者和用户！

---

<div align="center">

**[⬆️ 回到顶部](#claude-config)**

Made with ❤️ by the claude-config community

</div>
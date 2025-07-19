# Claude 配置文件仓库

这是一个个人的 Claude Code 配置仓库，包含了自定义的工作流程、命令和钩子脚本。

## 📁 项目结构

```
.claude/
├── CLAUDE.md              # 全局指令和开发规范
├── settings.json          # Claude Code 配置文件
├── claude-config.sh       # Shell 配置脚本
├── commands/              # 自定义命令
│   ├── check.md          # 检查代码质量
│   ├── next.md           # 规划下一步任务
│   ├── prompt.md         # 生成项目特定提示
│   └── ultrathink.md     # 深度思考模式
└── hooks/                 # 自动化钩子脚本
    ├── smart-lint.sh     # 智能代码检查
    ├── smart-test.sh     # 智能测试运行
    └── ...               # 其他工具脚本
```

## 🚀 快速开始

1. **克隆仓库**
   ```bash
   cd ~
   git clone https://github.com/ooneko/claude-config.git .claude
   ```

2. **配置 Shell 环境**
   ```bash
   # 添加到你的 shell 配置文件 (.bashrc 或 .zshrc)
   source ~/.claude/claude-config.sh
   ```

3. **启动 Claude Code**
   ```bash
   claude
   ```

## 🛠️ 主要功能

### 自定义命令

- `/check` - 运行代码质量检查
- `/next` - 规划下一步开发任务
- `/prompt` - 生成项目特定的 CLAUDE.md
- `/ultrathink` - 激活深度思考模式

### 自动化钩子

配置文件中的钩子会在不同的操作阶段自动运行：

- **代码编辑后**: 自动运行格式化和 lint 检查
- **测试触发时**: 智能选择并运行相关测试
- **提交前检查**: 确保代码质量符合标准

### 开发规范

`CLAUDE.md` 文件定义了详细的开发规范，包括：

- ✅ 强制的代码质量检查
- 🔄 Git 工作流程最佳实践
- 🚫 Go 语言特定的禁用模式
- 📝 清晰的进度跟踪要求

## ⚙️ 配置说明

### settings.json

主要配置项：

- `theme`: 主题设置（dark/light）
- `anonymousUsageTracking`: 匿名使用统计
- `defaultCommandCategories`: 默认启用的命令类别
- `memory`: 内存管理设置
- `hooks`: 自动化钩子配置

### 环境变量

通过 `claude-config.sh` 设置的环境变量：

- `CLAUDE_MODEL_OVERRIDE`: 模型选择覆盖
- `DEFAULT_MODEL_ALIAS`: 默认模型别名
- Shell 别名和辅助函数

## 📋 钩子脚本

### smart-lint.sh

智能检测项目类型并运行相应的 lint 工具：
- Go 项目: `golangci-lint`
- Node.js 项目: `npm run lint`
- Python 项目: `ruff check`

### smart-test.sh

根据修改的文件智能运行相关测试：
- 自动检测测试框架
- 只运行受影响的测试
- 支持多种语言和框架

## 🤝 贡献

这是一个个人配置仓库，但欢迎提出建议或分享你的配置想法。

## 📄 许可证

MIT License

## 👤 作者

ooneko <binhong.hua@gmail.com>
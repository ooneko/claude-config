# Claude 配置管理系统

这是一个专业的 Claude Code 配置仓库，集成了智能钩子、专业代理、自定义命令和动态配置管理，为高效开发工作流程提供完整的自动化解决方案。

## 📁 项目结构

```
claude-config/
├── CLAUDE.md                   # 全局开发规范和工作流程指令
├── settings.json               # Claude Code 核心配置文件
├── claude-config.sh            # 动态配置管理脚本
├── commands/                   # 专业自定义命令 (6个)
│   ├── archviz.md             # 交互式架构可视化生成器
│   ├── check.md               # 代码质量检查
│   ├── commit.md              # 智能提交管理
│   ├── next.md                # 任务规划助手
│   ├── prompt.md              # 项目特定提示生成器
│   └── ultrathink.md          # 深度思考模式
├── agents/                     # 专业智能代理 (7个)
│   ├── ai-engineer.md         # AI系统架构专家
│   ├── backend-developer.md   # 后端开发专家
│   ├── code-reviewer.md       # 代码审查专家
│   ├── frontend-developer.md  # 前端开发专家
│   ├── golang-pro.md          # Go语言专业开发
│   ├── product-manager.md     # 产品管理专家
│   └── vue-expert.md          # Vue.js生态专家
├── hooks/                      # 智能自动化钩子脚本
│   ├── smart-lint.sh          # 智能代码检查
│   ├── smart-test.sh          # 智能测试运行
│   ├── common-helpers.sh      # 通用辅助函数
│   ├── ntfy-notifier.sh       # 任务完成通知
│   └── [语言特定钩子]          # Go、Tilt等专用钩子
└── output-styles/              # 专业输出样式
    └── design.md              # 系统设计输出规范
```

## 🚀 快速开始

1. **克隆配置仓库**
   ```bash
   cd ~
   git clone https://github.com/your-username/claude-config.git .claude
   ```

2. **配置 Shell 环境**
   ```bash
   # 为 claude-config.sh 创建别名，添加到你的 shell 配置文件 (.bashrc 或 .zshrc)
   echo 'alias claude-config="$HOME/.claude/claude-config.sh"' >> ~/.zshrc
   source ~/.zshrc
   ```
   
   > 这样就可以在任何地方直接运行 `claude-config` 命令了

3. **启动 Claude Code**
   ```bash
   claude
   ```

## 🎛️ 动态配置管理

### claude-config.sh 配置工具

强大的配置管理脚本，支持动态切换：

```bash
# 查看当前配置状态
claude-config status

# 代理管理
claude-config proxy          # 切换代理 (开/关)
claude-config proxy on       # 启用代理
claude-config proxy off      # 禁用代理

# Hooks 管理
claude-config hooks on       # 启用智能钩子
claude-config hooks off      # 禁用钩子

# DeepSeek 配置管理
claude-config deepseek on    # 启用 DeepSeek API
claude-config deepseek off   # 禁用 DeepSeek
claude-config deepseek reset # 清除 API 密钥

# 配置备份
claude-config backup         # 备份当前配置
```

## 🛠️ 专业功能模块

### 🎯 自定义命令系统

6个专业命令，涵盖开发全流程：

- `/archviz` - **架构可视化**: 生成交互式HTML架构图
- `/check` - **质量检查**: 运行代码质量检查
- `/commit` - **智能提交**: 自动生成规范的提交信息
- `/next` - **任务规划**: 规划下一步开发任务
- `/prompt` - **提示生成**: 生成项目特定的 CLAUDE.md
- `/ultrathink` - **深度思考**: 激活最大推理能力

### 🤖 智能代理矩阵

7个专业代理，按技术栈和角色分工：

- **ai-engineer**: AI系统设计、模型实现、生产部署
- **backend-developer**: 可扩展API开发、微服务架构
- **code-reviewer**: 实用代码审查、可操作反馈
- **frontend-developer**: React专家、用户体验优化
- **golang-pro**: Go语言专家、并发编程、云原生
- **product-manager**: 产品策略、用户中心开发
- **vue-expert**: Vue 3 Composition API、Nuxt 3专家

### ⚡ 智能钩子系统

自动化钩子在关键时刻自动运行：

- **PostToolUse 钩子**: 代码编辑后自动运行
  - `smart-lint.sh`: 智能检测项目类型并运行相应 linter
  - `smart-test.sh`: 根据修改文件智能运行相关测试

- **Stop 钩子**: 会话结束时运行
  - `ntfy-notifier.sh`: 发送任务完成通知

#### smart-lint.sh 智能检查

自动检测项目类型并运行相应工具：
- **Go项目**: `golangci-lint run`
- **Node.js项目**: `npm run lint` 或 `eslint`
- **Python项目**: `ruff check` 或 `flake8`
- **Tilt项目**: `tilt verify`

#### smart-test.sh 智能测试

根据修改文件智能运行测试：
- 自动检测测试框架 (Jest, Pytest, Go test等)
- 只运行受影响的测试文件
- 支持多种语言和测试环境

## 📋 核心配置详解

### settings.json 配置结构

```json
{
  "env": {
    "http_proxy": "http://127.0.0.1:7890",
    "https_proxy": "http://127.0.0.1:7890",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxx",
    "ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic"
  },
  "hooks": {
    "PostToolUse": [...],
    "Stop": [...]
  },
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.js"
  }
}
```

### CLAUDE.md 开发规范

完整的开发伙伴关系规范，包括：

- **🚨 强制自动化检查**: 所有钩子问题都是阻塞性的
- **🔄 Git分支工作流**: 为每个任务创建新分支
- **📊 研究→规划→实现**: 永不直接跳到编码阶段
- **🤖 多代理使用**: 积极利用子代理并行处理
- **✅ 现实检查点**: 在关键时刻停下来验证
- **🚫 Go语言禁用模式**: 严格禁止 interface{}、time.Sleep等

## 🔧 环境集成

### 代理配置
- HTTP/HTTPS代理支持 (127.0.0.1:7890)
- 自动环境变量设置
- 一键开关切换

### DeepSeek API 集成
- 安全的 API 密钥管理
- 自动模型配置
- 透明的 Anthropic 兼容性

### 状态行集成
- 实时配置状态显示  
- 自定义 JavaScript 状态行
- 零填充布局优化

## 🎨 输出样式系统

### design.md 专业设计输出

专门针对系统架构和API设计的输出规范：
- **实用设计方法**: 问题驱动的设计过程
- **行业最佳实践**: SOLID、DRY、KISS原则集成
- **多格式思维**: 图表、规范、代码结构
- **验证焦点**: 可扩展性、可维护性、性能、安全

## 📊 使用统计

- **总命令数**: 6个专业命令
- **智能代理数**: 7个专业代理  
- **钩子脚本数**: 9个自动化脚本
- **支持语言**: Go, JavaScript, Python, Vue, React
- **集成工具**: golangci-lint, eslint, ruff, tilt

## 🤝 贡献

这是一个个人配置仓库，但欢迎提出建议或分享你的配置想法。

## 📄 许可证

MIT License

## 👤 作者

ooneko <binhong.hua@gmail.com>
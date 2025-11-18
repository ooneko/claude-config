---
allowed-tools: all
description: 通过结合 next.md 和您的参数来合成完整的提示词
---

## 🎯 提示词合成器

您将通过结合以下内容创建一个**完整的、可直接复制的提示词**：
1. 来自 ~/.claude/commands/next.md 的 next.md 命令模板
2. 此处提供的具体任务详情：$ARGUMENTS

### 📋 您的任务：

1. **读取** ~/.claude/commands/next.md 中的 next.md 命令文件
2. **提取** 核心提示词结构和要求
3. **整合** 用户的参数到提示词中，确保无缝衔接
4. **输出** 一个完整的、易于复制的提示词代码块

### 🎨 输出格式：

在 markdown 代码块中呈现合成的提示词，如下所示：

```
[结合了 next.md 指令和用户具体任务的完整合成提示词]
```

### ⚡ 合成规则：

1. **保持结构** - 维护来自 next.md 的工作流程、检查点和要求
2. **自然整合** - 用实际任务详情替换 `$ARGUMENTS` 占位符
3. **上下文感知** - 如果用户的参数引用了特定技术，强调相关部分
4. **完整且独立** - 输出在粘贴到新的 Claude 对话时应该完美运行
5. **无元注释** - 不要解释您在做什么，只需输出合成的提示词

### 🔧 增强指南：

- 如果任务提及特定语言（Go、Python 等），强调那些特定语言的规则
- 如果任务看起来复杂，确保突出 "ultrathink" 和 "multiple agents" 部分
- 如果任务涉及重构，强调 "delete old code" 要求
- 无论任务如何，都要保留所有关键要求（hooks、linting、testing）

### 📦 示例行为：

如果用户提供："implement a REST API for user management with JWT authentication"

您应该：
1. 读取 next.md
2. 用用户的任务替换 $ARGUMENTS
3. 强调相关部分（API 设计、安全性、测试）
4. 输出完整的、集成的提示词

**立即开始合成** - 读取 next.md 并创建完美的提示词！

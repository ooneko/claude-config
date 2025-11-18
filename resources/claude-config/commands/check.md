---
allowed-tools: all
description: 验证代码质量并修复所有问题
---

# 代码质量检查

修复质量验证过程中发现的所有问题。不要只是报告问题。

## 工作流程

1. **识别** - 运行所有验证命令
2. **修复** - 处理发现的每个问题
3. **验证** - 重新运行直到所有检查通过

## 验证命令

查找并运行所有适用的命令：
- **代码检查**: `make lint`, `golangci-lint run`, `npm run lint`, `ruff check`
- **测试**: `make test`, `go test ./...`, `npm test`, `pytest`
- **构建**: `make build`, `go build ./...`, `npm run build`
- **格式化**: `gofmt`, `prettier`, `black`
- **安全**: `gosec`, `npm audit`, `bandit`

## 并行修复策略

当存在多个问题时，启动代理并行修复：
```
代理 1: 修复模块 A 中的代码检查问题
代理 2: 修复测试失败
代理 3: 修复类型错误
```

## Go 特定标准

- 使用具体类型，而不是 `interface{}`
- 为错误包装上下文
- 添加 godoc 注释
- 使用 channel 进行同步
- 不要使用 `time.Sleep()` 进行协调

## 成功标准

所有验证命令都通过，没有警告或错误。
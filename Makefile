.PHONY: install test clean help

# Python 解释器
PYTHON := python3

# 测试相关变量
TEST_DIR := tests
TEST_PATTERN := test_*.py

# 安装和配置
install:
	@echo "配置 Shell 环境..."
	@echo "检测当前 Shell..."
	@if [ "$$SHELL" = "/bin/zsh" ] || [ "$$SHELL" = "/usr/bin/zsh" ]; then \
		SHELL_RC=~/.zshrc; \
	elif [ "$$SHELL" = "/bin/bash" ] || [ "$$SHELL" = "/usr/bin/bash" ]; then \
		SHELL_RC=~/.bashrc; \
	else \
		echo "警告: 未识别的 Shell，默认使用 ~/.bashrc"; \
		SHELL_RC=~/.bashrc; \
	fi; \
	echo "为 claude-config.py 创建别名到 $$SHELL_RC..."; \
	if ! grep -q "alias claude-config=" "$$SHELL_RC" 2>/dev/null; then \
		echo 'alias claude-config="python3 $$HOME/.claude/claude-config.py"' >> "$$SHELL_RC"; \
		echo "✅ 别名已添加到 $$SHELL_RC"; \
	else \
		echo "✅ 别名已存在于 $$SHELL_RC"; \
	fi; \
	echo ""; \
	echo "🎉 安装完成！"; \
	echo "请运行以下命令使配置生效:"; \
	echo "  source $$SHELL_RC"; \
	echo ""; \
	echo "或者重新打开终端，然后就可以在任何地方使用 'claude-config' 命令了"

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  install  - 配置 Shell 环境和创建别名"
	@echo "  test     - 运行所有单元测试"
	@echo "  clean    - 清理临时文件"
	@echo "  help     - 显示此帮助信息"

# 运行测试
test:
	@echo "运行单元测试..."
	@$(PYTHON) -m unittest discover -s $(TEST_DIR) -p $(TEST_PATTERN) -v

# 清理临时文件
clean:
	@echo "清理临时文件..."
	@find . -type f -name "*.pyc" -delete
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@find . -type f -name "*.pyo" -delete
	@find . -type f -name "*~" -delete
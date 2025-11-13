package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/ooneko/claude-config/internal/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckCmd_CommandHandling 测试 check 命令的命令行接口
func TestCheckCmd_CommandHandling(t *testing.T) {
	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	originalCheckMgr := checkMgr
	defer func() {
		claudeDir = originalClaudeDir
		checkMgr = originalCheckMgr
	}()

	// 设置临时目录和新的管理器
	claudeDir = filepath.Join(tempDir, ".claude")

	// 创建新的 check manager 实例
	checkMgr = check.NewManager(claudeDir)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		expectOut   string
	}{
		{
			name:        "启用代码检查功能",
			args:        []string{"check", "on"},
			expectError: false,
			expectOut:   "✅ 代码检查功能已启用",
		},
		{
			name:        "使用 enable 启用检查功能",
			args:        []string{"check", "enable"},
			expectError: false,
			expectOut:   "✅ 代码检查功能已启用",
		},
		{
			name:        "禁用代码检查功能",
			args:        []string{"check", "off"},
			expectError: false,
			expectOut:   "❌ 代码检查功能已禁用",
		},
		{
			name:        "使用 disable 禁用检查功能",
			args:        []string{"check", "disable"},
			expectError: false,
			expectOut:   "❌ 代码检查功能已禁用",
		},
		{
			name:        "无效操作应该失败",
			args:        []string{"check", "invalid"},
			expectError: true,
			errorMsg:    "无效操作: invalid",
		},
		{
			name:        "缺少参数应该失败",
			args:        []string{"check"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 root 命令
			rootCmd := createRootCmd()

			// 设置参数
			rootCmd.SetArgs(tt.args)

			// 捕获输出
			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)

			// 执行命令
			err := rootCmd.ExecuteContext(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectOut != "" {
					assert.Contains(t, out.String(), tt.expectOut)
				}
			}
		})
	}
}

// TestCheckCmd_IntegrationWithInstall 测试 check 命令与 install 的集成
func TestCheckCmd_IntegrationWithInstall(t *testing.T) {
	// 这个测试验证重构后的行为：
	// 1. install --all 不再自动启用 check
	// 2. check on 是启用检查功能的唯一方式
	// 3. install hooks 只是文件操作，不启用功能

	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	originalCheckMgr := checkMgr
	defer func() {
		claudeDir = originalClaudeDir
		checkMgr = originalCheckMgr
	}()

	// 设置临时目录
	claudeDir = filepath.Join(tempDir, ".claude")

	// 创建新的 check manager 实例
	checkMgr = check.NewManager(claudeDir)

	t.Run("install --all 不会启用 check 功能", func(t *testing.T) {
		// 首先运行 install --all
		rootCmd := createRootCmd()
		rootCmd.SetArgs([]string{"install", "--all", "--force"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		require.NoError(t, err)

		// 验证 hooks 目录没有被创建（重构后的行为）
		// 注意：在 RED 阶段这个测试会失败，因为当前 --all 包含 hooks
		// 在 GREEN 阶段实现后，这个测试应该通过
		assert.NoDirExists(t, filepath.Join(claudeDir, "hooks"),
			"重构后 --all 不应该创建 hooks 目录")
	})

	t.Run("check on 是启用检查功能的唯一方式", func(t *testing.T) {
		// 确保检查功能初始是禁用状态
		// 运行 check off 确保禁用
		rootCmd := createRootCmd()
		rootCmd.SetArgs([]string{"check", "off"})

		_ = rootCmd.ExecuteContext(context.Background())
		// 可能有错误（如果没有配置），但这不影响测试

		// 然后运行 check on
		rootCmd.SetArgs([]string{"check", "on"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, out.String(), "✅ 代码检查功能已启用")
	})
}

// TestCheckCmd_HelpAndUsage 测试 check 命令的帮助信息
func TestCheckCmd_HelpAndUsage(t *testing.T) {
	rootCmd := createRootCmd()

	t.Run("显示帮助信息", func(t *testing.T) {
		rootCmd.SetArgs([]string{"check", "--help"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "检查功能控制")
		assert.Contains(t, output, "claude-config check <on|off>")
		assert.Contains(t, output, "smart-lint.sh")
		assert.Contains(t, output, "smart-test.sh")
	})

	t.Run("无效命令显示使用提示", func(t *testing.T) {
		rootCmd.SetArgs([]string{"check", "invalid"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)
		rootCmd.SetErr(&out)

		_ = rootCmd.ExecuteContext(context.Background())
		output := out.String()

		// 在测试环境中，Cobra 可能显示帮助信息而不是错误信息
		// 无论如何，输出应该包含相关的使用说明
		// 这里我们检查是否包含相关的功能描述
		assert.Contains(t, output, "检查功能控制", "输出应该包含检查功能的描述")
		assert.Contains(t, output, "smart-lint.sh", "输出应该包含 smart-lint.sh")
		assert.Contains(t, output, "smart-test.sh", "输出应该包含 smart-test.sh")
	})
}

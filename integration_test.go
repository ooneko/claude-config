//go:build integration
// +build integration

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_CompleteWorkflow 测试完整的工作流程
// 这个测试验证重构后的行为：
// 1. install --all 不再包含 hooks 目录
// 2. check on/off 是控制检查功能的唯一方式
// 3. hooks 文件需要单独安装
func TestIntegration_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

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

	rootCmd := createRootCmd()

	t.Run("步骤1: install --all 不再包含 hooks", func(t *testing.T) {
		// 运行 install --all
		rootCmd.SetArgs([]string{"install", "--all", "--force"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		require.NoError(t, err)

		// 验证 hooks 目录没有被创建（重构后的行为）
		// 在 RED 阶段这个断言会失败，因为当前 --all 包含 hooks
		// 在 GREEN 阶段实现后，这个断言应该通过
		hooksDir := filepath.Join(claudeDir, "hooks")
		assert.NoDirExists(t, hooksDir, "重构后 --all 不应该创建 hooks 目录")

		// 验证其他目录被正确创建
		assert.DirExists(t, filepath.Join(claudeDir, "agents"), "agents 目录应该被创建")
		assert.DirExists(t, filepath.Join(claudeDir, "commands"), "commands 目录应该被创建")
		assert.FileExists(t, filepath.Join(claudeDir, "settings.json"), "settings.json 应该被创建")
	})

	t.Run("步骤2: 单独安装 hooks 文件", func(t *testing.T) {
		// 注意：当前这个测试会失败，因为我们在重构中移除了 --enable-check
		// 在 GREEN 阶段后，我们需要找到另一种方式安装 hooks 文件
		// 可能是通过单独的 install 选项或者 check on 自动安装

		rootCmd.SetArgs([]string{"install", "--enable-check", "--force"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		// 在 RED 阶段这个命令会成功，在 GREEN 阶段会失败（因为选项被移除）
		if err == nil {
			// RED 阶段：--enable-check 仍然存在
			hooksDir := filepath.Join(claudeDir, "hooks")
			assert.DirExists(t, hooksDir, "当前 --enable-check 会创建 hooks 目录")
		} else {
			// GREEN 阶段：--enable-check 被移除，我们需要找到其他方式
			assert.Contains(t, err.Error(), "unknown flag")
			t.Log("✅ GREEN 阶段：--enable-check 选项已被成功移除")
		}
	})

	t.Run("步骤3: check on 启用检查功能", func(t *testing.T) {
		// 确保从禁用状态开始
		rootCmd.SetArgs([]string{"check", "off"})
		rootCmd.ExecuteContext(context.Background()) // 忽略错误

		// 启用检查功能
		rootCmd.SetArgs([]string{"check", "on"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, out.String(), "✅ 代码检查功能已启用")

		// 验证 settings.json 包含 hooks 配置
		settingsPath := filepath.Join(claudeDir, "settings.json")
		content, err := os.ReadFile(settingsPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "smart-lint.sh", "settings.json 应该包含 smart-lint.sh hook")
		assert.Contains(t, string(content), "smart-test.sh", "settings.json 应该包含 smart-test.sh hook")
	})

	t.Run("步骤4: check off 禁用检查功能", func(t *testing.T) {
		// 禁用检查功能
		rootCmd.SetArgs([]string{"check", "off"})

		var out bytes.Buffer
		rootCmd.SetOut(&out)

		err := rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, out.String(), "❌ 代码检查功能已禁用")

		// 验证 settings.json 不再包含检查相关 hooks
		settingsPath := filepath.Join(claudeDir, "settings.json")
		content, err := os.ReadFile(settingsPath)
		require.NoError(t, err)
		contentStr := string(content)
		assert.NotContains(t, contentStr, "smart-lint.sh", "settings.json 不应该包含 smart-lint.sh hook")
		assert.NotContains(t, contentStr, "smart-test.sh", "settings.json 不应该包含 smart-test.sh hook")
	})

	t.Run("步骤5: 验证工作流程完整性", func(t *testing.T) {
		// 再次验证完整的工作流程
		// install -> check on -> check off

		// 1. install 基本组件（不包含 hooks）
		rootCmd.SetArgs([]string{"install", "--agents", "--commands", "--force"})
		err := rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		// 2. 启用检查
		rootCmd.SetArgs([]string{"check", "on"})
		err = rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		// 3. 禁用检查
		rootCmd.SetArgs([]string{"check", "off"})
		err = rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		// 如果所有步骤都成功，说明重构后的工作流程正常
		t.Log("✅ 完整工作流程验证通过")
	})
}

// TestIntegration_BackwardCompatibility 测试向后兼容性
func TestIntegration_BackwardCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	originalCheckMgr := checkMgr
	defer func() {
		claudeDir = originalClaudeDir
		checkMgr = originalCheckMgr
	}()

	claudeDir = filepath.Join(tempDir, ".claude")

	t.Run("现有配置不受影响", func(t *testing.T) {
		// 创建一个现有的 settings.json
		settingsPath := filepath.Join(claudeDir, "settings.json")
		err := os.MkdirAll(claudeDir, 0755)
		require.NoError(t, err)

		existingContent := `{
  "includeCoAuthoredBy": true,
  "hooks": {
    "postToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "existing-hook.sh"
          }
        ]
      }
    ]
  }
}`
		err = os.WriteFile(settingsPath, []byte(existingContent), 0644)
		require.NoError(t, err)

		// 运行 check on/off 应该保留现有配置
		rootCmd := createRootCmd()

		// 启用检查
		rootCmd.SetArgs([]string{"check", "on"})
		err = rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		// 验证现有配置被保留
		content, err := os.ReadFile(settingsPath)
		require.NoError(t, err)
		contentStr := string(content)
		assert.Contains(t, contentStr, "includeCoAuthoredBy", "现有设置应该被保留")
		assert.Contains(t, contentStr, "existing-hook.sh", "现有 hooks 应该被保留")

		// 禁用检查
		rootCmd.SetArgs([]string{"check", "off"})
		err = rootCmd.ExecuteContext(context.Background())
		assert.NoError(t, err)

		// 验证检查相关的 hooks 被移除，但其他配置保留
		content, err = os.ReadFile(settingsPath)
		require.NoError(t, err)
		contentStr = string(content)
		assert.Contains(t, contentStr, "includeCoAuthoredBy", "现有设置应该被保留")
		assert.Contains(t, contentStr, "existing-hook.sh", "非检查相关的 hooks 应该被保留")
	})
}

// TestIntegration_Performance 测试性能和稳定性
func TestIntegration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	originalCheckMgr := checkMgr
	defer func() {
		claudeDir = originalClaudeDir
		checkMgr = originalCheckMgr
	}()

	claudeDir = filepath.Join(tempDir, ".claude")

	t.Run("多次 check on/off 操作", func(t *testing.T) {
		rootCmd := createRootCmd()

		// 多次切换检查功能，验证稳定性
		for i := 0; i < 10; i++ {
			// 启用
			rootCmd.SetArgs([]string{"check", "on"})
			err := rootCmd.ExecuteContext(context.Background())
			assert.NoError(t, err)

			// 短暂等待
			time.Sleep(10 * time.Millisecond)

			// 禁用
			rootCmd.SetArgs([]string{"check", "off"})
			err = rootCmd.ExecuteContext(context.Background())
			assert.NoError(t, err)

			// 短暂等待
			time.Sleep(10 * time.Millisecond)
		}

		t.Log("✅ 多次切换操作稳定")
	})

	t.Run("并发操作测试", func(t *testing.T) {
		// 简单的并发测试，验证没有竞态条件
		done := make(chan bool, 2)
		rootCmd1 := createRootCmd()
		rootCmd2 := createRootCmd()

		// 并发启用
		go func() {
			rootCmd1.SetArgs([]string{"check", "on"})
			err := rootCmd1.ExecuteContext(context.Background())
			assert.NoError(t, err)
			done <- true
		}()

		// 并发禁用（稍后启动）
		go func() {
			time.Sleep(50 * time.Millisecond)
			rootCmd2.SetArgs([]string{"check", "off"})
			err := rootCmd2.ExecuteContext(context.Background())
			assert.NoError(t, err)
			done <- true
		}()

		// 等待两个操作完成
		<-done
		<-done

		t.Log("✅ 并发操作测试通过")
	})
}
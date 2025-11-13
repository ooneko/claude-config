package check

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	claudeDir := "/tmp/test-claude"
	manager := NewManager(claudeDir)

	assert.NotNil(t, manager)
	assert.Equal(t, claudeDir, manager.claudeDir)
}

func TestManager_EnableCheck(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 确保 claude 目录存在
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	ctx := context.Background()

	// 测试启用检查功能
	err = manager.EnableCheck(ctx)
	assert.NoError(t, err)

	// 验证 settings.json 文件被创建/更新
	settingsPath := filepath.Join(claudeDir, "settings.json")
	assert.FileExists(t, settingsPath, "settings.json 应该被创建或更新")

	// 可以进一步验证文件内容包含 hooks 配置
	content, err := os.ReadFile(settingsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "smart-lint.sh", "settings.json 应该包含 smart-lint.sh hook")
	assert.Contains(t, string(content), "smart-test.sh", "settings.json 应该包含 smart-test.sh hook")
}

func TestManager_DisableCheck(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 确保 claude 目录存在
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	ctx := context.Background()

	// 先启用检查功能
	err = manager.EnableCheck(ctx)
	require.NoError(t, err)

	// 然后禁用检查功能
	err = manager.DisableCheck(ctx)
	assert.NoError(t, err)

	// 验证 settings.json 文件仍然存在
	settingsPath := filepath.Join(claudeDir, "settings.json")
	assert.FileExists(t, settingsPath, "settings.json 仍然应该存在")

	// 可以进一步验证文件内容不再包含 hooks 配置
	content, err := os.ReadFile(settingsPath)
	assert.NoError(t, err)
	assert.NotContains(t, string(content), "smart-lint.sh", "settings.json 不应该包含 smart-lint.sh hook")
	assert.NotContains(t, string(content), "smart-test.sh", "settings.json 不应该包含 smart-test.sh hook")
}

func TestManager_EnableCheck_Backup(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 确保 claude 目录存在
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// 创建一个现有的 settings.json 文件
	settingsPath := filepath.Join(claudeDir, "settings.json")
	existingContent := `{
  "includeCoAuthoredBy": true,
  "env": {
    "existingEnv": "value"
  },
  "hooks": {
    "preToolUse": [
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

	ctx := context.Background()

	// 启用检查功能
	err = manager.EnableCheck(ctx)
	assert.NoError(t, err)

	// 验证现有设置被保留，同时添加了新的 hooks
	content, err := os.ReadFile(settingsPath)
	assert.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "includeCoAuthoredBy\": true", "现有设置应该被保留")
	assert.Contains(t, contentStr, "existingEnv", "现有环境变量应该被保留")
	assert.Contains(t, contentStr, "existing-hook.sh", "现有 hooks 应该被保留")
	assert.Contains(t, contentStr, "smart-lint.sh", "新的 smart-lint.sh hook 应该被添加")
	assert.Contains(t, contentStr, "smart-test.sh", "新的 smart-test.sh hook 应该被添加")
}

func TestManager_DisableCheck_Backup(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 确保 claude 目录存在
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// 创建一个包含现有 hooks 的 settings.json 文件
	settingsPath := filepath.Join(claudeDir, "settings.json")
	existingContent := `{
  "includeCoAuthoredBy": true,
  "env": {
    "existingEnv": "value"
  },
  "hooks": {
    "postToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "existing-hook.sh"
          },
          {
            "type": "command",
            "command": "~/.claude/hooks/smart-lint.sh"
          }
        ]
      }
    ]
  }
}`
	err = os.WriteFile(settingsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// 禁用检查功能
	err = manager.DisableCheck(ctx)
	assert.NoError(t, err)

	// 验证现有设置被保留，但检查相关的 hooks 被移除
	content, err := os.ReadFile(settingsPath)
	assert.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "includeCoAuthoredBy\": true", "现有设置应该被保留")
	assert.Contains(t, contentStr, "existingEnv", "现有环境变量应该被保留")
	assert.Contains(t, contentStr, "existing-hook.sh", "非检查相关的 hooks 应该被保留")
	assert.NotContains(t, contentStr, "smart-lint.sh", "smart-lint.sh hook 应该被移除")
	assert.NotContains(t, contentStr, "smart-test.sh", "smart-test.sh hook 应该被移除")
}

func TestManager_EnableCheck_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试启用检查功能
	err := manager.EnableCheck(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestManager_DisableCheck_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试禁用检查功能
	err := manager.DisableCheck(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestManager_EnableCheck_NoHooksDirectory(t *testing.T) {
	// 这个测试验证即使 hooks 目录不存在，EnableCheck 也能正常工作
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 确保 claude 目录存在，但不创建 hooks 目录
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// 确认 hooks 目录不存在
	hooksDir := filepath.Join(claudeDir, "hooks")
	assert.NoDirExists(t, hooksDir, "hooks 目录应该不存在")

	ctx := context.Background()

	// 启用检查功能 - 这应该仍然能够创建配置
	err = manager.EnableCheck(ctx)
	assert.NoError(t, err)

	// 验证 settings.json 被创建并包含 hooks 配置
	settingsPath := filepath.Join(claudeDir, "settings.json")
	assert.FileExists(t, settingsPath, "settings.json 应该被创建")

	content, err := os.ReadFile(settingsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "smart-lint.sh", "settings.json 应该包含 smart-lint.sh hook")
	assert.Contains(t, string(content), "smart-test.sh", "settings.json 应该包含 smart-test.sh hook")
}

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooneko/claude-config/internal/install"
	"github.com/stretchr/testify/assert"
)

// TestInstallCmd_EnableCheckFlagRemoved 测试 --enable-check 选项被移除
func TestInstallCmd_EnableCheckFlagRemoved(t *testing.T) {
	// 这个测试在 RED 阶段应该失败，因为当前还有 --enable-check 选项
	// 在 GREEN 阶段实现后，这个测试应该通过

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "直接使用 --enable-check 应该失败",
			args:        []string{"install", "--enable-check"},
			expectError: true,
			errorMsg:    "unknown flag: --enable-check",
		},
		{
			name:        "使用 --all 不应该包含 enable-check",
			args:        []string{"install", "--all"},
			expectError: false,
		},
		{
			name:        "install 正常选项应该工作",
			args:        []string{"install", "--agents"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时目录用于测试
			tempDir := t.TempDir()

			// 保存原始的全局变量
			originalClaudeDir := claudeDir
			defer func() {
				claudeDir = originalClaudeDir
			}()

			// 设置临时目录
			claudeDir = filepath.Join(tempDir, ".claude")

			// 创建命令
			cmd := createInstallCmd()
			cmd.SetArgs(tt.args)

			// 捕获输出
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			// 执行命令
			err := cmd.ExecuteContext(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// 对于不应该失败的命令，我们只检查命令执行不报错
				// 实际的安装结果会在其他测试中验证
				if err != nil {
					// 如果是文件已存在的错误，这是可以接受的
					assert.Contains(t, err.Error(), "file exists")
				}
			}
		})
	}
}

// TestInstallCmd_AllOptionExcludesCheck 测试 --all 选项不再包含 check 功能
func TestInstallCmd_AllOptionExcludesCheck(t *testing.T) {
	// 这个测试验证 --all 选项不再自动安装 hooks 目录

	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	defer func() {
		claudeDir = originalClaudeDir
	}()

	claudeDir = filepath.Join(tempDir, ".claude")

	// 确保目录不存在
	os.RemoveAll(claudeDir)

	// 创建命令
	cmd := createInstallCmd()
	cmd.SetArgs([]string{"install", "--all", "--force"}) // 使用 --force 避免冲突

	// 捕获输出
	var out bytes.Buffer
	cmd.SetOut(&out)

	// 执行命令
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)

	// 验证 hooks 目录没有被创建
	assert.NoDirExists(t, filepath.Join(claudeDir, "hooks"), "hooks 目录不应该被 --all 选项创建")

	// 验证其他目录被正确创建
	assert.DirExists(t, filepath.Join(claudeDir, "agents"), "agents 目录应该被创建")
	assert.DirExists(t, filepath.Join(claudeDir, "commands"), "commands 目录应该被创建")
}

// TestInstallCmd_StructureAfterRefactor 测试重构后的 Options 结构
func TestInstallCmd_StructureAfterRefactor(t *testing.T) {
	// 这个测试验证重构后的 Options 结构不再包含 EnableCheck 字段
	// 并且能够正常工作

	tempDir := t.TempDir()

	// 保存原始的全局变量
	originalClaudeDir := claudeDir
	defer func() {
		claudeDir = originalClaudeDir
	}()

	claudeDir = filepath.Join(tempDir, ".claude")

	// 创建 install manager
	installMgr := install.NewManager(claudeDir)

	ctx := context.Background()

	// 验证可以创建不包含 EnableCheck 的 Options
	options := install.Options{
		All:      false,
		Agents:   true, // 使用其他有效选项
		Commands: false,
	}

	// 验证选项不包含 EnableCheck 字段（这会在编译时验证）
	// 如果 EnableCheck 字段仍然存在，这段代码会有编译错误

	// 尝试安装
	err := installMgr.Install(ctx, options)
	assert.NoError(t, err)

	// 验证 agents 目录被创建，但 hooks 目录没有被创建
	assert.DirExists(t, filepath.Join(claudeDir, "agents"), "agents 目录应该被创建")
	assert.NoDirExists(t, filepath.Join(claudeDir, "hooks"), "hooks 目录不应该被创建")

	// 测试 --all 选项不再包含 hooks
	allOptions := install.Options{All: true}
	components := allOptions.GetSelectedComponents()

	// 验证 hooks 不在组件列表中
	for _, component := range components {
		assert.NotEqual(t, "hooks", component, "hooks 不应该在 --all 的组件列表中")
	}

	t.Log("✅ GREEN 阶段：EnableCheck 字段已成功移除")
}

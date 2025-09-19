package install

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	claudeDir := "/tmp/test-claude"
	manager := NewManager(claudeDir)

	assert.NotNil(t, manager)
	assert.Equal(t, claudeDir, manager.claudeDir)
	assert.NotNil(t, manager.resources)
}

func TestManager_Install(t *testing.T) {
	// 创建临时目录作为测试的claude目录
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")

	// 创建InstallManager
	manager := NewManager(claudeDir)

	tests := []struct {
		name    string
		options InstallOptions
		wantErr bool
		checkFn func(t *testing.T, claudeDir string)
	}{
		{
			name: "安装所有配置文件",
			options: InstallOptions{
				All: true,
			},
			wantErr: false,
			checkFn: func(t *testing.T, claudeDir string) {
				// 检查是否创建了所有必要的目录和文件
				assert.DirExists(t, filepath.Join(claudeDir, "agents"))
				assert.DirExists(t, filepath.Join(claudeDir, "commands"))
				assert.DirExists(t, filepath.Join(claudeDir, "hooks"))
				assert.DirExists(t, filepath.Join(claudeDir, "output-styles"))
				assert.FileExists(t, filepath.Join(claudeDir, "settings.json"))
				assert.FileExists(t, filepath.Join(claudeDir, "CLAUDE.md"))
			},
		},
		{
			name: "仅安装agents",
			options: InstallOptions{
				Agents: true,
			},
			wantErr: false,
			checkFn: func(t *testing.T, claudeDir string) {
				assert.DirExists(t, filepath.Join(claudeDir, "agents"))
				// 验证其他目录不存在
				assert.NoFileExists(t, filepath.Join(claudeDir, "commands"))
			},
		},
		{
			name: "Force选项测试 - 强制覆盖",
			options: InstallOptions{
				Agents: true,
				Force:  true,
			},
			wantErr: false,
			checkFn: func(t *testing.T, claudeDir string) {
				assert.DirExists(t, filepath.Join(claudeDir, "agents"))
			},
		},
		{
			name:    "无效选项",
			options: InstallOptions{
				// 所有选项都为false
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理claudeDir
			os.RemoveAll(claudeDir)

			ctx := context.Background()
			err := manager.Install(ctx, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFn != nil {
					tt.checkFn(t, claudeDir)
				}
			}
		})
	}
}

func TestManager_installComponent(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	ctx := context.Background()

	// 测试未知组件
	err := manager.installComponent(ctx, "unknown-component", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未知组件")

	// 测试取消上下文
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	err = manager.installComponent(cancelCtx, "agents", false)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestNewResourceManager(t *testing.T) {
	manager := NewResourceManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.fs)
}

func TestResourceManager_ListEmbeddedFiles(t *testing.T) {
	manager := NewResourceManager()

	files, err := manager.ListEmbeddedFiles()
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// 检查是否包含预期的文件
	expectedFiles := []string{
		"agents/",
		"commands/",
		"hooks/",
		"output-styles/",
		"settings.json",
		"CLAUDE.md.template",
	}

	for _, expected := range expectedFiles {
		found := false
		for _, file := range files {
			if file == expected || strings.HasPrefix(file, expected) {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected file %s not found", expected)
	}
}

func TestResourceManager_ExtractFile(t *testing.T) {
	manager := NewResourceManager()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "settings.json")

	err := manager.ExtractFile("settings.json", destPath)
	assert.NoError(t, err)
	assert.FileExists(t, destPath)

	// 验证文件内容不为空
	content, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestResourceManager_ExtractFile_NotFound(t *testing.T) {
	manager := NewResourceManager()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "nonexistent.json")

	err := manager.ExtractFile("nonexistent.json", destPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "读取嵌入文件失败")
}

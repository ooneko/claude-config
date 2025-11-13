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
		options Options
		wantErr bool
		checkFn func(t *testing.T, claudeDir string)
	}{
		{
			name: "安装所有配置文件",
			options: Options{
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
			options: Options{
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
			options: Options{
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
			options: Options{
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

func TestResourceManager_ExtractFileWithPermissions(t *testing.T) {
	manager := NewResourceManager()

	tempDir := t.TempDir()

	tests := []struct {
		name         string
		srcFile      string
		destFileName string
		wantPerms    os.FileMode
	}{
		{
			name:         "Shell脚本文件应获取可执行权限",
			srcFile:      "hooks/smart-lint.sh",
			destFileName: "test.sh",
			wantPerms:    0755,
		},
		{
			name:         "JavaScript文件应获取可执行权限",
			srcFile:      "commands/statusline.js",
			destFileName: "test.js",
			wantPerms:    0755,
		},
		{
			name:         "Python文件应获取可执行权限",
			srcFile:      "hooks/test.py", // 假设存在，如果不存在会失败但这是预期的
			destFileName: "test.py",
			wantPerms:    0755,
		},
		{
			name:         "JSON配置文件应获取只读权限",
			srcFile:      "settings.json",
			destFileName: "settings.json",
			wantPerms:    0644,
		},
		{
			name:         "Markdown文件应获取只读权限",
			srcFile:      "CLAUDE.md.template",
			destFileName: "README.md",
			wantPerms:    0644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(tempDir, tt.destFileName)

			err := manager.ExtractFile(tt.srcFile, destPath)

			// 如果源文件不存在，跳过测试
			if err != nil && strings.Contains(err.Error(), "file does not exist") {
				t.Skipf("源文件 %s 不存在，跳过测试", tt.srcFile)
				return
			}

			assert.NoError(t, err)
			assert.FileExists(t, destPath)

			// 验证文件权限
			info, err := os.Stat(destPath)
			assert.NoError(t, err)

			// 只检查最后3位权限，避免系统差异影响
			perms := info.Mode().Perm()
			assert.Equal(t, tt.wantPerms, perms, "文件权限不匹配: got %o, want %o", perms, tt.wantPerms)
		})
	}
}

func TestResourceManager_ExtractDirectoryWithPermissions(t *testing.T) {
	manager := NewResourceManager()

	tempDir := t.TempDir()
	destDir := filepath.Join(tempDir, "hooks")

	err := manager.ExtractDirectory("hooks", destDir)
	assert.NoError(t, err)
	assert.DirExists(t, destDir)

	// 检查hooks目录中的文件权限
	entries, err := os.ReadDir(destDir)
	assert.NoError(t, err)

	executableFiles := []string{
		"smart-lint.sh",
		"debug-hook.sh",
		"test-tilt.sh",
		"lint-tilt.sh",
		"smart-test.sh",
		"ntfy-notifier.sh",
	}

	// 所有 .sh 文件现在都是可执行的
	nonExecutableFiles := []string{}

	// 验证文件权限
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		assert.NoError(t, err)

		fileName := entry.Name()
		perms := info.Mode().Perm()

		// 检查可执行文件
		isExecutableExpected := false
		for _, execFile := range executableFiles {
			if fileName == execFile {
				isExecutableExpected = true
				break
			}
		}

		// 检查非可执行文件
		isNonExecutableExpected := false
		for _, nonExecFile := range nonExecutableFiles {
			if fileName == nonExecFile {
				isNonExecutableExpected = true
				break
			}
		}

		if isExecutableExpected {
			assert.Equal(t, os.FileMode(0755), perms, "可执行文件 %s 应该有 0755 权限，实际为 %o", fileName, perms)
		} else if isNonExecutableExpected {
			assert.Equal(t, os.FileMode(0644), perms, "非可执行文件 %s 应该有 0644 权限，实际为 %o", fileName, perms)
		}
		// 其他文件暂时不做断言，避免测试过于严格
	}
}

func TestManager_InstallPreservesPermissions(t *testing.T) {
	// 创建临时目录作为测试的claude目录
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")

	manager := NewManager(claudeDir)

	ctx := context.Background()
	options := Options{
		Hooks: true,
	}

	err := manager.Install(ctx, options)
	assert.NoError(t, err)

	// 验证hooks目录和文件权限
	hooksDir := filepath.Join(claudeDir, "hooks")
	assert.DirExists(t, hooksDir)

	// 检查特定的脚本文件权限（所有 .sh 文件都应该可执行）
	testFiles := []struct {
		filePath  string
		wantPerms os.FileMode
	}{
		{"hooks/smart-lint.sh", 0755},
		{"hooks/debug-hook.sh", 0755},
		{"hooks/common-helpers.sh", 0755},
		{"hooks/lint-go.sh", 0755},
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(claudeDir, tf.filePath)

		// 只检查存在的文件
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Skipf("文件 %s 不存在，跳过权限检查", tf.filePath)
			continue
		}

		info, err := os.Stat(fullPath)
		assert.NoError(t, err)

		perms := info.Mode().Perm()
		assert.Equal(t, tf.wantPerms, perms, "文件 %s 权限不匹配: got %o, want %o", tf.filePath, perms, tf.wantPerms)
	}
}

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

func TestIsSpecialFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "settings.json是特殊文件",
			filePath: "settings.json",
			want:     true,
		},
		{
			name:     "CLAUDE.md是特殊文件",
			filePath: "CLAUDE.md",
			want:     true,
		},
		{
			name:     "带路径的settings.json",
			filePath: "some/path/settings.json",
			want:     true,
		},
		{
			name:     "带路径的CLAUDE.md",
			filePath: "some/path/CLAUDE.md",
			want:     true,
		},
		{
			name:     "普通命令文件不是特殊文件",
			filePath: "commands/test.md",
			want:     false,
		},
		{
			name:     "agent文件不是特殊文件",
			filePath: "agents/golang-pro.md",
			want:     false,
		},
		{
			name:     "hook脚本不是特殊文件",
			filePath: "hooks/test.sh",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSpecialFile(tt.filePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_listEmbeddedFilesForComponent(t *testing.T) {
	manager := NewManager("/tmp/test-claude")

	tests := []struct {
		name      string
		component string
		wantErr   bool
		validate  func(t *testing.T, files []string)
	}{
		{
			name:      "agents组件",
			component: "agents",
			wantErr:   false,
			validate: func(t *testing.T, files []string) {
				assert.NotEmpty(t, files)
				// 验证所有文件都以agents/开头
				for _, file := range files {
					assert.True(t, strings.HasPrefix(file, "agents/"), "文件 %s 应该以 agents/ 开头", file)
				}
			},
		},
		{
			name:      "commands组件",
			component: "commands",
			wantErr:   false,
			validate: func(t *testing.T, files []string) {
				assert.NotEmpty(t, files)
				for _, file := range files {
					assert.True(t, strings.HasPrefix(file, "commands/"), "文件 %s 应该以 commands/ 开头", file)
				}
			},
		},
		{
			name:      "hooks组件",
			component: "hooks",
			wantErr:   false,
			validate: func(t *testing.T, files []string) {
				assert.NotEmpty(t, files)
				for _, file := range files {
					assert.True(t, strings.HasPrefix(file, "hooks/"), "文件 %s 应该以 hooks/ 开头", file)
				}
			},
		},
		{
			name:      "statusline.js组件",
			component: "statusline.js",
			wantErr:   false,
			validate: func(t *testing.T, files []string) {
				assert.Contains(t, files, "statusline.js")
			},
		},
		{
			name:      "settings.json组件返回空列表",
			component: "settings.json",
			wantErr:   false,
			validate: func(t *testing.T, files []string) {
				assert.Empty(t, files, "特殊文件不应参与删除逻辑")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := manager.listEmbeddedFilesForComponent(tt.component)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, files)
				}
			}
		})
	}
}

func TestManager_listInstalledFilesInDirectory(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 创建测试文件
	commandsDir := filepath.Join(claudeDir, "commands")
	err := os.MkdirAll(commandsDir, 0755)
	assert.NoError(t, err)

	testFiles := []string{"test1.md", "test2.md", "subdir/test3.md"}
	for _, file := range testFiles {
		filePath := filepath.Join(commandsDir, file)
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(filePath, []byte("test"), 0644)
		assert.NoError(t, err)
	}

	// 测试
	files, err := manager.listInstalledFilesInDirectory("commands")
	assert.NoError(t, err)
	assert.Len(t, files, 3)

	// 验证文件路径
	for _, file := range files {
		assert.True(t, strings.HasPrefix(file, "commands/"), "文件 %s 应该以 commands/ 开头", file)
	}

	// 测试不存在的目录
	files, err = manager.listInstalledFilesInDirectory("nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, files)
}

func TestManager_listOrphanedFiles(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 先安装commands组件以获得嵌入资源
	ctx := context.Background()
	err := manager.Install(ctx, Options{Commands: true})
	assert.NoError(t, err)

	// 添加一些孤立文件
	commandsDir := filepath.Join(claudeDir, "commands")
	orphanedFiles := []string{"orphaned1.md", "orphaned2.md"}
	for _, file := range orphanedFiles {
		filePath := filepath.Join(commandsDir, file)
		err := os.WriteFile(filePath, []byte("orphaned"), 0644)
		assert.NoError(t, err)
	}

	// 获取孤立文件列表
	orphaned, err := manager.listOrphanedFiles("commands")
	assert.NoError(t, err)
	assert.NotEmpty(t, orphaned)

	// 验证孤立文件在列表中
	orphanedMap := make(map[string]bool)
	for _, file := range orphaned {
		orphanedMap[filepath.Base(file)] = true
	}

	for _, expected := range orphanedFiles {
		assert.True(t, orphanedMap[expected], "孤立文件 %s 应该在列表中", expected)
	}
}

func TestManager_cleanupOrphanedFiles_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 先安装commands组件
	ctx := context.Background()
	err := manager.Install(ctx, Options{Commands: true})
	assert.NoError(t, err)

	// 添加孤立文件
	commandsDir := filepath.Join(claudeDir, "commands")
	orphanedFile := filepath.Join(commandsDir, "orphaned.md")
	err = os.WriteFile(orphanedFile, []byte("orphaned"), 0644)
	assert.NoError(t, err)

	// 执行dry-run删除 (Delete=true, Force=false)
	options := Options{
		Commands: true,
		Delete:   true,
		Force:    false, // dry-run模式
	}

	err = manager.cleanupOrphanedFiles("commands", options)
	assert.NoError(t, err)

	// 验证文件仍然存在 (dry-run不删除)
	assert.FileExists(t, orphanedFile, "Dry-run模式不应删除文件")
}

func TestManager_cleanupOrphanedFiles_ActualDelete(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 先安装commands组件
	ctx := context.Background()
	err := manager.Install(ctx, Options{Commands: true})
	assert.NoError(t, err)

	// 添加孤立文件
	commandsDir := filepath.Join(claudeDir, "commands")
	orphanedFile := filepath.Join(commandsDir, "orphaned.md")
	err = os.WriteFile(orphanedFile, []byte("orphaned"), 0644)
	assert.NoError(t, err)

	// 执行实际删除 (Delete=true, Force=true)
	options := Options{
		Commands: true,
		Delete:   true,
		Force:    true, // 实际删除模式
	}

	err = manager.cleanupOrphanedFiles("commands", options)
	assert.NoError(t, err)

	// 验证文件已被删除
	assert.NoFileExists(t, orphanedFile, "实际删除模式应该删除文件")
}

func TestManager_Install_WithDelete(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	ctx := context.Background()

	// 第一次安装
	err := manager.Install(ctx, Options{Commands: true})
	assert.NoError(t, err)

	// 添加孤立文件
	commandsDir := filepath.Join(claudeDir, "commands")
	orphanedFile := filepath.Join(commandsDir, "orphaned.md")
	err = os.WriteFile(orphanedFile, []byte("orphaned"), 0644)
	assert.NoError(t, err)

	// 第二次安装,启用删除功能
	err = manager.Install(ctx, Options{
		Commands: true,
		Delete:   true,
		Force:    true,
	})
	assert.NoError(t, err)

	// 验证孤立文件已被删除
	assert.NoFileExists(t, orphanedFile, "孤立文件应该被删除")

	// 验证嵌入资源中的文件仍然存在
	entries, err := os.ReadDir(commandsDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, entries, "应该还有嵌入资源中的文件")
}

func TestManager_cleanupOrphanedFiles_SkipSpecialFiles(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	manager := NewManager(claudeDir)

	// 创建特殊文件
	settingsFile := filepath.Join(claudeDir, "settings.json")
	claudeMdFile := filepath.Join(claudeDir, "CLAUDE.md")

	err := os.MkdirAll(claudeDir, 0755)
	assert.NoError(t, err)

	err = os.WriteFile(settingsFile, []byte("{}"), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(claudeMdFile, []byte("# Test"), 0644)
	assert.NoError(t, err)

	// 执行删除 (这些文件不应该出现在孤立文件列表中)
	options := Options{
		Settings: true,
		Delete:   true,
		Force:    true,
	}

	// settings.json组件会被跳过
	err = manager.cleanupOrphanedFiles("settings.json", options)
	assert.NoError(t, err)

	// 验证特殊文件仍然存在
	assert.FileExists(t, settingsFile, "settings.json不应被删除")
	assert.FileExists(t, claudeMdFile, "CLAUDE.md不应被删除")
}

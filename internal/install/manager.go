package install

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ooneko/claude-config/resources"
)

// Manager install功能管理器
type Manager struct {
	claudeDir string
	resources *ResourceManager
}

// NewManager 创建新的install管理器
func NewManager(claudeDir string) *Manager {
	return &Manager{
		claudeDir: claudeDir,
		resources: NewResourceManager(),
	}
}

// Install 安装配置文件
func (m *Manager) Install(ctx context.Context, options Options) error {
	if err := options.Validate(); err != nil {
		return fmt.Errorf("无效的安装选项: %w", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(m.claudeDir, 0755); err != nil {
		return fmt.Errorf("创建Claude目录失败: %w", err)
	}

	components := options.GetSelectedComponents()

	for _, component := range components {
		if err := m.installComponent(ctx, component, options.Force); err != nil {
			return fmt.Errorf("安装组件%s失败: %w", component, err)
		}
	}

	return nil
}

// installComponent 安装单个组件
func (m *Manager) installComponent(ctx context.Context, component string, force bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	switch component {
	case "agents", "commands", "hooks", "output-styles":
		return m.installDirectory(component, force)
	case "settings.json":
		return m.installSettingsJSON()
	case "CLAUDE.md.template":
		return m.installClaudeMd(force)
	case "statusline.js":
		return m.installStatuslineJs(force)
	default:
		return fmt.Errorf("未知组件: %s", component)
	}
}

// installDirectory 安装目录 - 根据force参数决定是否覆盖现有目录
func (m *Manager) installDirectory(dirName string, force bool) error {
	targetDir := filepath.Join(m.claudeDir, dirName)

	// 如果不强制覆盖，检查目录是否存在
	if !force {
		if _, err := os.Stat(targetDir); err == nil {
			fmt.Printf("⚠️  目录 %s 已存在，跳过安装（使用 --force 强制覆盖）\n", dirName)
			return nil
		}
	}

	return m.resources.ExtractDirectory(dirName, targetDir)
}

// installSettingsJSON 安装settings.json - 始终使用智能合并
func (m *Manager) installSettingsJSON() error {
	targetPath := filepath.Join(m.claudeDir, "settings.json")

	// 创建临时文件来存储源文件内容
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "settings_source.json")

	if err := m.resources.ExtractFile("settings.json", tempFile); err != nil {
		return fmt.Errorf("提取源settings.json失败: %w", err)
	}
	defer os.Remove(tempFile) // 清理临时文件

	// 使用智能合并器合并文件
	merger := NewSettingsJSONMerger()
	return merger.MergeSettings(targetPath, tempFile)
}

// installClaudeMd 安装CLAUDE.md文件 - 总是覆盖现有文件
func (m *Manager) installClaudeMd(_ bool) error {
	targetPath := filepath.Join(m.claudeDir, "CLAUDE.md")
	// CLAUDE.md 默认总是覆盖，不受force参数影响
	return m.resources.ExtractFile("CLAUDE.md.template", targetPath)
}

// installStatuslineJs 安装statusline.js文件 - 根据force参数决定是否覆盖现有文件，并设置可执行权限
func (m *Manager) installStatuslineJs(force bool) error {
	targetPath := filepath.Join(m.claudeDir, "statusline.js")

	// 如果不强制覆盖，检查文件是否存在
	if !force {
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("⚠️  文件 statusline.js 已存在，跳过安装（使用 --force 强制覆盖）\n")
			return nil
		}
	}

	// 提取文件
	if err := m.resources.ExtractFile("statusline.js", targetPath); err != nil {
		return err
	}

	// 设置可执行权限 (0755)
	return os.Chmod(targetPath, 0755)
}

// ResourceManager embed资源管理器
type ResourceManager struct {
	fs embed.FS
}

// NewResourceManager 创建新的资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		fs: resources.EmbeddedFiles,
	}
}

// ListEmbeddedFiles 列出所有嵌入的文件
func (rm *ResourceManager) ListEmbeddedFiles() ([]string, error) {
	var files []string

	err := fs.WalkDir(rm.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		// 移除claude-config前缀
		if strings.HasPrefix(path, "claude-config/") {
			relativePath := path[len("claude-config/"):]
			if d.IsDir() {
				files = append(files, relativePath+"/")
			} else {
				files = append(files, relativePath)
			}
		}

		return nil
	})

	return files, err
}

// ExtractFile 提取单个文件
func (rm *ResourceManager) ExtractFile(srcPath, destPath string) error {
	fullSrcPath := filepath.Join("claude-config", srcPath)

	data, err := rm.fs.ReadFile(fullSrcPath)
	if err != nil {
		return fmt.Errorf("读取嵌入文件失败: %w", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	return os.WriteFile(destPath, data, GetFilePermissions(destPath))
}

// ExtractDirectory 提取目录
func (rm *ResourceManager) ExtractDirectory(srcDir, destDir string) error {
	fullSrcDir := filepath.Join("claude-config", srcDir)

	return fs.WalkDir(rm.fs, fullSrcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(fullSrcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := rm.fs.ReadFile(path)
		if err != nil {
			return err
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(destPath, data, GetFilePermissions(destPath))
	})
}

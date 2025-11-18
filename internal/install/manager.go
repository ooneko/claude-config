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

	// 第一阶段: 安装组件
	for _, component := range components {
		if err := m.installComponent(ctx, component, options.Force); err != nil {
			return fmt.Errorf("安装组件%s失败: %w", component, err)
		}
	}

	// 第二阶段: 清理孤立文件(如果启用了删除功能)
	if options.Delete {
		for _, component := range components {
			if err := m.cleanupOrphanedFiles(component, options); err != nil {
				return fmt.Errorf("清理组件%s的孤立文件失败: %w", component, err)
			}
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

// isSpecialFile 检查文件是否为特殊文件(不应被删除的文件)
func isSpecialFile(filePath string) bool {
	// 标准化路径分隔符
	normalizedPath := filepath.ToSlash(filePath)

	// settings.json 和 CLAUDE.md 永不删除
	specialFiles := []string{
		"settings.json",
		"CLAUDE.md",
	}

	for _, special := range specialFiles {
		if normalizedPath == special || strings.HasSuffix(normalizedPath, "/"+special) {
			return true
		}
	}

	return false
}

// listEmbeddedFilesForComponent 获取指定组件的嵌入资源文件列表
func (m *Manager) listEmbeddedFilesForComponent(component string) ([]string, error) {
	var files []string

	// 对于目录型组件,遍历嵌入资源中的对应目录
	if component == "agents" || component == "commands" || component == "hooks" || component == "output-styles" {
		fullSrcDir := filepath.Join("claude-config", component)

		err := fs.WalkDir(m.resources.fs, fullSrcDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// 跳过目录本身
			if d.IsDir() {
				return nil
			}

			// 计算相对路径
			relPath, err := filepath.Rel("claude-config", path)
			if err != nil {
				return err
			}

			files = append(files, relPath)
			return nil
		})

		return files, err
	}

	// 对于单文件组件
	switch component {
	case "statusline.js":
		files = append(files, "statusline.js")
	case "settings.json", "CLAUDE.md.template":
		// 这些特殊文件不参与删除逻辑
		return files, nil
	}

	return files, nil
}

// listInstalledFilesInDirectory 获取目标目录中已安装的文件列表
func (m *Manager) listInstalledFilesInDirectory(component string) ([]string, error) {
	var files []string

	targetDir := filepath.Join(m.claudeDir, component)

	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return files, nil // 目录不存在,返回空列表
	}

	// 遍历目录
	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 计算相对于 claudeDir 的路径
		relPath, err := filepath.Rel(m.claudeDir, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	return files, err
}

// listOrphanedFiles 获取孤立文件列表(在目标目录中存在但在嵌入资源中不存在的文件)
func (m *Manager) listOrphanedFiles(component string) ([]string, error) {
	// 获取嵌入资源文件列表
	embeddedFiles, err := m.listEmbeddedFilesForComponent(component)
	if err != nil {
		return nil, fmt.Errorf("获取嵌入资源文件列表失败: %w", err)
	}

	// 获取已安装文件列表
	installedFiles, err := m.listInstalledFilesInDirectory(component)
	if err != nil {
		return nil, fmt.Errorf("获取已安装文件列表失败: %w", err)
	}

	// 创建嵌入文件的映射,便于快速查找
	embeddedSet := make(map[string]bool)
	for _, file := range embeddedFiles {
		// 标准化路径
		normalizedPath := filepath.ToSlash(file)
		embeddedSet[normalizedPath] = true
	}

	// 找出孤立文件
	var orphanedFiles []string
	for _, installedFile := range installedFiles {
		normalizedPath := filepath.ToSlash(installedFile)

		// 跳过特殊文件
		if isSpecialFile(normalizedPath) {
			continue
		}

		// 如果不在嵌入资源中,则为孤立文件
		if !embeddedSet[normalizedPath] {
			orphanedFiles = append(orphanedFiles, installedFile)
		}
	}

	return orphanedFiles, nil
}

// deleteOrphanedFiles 删除孤立文件(或执行dry-run)
func (m *Manager) deleteOrphanedFiles(orphanedFiles []string, dryRun bool) (int, error) {
	count := 0

	for _, file := range orphanedFiles {
		fullPath := filepath.Join(m.claudeDir, file)

		if dryRun {
			// Dry-run模式: 只显示,不删除
			fmt.Printf("🗑️  %s\n", file)
		} else {
			// 实际删除
			if err := os.Remove(fullPath); err != nil {
				return count, fmt.Errorf("删除文件失败 %s: %w", file, err)
			}
			fmt.Printf("🗑️  已删除: %s\n", file)
		}
		count++
	}

	return count, nil
}

// cleanupOrphanedFiles 清理孤立文件的主入口
func (m *Manager) cleanupOrphanedFiles(component string, options Options) error {
	// 如果未启用删除功能,直接返回
	if !options.Delete {
		return nil
	}

	// 跳过特殊组件
	if component == "settings.json" || component == "CLAUDE.md.template" {
		return nil
	}

	// 获取孤立文件列表
	orphanedFiles, err := m.listOrphanedFiles(component)
	if err != nil {
		return err
	}

	// 如果没有孤立文件,直接返回
	if len(orphanedFiles) == 0 {
		return nil
	}

	// 确定是dry-run还是实际删除
	dryRun := !options.Force

	// 输出标题
	if dryRun {
		fmt.Printf("\n🔍 Dry-run 模式: 以下文件将被删除 (使用 --force 实际执行删除):\n\n")
	} else {
		fmt.Printf("\n⚠️  警告: 即将删除以下文件\n\n")
	}

	// 删除或显示文件
	count, err := m.deleteOrphanedFiles(orphanedFiles, dryRun)
	if err != nil {
		return err
	}

	// 输出汇总
	fmt.Println()
	if dryRun {
		fmt.Printf("📊 总计: %d 个文件将被删除\n", count)
		fmt.Println("\n💡 提示: 使用 --force 参数实际执行删除")
	} else {
		fmt.Printf("✅ 成功删除 %d 个孤立文件\n", count)
	}

	return nil
}

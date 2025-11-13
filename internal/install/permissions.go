package install

import (
	"os"
	"path/filepath"
	"strings"
)

// IsExecutableFile 检查文件是否应该设置为可执行
// 基于文件扩展名来判断，支持常见的脚本文件扩展名
func IsExecutableFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// 可执行文件的扩展名列表
	executableExts := map[string]bool{
		".sh":  true, // Shell脚本
		".js":  true, // JavaScript (Node.js脚本)
		".py":  true, // Python脚本
		".pl":  true, // Perl脚本
		".rb":  true, // Ruby脚本
		".php": true, // PHP脚本
		".bat": true, // Windows批处理
		".cmd": true, // Windows命令脚本
	}

	return executableExts[ext]
}

// GetFilePermissions 根据文件路径返回适当的权限
// 可执行文件返回 0755，普通文件返回 0644
func GetFilePermissions(filePath string) os.FileMode {
	if IsExecutableFile(filePath) {
		return 0755 // 可执行权限：rwxr-xr-x
	}
	return 0644 // 默认文件权限：rw-r--r--
}

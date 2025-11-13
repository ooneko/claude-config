package install

import (
	"os"
	"runtime"
	"testing"
	"testing/quick"
)

func TestIsExecutableFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"Shell脚本可执行", "hooks/test.sh", true},
		{"JavaScript文件可执行", "statusline.js", true},
		{"Python脚本可执行", "scripts/run.py", true},
		{"Perl脚本可执行", "scripts/run.pl", true},
		{"Ruby脚本可执行", "scripts/run.rb", true},
		{"PHP脚本可执行", "scripts/run.php", true},
		{"Windows批处理可执行", "scripts/run.bat", true},
		{"Windows命令脚本可执行", "scripts/run.cmd", true},
		{"配置文件不可执行", "settings.json", false},
		{"Markdown文件不可执行", "README.md", false},
		{"TOML配置文件不可执行", "config.toml", false},
		{"YAML配置文件不可执行", "config.yaml", false},
		{"文本文件不可执行", "notes.txt", false},
		{"隐藏配置文件不可执行", ".env", false},
		{"没有扩展名的文件不可执行", "Makefile", false},
		{"空路径不可执行", "", false},
		{"大写扩展名脚本可执行", "hooks/TEST.SH", true},
		{"混合大小写扩展名可执行", "scripts/Test.Py", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExecutableFile(tt.filePath)
			if got != tt.want {
				t.Errorf("IsExecutableFile(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestGetFilePermissions(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     os.FileMode
	}{
		{"Shell脚本获取可执行权限", "hooks/test.sh", 0755},
		{"JavaScript文件获取可执行权限", "statusline.js", 0755},
		{"Python脚本获取可执行权限", "scripts/run.py", 0755},
		{"Perl脚本获取可执行权限", "scripts/run.pl", 0755},
		{"Ruby脚本获取可执行权限", "scripts/run.rb", 0755},
		{"PHP脚本获取可执行权限", "scripts/run.php", 0755},
		{"Windows批处理获取可执行权限", "scripts/run.bat", 0755},
		{"Windows命令脚本获取可执行权限", "scripts/run.cmd", 0755},
		{"JSON配置文件获取只读权限", "settings.json", 0644},
		{"Markdown文件获取只读权限", "README.md", 0644},
		{"TOML配置文件获取只读权限", "config.toml", 0644},
		{"YAML配置文件获取只读权限", "config.yaml", 0644},
		{"文本文件获取只读权限", "notes.txt", 0644},
		{"隐藏配置文件获取只读权限", ".env", 0644},
		{"没有扩展名的文件获取只读权限", "Makefile", 0644},
		{"空路径获取只读权限", "", 0644},
		{"大写扩展名脚本获取可执行权限", "hooks/TEST.SH", 0755},
		{"混合大小写扩展名获取可执行权限", "scripts/Test.Py", 0755},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFilePermissions(tt.filePath)
			if got != tt.want {
				t.Errorf("GetFilePermissions(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestIsExecutableFilePropertyBased 使用属性测试验证扩展名检测的一致性
func TestIsExecutableFilePropertyBased(t *testing.T) {
	f := func(ext string, _ bool) bool {
		// 构造一个文件路径
		filePath := "test" + ext

		// 检查我们的实现
		result := IsExecutableFile(filePath)

		// 对于已知的可执行扩展名，结果应该一致
		executableExts := map[string]bool{
			".sh": true, ".js": true, ".py": true, ".pl": true, ".rb": true, ".php": true, ".bat": true, ".cmd": true,
		}

		if executableExts[ext] {
			return result == true
		} else if ext == "" || ext == ".txt" || ext == ".json" || ext == ".md" {
			return result == false
		}

		// 对于其他扩展名，只要结果一致就行
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error("属性测试失败:", err)
	}
}

// TestGetFilePermissionsQuickCheck 快速检查权限设置的一致性
func TestGetFilePermissionsQuickCheck(t *testing.T) {
	f := func(filePath string) bool {
		perms := GetFilePermissions(filePath)

		// 权限应该是 0644 或 0755
		if perms != 0644 && perms != 0755 {
			return false
		}

		// 如果文件是可执行的，权限应该是 0755
		if IsExecutableFile(filePath) {
			return perms == 0755
		}

		// 如果文件不可执行，权限应该是 0644
		return perms == 0644
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error("权限快速检查失败:", err)
	}
}

// TestGetFilePermissionsBoundaryConditions 测试边界条件
func TestGetFilePermissionsBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     os.FileMode
	}{
		{"空字符串", "", 0644},
		{"只有扩展名", ".sh", 0755},
		{"路径包含空格", "/path with spaces/test.sh", 0755},
		{"路径包含特殊字符", "/path/with-dashes/test.sh", 0755},
		{"路径包含下划线", "/path_with_underscores/test.sh", 0755},
		{"嵌套路径", "/very/deep/nested/path/test.sh", 0755},
		{"连续点号", "test..sh", 0755},
		{"多个扩展名", "test.backup.sh", 0755},
		{"大写路径", "/PATH/TEST.SH", 0755},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFilePermissions(tt.filePath)
			if got != tt.want {
				t.Errorf("GetFilePermissions(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestPermissionsConsistencyAcrossPlatforms 确保权限在不同平台上的一致性
func TestPermissionsConsistencyAcrossPlatforms(t *testing.T) {
	testFiles := []string{
		"test.sh", "test.js", "test.py", "test.json", "test.md", "test.txt",
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			perms := GetFilePermissions(file)

			// 在所有平台上，权限值应该相同
			if IsExecutableFile(file) {
				if perms != 0755 {
					t.Errorf("可执行文件 %s 在 %s 平台上权限应为 0755，实际为 %o",
						file, runtime.GOOS, perms)
				}
			} else {
				if perms != 0644 {
					t.Errorf("不可执行文件 %s 在 %s 平台上权限应为 0644，实际为 %o",
						file, runtime.GOOS, perms)
				}
			}
		})
	}
}

// BenchmarkIsExecutableFile 性能基准测试
func BenchmarkIsExecutableFile(b *testing.B) {
	testPaths := []string{
		"test.sh", "test.js", "test.py", "test.json", "test.md",
		"/very/long/path/to/executable/script.sh",
		"/another/path/with/many/components/test.js",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			IsExecutableFile(path)
		}
	}
}

// BenchmarkGetFilePermissions 性能基准测试
func BenchmarkGetFilePermissions(b *testing.B) {
	testPaths := []string{
		"test.sh", "test.js", "test.py", "test.json", "test.md",
		"/very/long/path/to/executable/script.sh",
		"/another/path/with/many/components/test.js",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			GetFilePermissions(path)
		}
	}
}

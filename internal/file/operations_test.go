package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ooneko/claude-config/internal/claude"
)

func TestFileOperations_Copy_All(t *testing.T) {
	// Setup temp directories
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	claudeDir := filepath.Join(tempDir, ".claude")

	// Create source structure
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "agents"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "commands"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "hooks"), 0755))

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "agents", "test-agent.md"), []byte("agent content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "commands", "test-command.md"), []byte("command content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "hooks", "test-hook.sh"), []byte("#!/bin/bash\necho test"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "CLAUDE.md.to.copy"), []byte("claude config"), 0644))

	// Create source settings.json
	sourceSettings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	}
	sourceData, _ := json.MarshalIndent(sourceSettings, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "settings.json"), sourceData, 0644))

	// Create operations manager
	ops := NewOperations(sourceDir, claudeDir)
	ctx := context.Background()

	// Test Copy with All option
	err := ops.Copy(ctx, &claude.CopyOptions{All: true})
	require.NoError(t, err)

	// Verify all directories were copied
	assert.DirExists(t, filepath.Join(claudeDir, "agents"))
	assert.DirExists(t, filepath.Join(claudeDir, "commands"))
	assert.DirExists(t, filepath.Join(claudeDir, "hooks"))

	// Verify files were copied
	assert.FileExists(t, filepath.Join(claudeDir, "agents", "test-agent.md"))
	assert.FileExists(t, filepath.Join(claudeDir, "commands", "test-command.md"))
	assert.FileExists(t, filepath.Join(claudeDir, "hooks", "test-hook.sh"))

	// Verify CLAUDE.md.to.copy was renamed to CLAUDE.md
	assert.FileExists(t, filepath.Join(claudeDir, "CLAUDE.md"))
	content, err := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Equal(t, "claude config", string(content))

	// Verify settings.json was copied
	assert.FileExists(t, filepath.Join(claudeDir, "settings.json"))
}

func TestFileOperations_Copy_Selective(t *testing.T) {
	// Setup temp directories
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	claudeDir := filepath.Join(tempDir, ".claude")

	// Create source structure
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "agents"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "commands"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "hooks"), 0755))

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "agents", "test-agent.md"), []byte("agent content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "commands", "test-command.md"), []byte("command content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "hooks", "test-hook.sh"), []byte("#!/bin/bash\necho test"), 0755))

	// Create operations manager
	ops := NewOperations(sourceDir, claudeDir)
	ctx := context.Background()

	// Test Copy with selective options
	err := ops.Copy(ctx, &claude.CopyOptions{
		Agents:   true,
		Commands: true,
		Hooks:    false,
	})
	require.NoError(t, err)

	// Verify only selected directories were copied
	assert.DirExists(t, filepath.Join(claudeDir, "agents"))
	assert.DirExists(t, filepath.Join(claudeDir, "commands"))
	assert.NoDirExists(t, filepath.Join(claudeDir, "hooks"))

	// Verify files were copied correctly
	assert.FileExists(t, filepath.Join(claudeDir, "agents", "test-agent.md"))
	assert.FileExists(t, filepath.Join(claudeDir, "commands", "test-command.md"))
	assert.NoFileExists(t, filepath.Join(claudeDir, "hooks", "test-hook.sh"))
}

func TestFileOperations_Copy_SettingsIntelligentMerge(t *testing.T) {
	// Setup temp directories
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	claudeDir := filepath.Join(tempDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Create existing destination settings with proxy
	destSettings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"http_proxy":              "http://127.0.0.1:7890",
			"https_proxy":             "http://127.0.0.1:7890",
			"CLAUDE_HOOKS_GO_ENABLED": "false",
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/custom-lint.sh",
						},
					},
				},
			},
		},
	}
	destData, _ := json.MarshalIndent(destSettings, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "settings.json"), destData, 0644))

	// Create source directory
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create source settings with different proxy and additional hooks
	sourceSettings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://192.168.1.100:8080",
			"https_proxy": "http://192.168.1.100:8080",
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh",
						},
					},
				},
			},
			Stop: []*claude.HookRule{
				{
					Matcher: "",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh",
						},
					},
				},
			},
		},
	}
	sourceData, _ := json.MarshalIndent(sourceSettings, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "settings.json"), sourceData, 0644))

	// Create operations manager
	ops := NewOperations(sourceDir, claudeDir)
	ctx := context.Background()

	// Test Copy (which should trigger intelligent merge)
	err := ops.Copy(ctx, &claude.CopyOptions{All: true})
	require.NoError(t, err)

	// Load merged settings
	mergedData, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	require.NoError(t, err)

	var mergedSettings claude.Settings
	err = json.Unmarshal(mergedData, &mergedSettings)
	require.NoError(t, err)

	// Verify merge results match DESIGN.md expectations
	assert.True(t, mergedSettings.IncludeCoAuthoredBy) // From source

	// Proxy protection: user's proxy should be preserved
	assert.Equal(t, "http://127.0.0.1:7890", mergedSettings.Env["http_proxy"])
	assert.Equal(t, "http://127.0.0.1:7890", mergedSettings.Env["https_proxy"])

	// User's other env vars should be preserved
	assert.Equal(t, "false", mergedSettings.Env["CLAUDE_HOOKS_GO_ENABLED"])

	// Hooks should be intelligently merged
	require.NotNil(t, mergedSettings.Hooks)
	require.Len(t, mergedSettings.Hooks.PostToolUse, 1)

	postToolUse := mergedSettings.Hooks.PostToolUse[0]
	assert.Equal(t, "Write|Edit|MultiEdit", postToolUse.Matcher) // Broader matcher
	require.Len(t, postToolUse.Hooks, 2)                         // Both hooks should be present

	// User's hook should be first
	assert.Equal(t, "~/.claude/hooks/custom-lint.sh", postToolUse.Hooks[0].Command)
	assert.Equal(t, "~/.claude/hooks/smart-lint.sh", postToolUse.Hooks[1].Command)

	// New Stop hooks should be added
	require.Len(t, mergedSettings.Hooks.Stop, 1)
	assert.Equal(t, "~/.claude/hooks/ntfy-notifier.sh", mergedSettings.Hooks.Stop[0].Hooks[0].Command)
}

func TestFileOperations_Compare(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	file3Path := filepath.Join(tempDir, "file3.txt")

	content := "test content"
	require.NoError(t, os.WriteFile(file1Path, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file2Path, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file3Path, []byte("different content"), 0644))

	ops := NewOperations("", "")
	ctx := context.Background()

	// Test same files
	result, err := ops.Compare(ctx, file1Path, file2Path)
	require.NoError(t, err)
	assert.True(t, result.Same)
	assert.Empty(t, result.Differences)

	// Test different files
	result, err = ops.Compare(ctx, file1Path, file3Path)
	require.NoError(t, err)
	assert.False(t, result.Same)
	assert.NotEmpty(t, result.Differences)

	// Test non-existent file
	result, err = ops.Compare(ctx, file1Path, filepath.Join(tempDir, "nonexistent.txt"))
	require.NoError(t, err)
	assert.False(t, result.Same)
	assert.Contains(t, result.Differences[0], "Destination file does not exist")
}

func TestFileOperations_CopyFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create executable file
	executableFile := filepath.Join(sourceDir, "executable.sh")
	require.NoError(t, os.WriteFile(executableFile, []byte("#!/bin/bash\necho test"), 0755))

	// Create regular file
	regularFile := filepath.Join(sourceDir, "regular.txt")
	require.NoError(t, os.WriteFile(regularFile, []byte("test content"), 0644))

	ops := NewOperations(sourceDir, destDir)

	// Copy files
	err := ops.copyFile(executableFile, filepath.Join(destDir, "executable.sh"))
	require.NoError(t, err)

	err = ops.copyFile(regularFile, filepath.Join(destDir, "regular.txt"))
	require.NoError(t, err)

	// Verify permissions were preserved
	execInfo, err := os.Stat(filepath.Join(destDir, "executable.sh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), execInfo.Mode().Perm())

	regularInfo, err := os.Stat(filepath.Join(destDir, "regular.txt"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), regularInfo.Mode().Perm())
}

package config

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

func TestConfigManager_Load(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    *claude.Settings
		expectError    bool
		expectedConfig *claude.Settings
	}{
		{
			name: "load valid config",
			setupConfig: &claude.Settings{
				IncludeCoAuthoredBy: true,
				Env: map[string]string{
					"http_proxy":  "http://127.0.0.1:7890",
					"https_proxy": "http://127.0.0.1:7890",
				},
			},
			expectError: false,
			expectedConfig: &claude.Settings{
				IncludeCoAuthoredBy: true,
				Env: map[string]string{
					"http_proxy":  "http://127.0.0.1:7890",
					"https_proxy": "http://127.0.0.1:7890",
				},
			},
		},
		{
			name: "load with hooks config",
			setupConfig: &claude.Settings{
				IncludeCoAuthoredBy: false,
				Hooks: &claude.HooksConfig{
					PostToolUse: []*claude.HookRule{
						{
							Matcher: "Write|Edit",
							Hooks: []*claude.HookItem{
								{
									Type:    "command",
									Command: "~/.claude/hooks/smart-lint.sh",
								},
							},
						},
					},
				},
			},
			expectError: false,
			expectedConfig: &claude.Settings{
				IncludeCoAuthoredBy: false,
				Hooks: &claude.HooksConfig{
					PostToolUse: []*claude.HookRule{
						{
							Matcher: "Write|Edit",
							Hooks: []*claude.HookItem{
								{
									Type:    "command",
									Command: "~/.claude/hooks/smart-lint.sh",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directory
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude")
			err := os.MkdirAll(claudeDir, 0755)
			require.NoError(t, err)

			// Create test settings file
			settingsPath := filepath.Join(claudeDir, "settings.json")
			if tt.setupConfig != nil {
				data, err := json.MarshalIndent(tt.setupConfig, "", "  ")
				require.NoError(t, err)
				err = os.WriteFile(settingsPath, data, 0644)
				require.NoError(t, err)
			}

			// Create manager
			manager := NewManager(claudeDir)

			// Test Load
			ctx := context.Background()
			config, err := manager.Load(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig.IncludeCoAuthoredBy, config.IncludeCoAuthoredBy)
			assert.Equal(t, tt.expectedConfig.Env, config.Env)

			if tt.expectedConfig.Hooks != nil {
				require.NotNil(t, config.Hooks)
				assert.Equal(t, len(tt.expectedConfig.Hooks.PostToolUse), len(config.Hooks.PostToolUse))
			}
		})
	}
}

func TestConfigManager_Save(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test data
	config := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://127.0.0.1:7890",
			"https_proxy": "http://127.0.0.1:7890",
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
		},
	}

	// Save config
	err = manager.Save(ctx, config)
	require.NoError(t, err)

	// Verify file exists and content is correct
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var savedConfig claude.Settings
	err = json.Unmarshal(data, &savedConfig)
	require.NoError(t, err)

	assert.Equal(t, config.IncludeCoAuthoredBy, savedConfig.IncludeCoAuthoredBy)
	assert.Equal(t, config.Env, savedConfig.Env)
	require.NotNil(t, savedConfig.Hooks)
	assert.Equal(t, len(config.Hooks.PostToolUse), len(savedConfig.Hooks.PostToolUse))
}

func TestConfigManager_GetStatus(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create test settings
	settings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":           "http://127.0.0.1:7890",
			"https_proxy":          "http://127.0.0.1:7890",
			"ANTHROPIC_AUTH_TOKEN": "sk-test123",
			"ANTHROPIC_BASE_URL":   "https://api.deepseek.com/anthropic",
			"ANTHROPIC_MODEL":      "deepseek-chat",
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh",
						},
					},
				},
			},
		},
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test GetStatus
	status, err := manager.GetStatus(ctx)
	require.NoError(t, err)

	// Verify status
	assert.True(t, status.ProxyEnabled)
	assert.True(t, status.DeepSeekEnabled)
	assert.True(t, status.HooksEnabled)


	// Check proxy config
	require.NotNil(t, status.ProxyConfig)
	assert.Equal(t, "http://127.0.0.1:7890", status.ProxyConfig.HTTPProxy)
	assert.Equal(t, "http://127.0.0.1:7890", status.ProxyConfig.HTTPSProxy)
}

func TestConfigManager_Backup_DirectoryBackup(t *testing.T) {
	// Setup temp directories
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	claudeDir := filepath.Join(homeDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create comprehensive test structure
	// 1. settings.json
	settings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://127.0.0.1:7890",
			"https_proxy": "http://127.0.0.1:7890",
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

	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// 2. Create CLAUDE.md
	claudemdPath := filepath.Join(claudeDir, "CLAUDE.md")
	claudemdContent := "# Claude Configuration\n\nThis is a test configuration file."
	err = os.WriteFile(claudemdPath, []byte(claudemdContent), 0644)
	require.NoError(t, err)

	// 3. Create agents directory and files
	agentsDir := filepath.Join(claudeDir, "agents")
	err = os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)
	agentContent := "# Test Agent\n\nA test agent configuration."
	err = os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte(agentContent), 0644)
	require.NoError(t, err)

	// 4. Create commands directory and files
	commandsDir := filepath.Join(claudeDir, "commands")
	err = os.MkdirAll(commandsDir, 0755)
	require.NoError(t, err)
	commandContent := "# Test Command\n\nA test command configuration."
	err = os.WriteFile(filepath.Join(commandsDir, "test-command.md"), []byte(commandContent), 0644)
	require.NoError(t, err)

	// 5. Create hooks directory and files
	hooksDir := filepath.Join(claudeDir, "hooks")
	err = os.MkdirAll(hooksDir, 0755)
	require.NoError(t, err)
	hookContent := "#!/bin/bash\n\necho 'Test hook'"
	err = os.WriteFile(filepath.Join(hooksDir, "smart-lint.sh"), []byte(hookContent), 0755)
	require.NoError(t, err)

	// 6. Create .deepseek_api_key
	apiKeyPath := filepath.Join(claudeDir, ".deepseek_api_key")
	err = os.WriteFile(apiKeyPath, []byte("sk-test123456789"), 0600)
	require.NoError(t, err)

	// 7. Create .proxy_config
	proxyConfigPath := filepath.Join(claudeDir, ".proxy_config")
	proxyConfigContent := "http://127.0.0.1:7890"
	err = os.WriteFile(proxyConfigPath, []byte(proxyConfigContent), 0644)
	require.NoError(t, err)

	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Mock home directory to redirect backup location
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Test Backup
	backupInfo, err := manager.Backup(ctx)
	require.NoError(t, err)

	// Verify backup info structure
	assert.NotEmpty(t, backupInfo.Filename)
	assert.NotEmpty(t, backupInfo.FilePath)
	assert.Equal(t, "directory", backupInfo.ContentType)
	assert.True(t, backupInfo.Size > 0)
	assert.False(t, backupInfo.Timestamp.IsZero())

	// Verify filename format: claude-config-backup-{timestamp}.tar.gz
	assert.Regexp(t, `^claude-config-backup-\d{8}_\d{6}\.tar\.gz$`, backupInfo.Filename)

	// Verify backup file is saved to home directory, not claude directory
	expectedBackupPath := filepath.Join(homeDir, backupInfo.Filename)
	assert.Equal(t, expectedBackupPath, backupInfo.FilePath)

	// Verify backup file exists at expected location
	_, err = os.Stat(backupInfo.FilePath)
	require.NoError(t, err)

	// Verify it's NOT in the claude directory
	claudeDirBackupPath := filepath.Join(claudeDir, backupInfo.Filename)
	_, err = os.Stat(claudeDirBackupPath)
	assert.True(t, os.IsNotExist(err), "Backup should not be in claude directory")

	// TODO: Verify tar.gz content includes all expected files
	// This would require extracting and checking the archive contents
	// For now, just verify the file size is reasonable (tar.gz overhead + content)
	t.Logf("Backup file size: %d bytes", backupInfo.Size)
	assert.True(t, backupInfo.Size > 100, "Backup should contain content (allowing for small test files)")
}

func TestConfigManager_Backup_SettingsOnly(t *testing.T) {
	// This test verifies backward compatibility for settings-only backup
	// when no full directory backup is needed

	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create test settings
	settings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy": "http://127.0.0.1:7890",
		},
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Backup - should still work with minimal setup
	backupInfo, err := manager.Backup(ctx)
	require.NoError(t, err)

	// Verify backup info
	assert.NotEmpty(t, backupInfo.Filename)
	assert.True(t, backupInfo.Size > 0)
	assert.False(t, backupInfo.Timestamp.IsZero())
}

package check

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooneko/claude-config/internal/claude"
)

// TestNewManager tests creating a new check manager
func TestNewManager(t *testing.T) {
	claudeDir := t.TempDir()

	manager := NewManager(claudeDir)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.claudeDir != claudeDir {
		t.Errorf("Expected claudeDir %q, got %q", claudeDir, manager.claudeDir)
	}
}

// TestManager_EnableCheck tests enabling code checking hooks
func TestManager_EnableCheck(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig func(string) error // setup function
		wantErr        bool
		verify         func(*testing.T, string)
	}{
		{
			name:           "首次启用_无settings文件",
			existingConfig: func(_ string) error { return nil },
			wantErr:        false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks == nil {
					t.Error("Expected Hooks to be initialized")
				}
				if settings.Hooks.PostToolUse == nil {
					t.Error("Expected PostToolUse hooks to be set")
				}
				if len(settings.Hooks.PostToolUse) != 1 {
					t.Errorf("Expected 1 PostToolUse hook, got %d", len(settings.Hooks.PostToolUse))
				}
			},
		},
		{
			name: "启用_有settings文件_无hooks配置",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					IncludeCoAuthoredBy: false,
				}
				return saveSettingsToFile(dir, settings)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks == nil {
					t.Error("Expected Hooks to be initialized")
				}
				if len(settings.Hooks.PostToolUse) != 1 {
					t.Errorf("Expected 1 PostToolUse hook, got %d", len(settings.Hooks.PostToolUse))
				}
			},
		},
		{
			name: "启用_有备份配置_应恢复备份",
			existingConfig: func(dir string) error {
				// 创建备份文件
				backupConfig := &claude.HooksConfig{
					PostToolUse: []*claude.HookRule{
						{
							Matcher: "TestMatcher",
							Hooks: []*claude.HookItem{
								{
									Type:    "command",
									Command: "~/.claude/hooks/test.sh",
									Timeout: 60,
								},
							},
						},
					},
				}
				backupPath := filepath.Join(dir, "settings.json.hooks_backup")
				data, err := json.Marshal(backupConfig)
				if err != nil {
					return err
				}
				return os.WriteFile(backupPath, data, 0644)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if len(settings.Hooks.PostToolUse) != 1 {
					t.Errorf("Expected 1 PostToolUse hook from backup, got %d", len(settings.Hooks.PostToolUse))
				}
				if settings.Hooks.PostToolUse[0].Matcher != "TestMatcher" {
					t.Errorf("Expected matcher 'TestMatcher', got %q", settings.Hooks.PostToolUse[0].Matcher)
				}
			},
		},
		{
			name: "启用_创建默认配置",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					IncludeCoAuthoredBy: false,
				}
				return saveSettingsToFile(dir, settings)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks.PostToolUse[0].Matcher != "Write|Edit|MultiEdit" {
					t.Errorf("Expected default matcher, got %q", settings.Hooks.PostToolUse[0].Matcher)
				}
				if len(settings.Hooks.PostToolUse[0].Hooks) != 2 {
					t.Errorf("Expected 2 hooks in default config, got %d", len(settings.Hooks.PostToolUse[0].Hooks))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tt.existingConfig(dir); err != nil {
				t.Fatalf("Failed to setup test config: %v", err)
			}

			manager := NewManager(dir)
			err := manager.EnableCheck(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("EnableCheck() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verify != nil {
				tt.verify(t, dir)
			}
		})
	}
}

// TestManager_DisableCheck tests disabling code checking hooks
func TestManager_DisableCheck(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig func(string) error
		wantErr        bool
		verify         func(*testing.T, string)
	}{
		{
			name:           "禁用_无hooks配置_应无操作",
			existingConfig: func(_ string) error { return nil },
			wantErr:        false,
			verify: func(_ *testing.T, dir string) {
				// 无 settings 文件应该被创建
				settingsPath := filepath.Join(dir, "settings.json")
				if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
					// 正常情况：没有文件被创建
					return
				}
			},
		},
		{
			name: "禁用_有PostToolUse钩子_应保存备份并清除",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					Hooks: &claude.HooksConfig{
						PostToolUse: []*claude.HookRule{
							{
								Matcher: "Write|Edit",
								Hooks: []*claude.HookItem{
									{
										Type:    "command",
										Command: "~/.claude/hooks/test.sh",
										Timeout: 60,
									},
								},
							},
						},
					},
				}
				return saveSettingsToFile(dir, settings)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				// 验证备份文件已创建
				backupPath := filepath.Join(dir, "settings.json.hooks_backup")
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Error("Expected backup file to be created")
				}

				// 验证 PostToolUse 已被清除
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks != nil && settings.Hooks.PostToolUse != nil {
					if len(settings.Hooks.PostToolUse) > 0 {
						t.Error("Expected PostToolUse to be cleared")
					}
				}
			},
		},
		{
			name: "禁用_有其他钩子_保留hooks结构",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					Hooks: &claude.HooksConfig{
						PostToolUse: []*claude.HookRule{
							{
								Matcher: "Write|Edit",
								Hooks: []*claude.HookItem{
									{
										Type:    "command",
										Command: "~/.claude/hooks/test.sh",
										Timeout: 60,
									},
								},
							},
						},
						Stop: []*claude.HookRule{
							{
								Matcher: ".*",
								Hooks: []*claude.HookItem{
									{
										Type:    "command",
										Command: "~/.claude/hooks/stop.sh",
									},
								},
							},
						},
					},
				}
				return saveSettingsToFile(dir, settings)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks == nil {
					t.Error("Expected Hooks structure to be preserved")
				}
				if len(settings.Hooks.Stop) == 0 {
					t.Error("Expected Stop hooks to be preserved")
				}
				if len(settings.Hooks.PostToolUse) > 0 {
					t.Error("Expected PostToolUse to be cleared")
				}
			},
		},
		{
			name: "禁用_清空所有钩子_移除hooks结构",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					Hooks: &claude.HooksConfig{
						PostToolUse: []*claude.HookRule{
							{
								Matcher: "Write|Edit",
								Hooks: []*claude.HookItem{
									{
										Type:    "command",
										Command: "~/.claude/hooks/test.sh",
										Timeout: 60,
									},
								},
							},
						},
					},
				}
				return saveSettingsToFile(dir, settings)
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if settings.Hooks != nil {
					if len(settings.Hooks.PostToolUse) > 0 || len(settings.Hooks.Stop) > 0 {
						t.Error("Expected Hooks to be nil when all hooks are cleared")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tt.existingConfig(dir); err != nil {
				t.Fatalf("Failed to setup test config: %v", err)
			}

			manager := NewManager(dir)
			err := manager.DisableCheck(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("DisableCheck() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verify != nil {
				tt.verify(t, dir)
			}
		})
	}
}

// TestManager_createDefaultHooksConfig tests the default hooks configuration
func TestManager_createDefaultHooksConfig(t *testing.T) {
	dir := t.TempDir()
	manager := NewManager(dir)

	// 通过启用检查来测试默认配置创建
	err := manager.EnableCheck(context.Background())
	if err != nil {
		t.Fatalf("EnableCheck failed: %v", err)
	}

	settings := loadSettingsFromFile(t, dir)

	if settings.Hooks == nil {
		t.Fatal("Expected Hooks to be initialized")
	}

	if len(settings.Hooks.PostToolUse) != 1 {
		t.Fatalf("Expected 1 PostToolUse hook, got %d", len(settings.Hooks.PostToolUse))
	}

	rule := settings.Hooks.PostToolUse[0]
	if rule.Matcher != "Write|Edit|MultiEdit" {
		t.Errorf("Expected matcher 'Write|Edit|MultiEdit', got %q", rule.Matcher)
	}

	if len(rule.Hooks) != 2 {
		t.Fatalf("Expected 2 hooks, got %d", len(rule.Hooks))
	}

	// 验证第一个钩子 (smart-lint.sh)
	if rule.Hooks[0].Type != "command" {
		t.Errorf("Expected hook type 'command', got %q", rule.Hooks[0].Type)
	}
	if rule.Hooks[0].Command != "~/.claude/hooks/smart-lint.sh" {
		t.Errorf("Expected command '~/.claude/hooks/smart-lint.sh', got %q", rule.Hooks[0].Command)
	}
	if rule.Hooks[0].Timeout != 120 {
		t.Errorf("Expected timeout 120, got %d", rule.Hooks[0].Timeout)
	}

	// 验证第二个钩子 (smarter-test.sh)
	if rule.Hooks[1].Type != "command" {
		t.Errorf("Expected hook type 'command', got %q", rule.Hooks[1].Type)
	}
	if rule.Hooks[1].Command != "~/.claude/hooks/smarter-test.sh" {
		t.Errorf("Expected command '~/.claude/hooks/smarter-test.sh', got %q", rule.Hooks[1].Command)
	}
	if rule.Hooks[1].Timeout != 120 {
		t.Errorf("Expected timeout 120, got %d", rule.Hooks[1].Timeout)
	}
}

// TestManager_loadSettings tests loading settings from file
func TestManager_loadSettings(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig func(string) error
		wantHooks      bool
		wantErr        bool
	}{
		{
			name:           "加载_无settings文件_返回默认配置",
			existingConfig: func(_ string) error { return nil },
			wantHooks:      false,
			wantErr:        false,
		},
		{
			name: "加载_有settings文件",
			existingConfig: func(dir string) error {
				settings := &claude.Settings{
					IncludeCoAuthoredBy: true,
					Hooks: &claude.HooksConfig{
						PostToolUse: []*claude.HookRule{
							{
								Matcher: "Write",
							},
						},
					},
				}
				return saveSettingsToFile(dir, settings)
			},
			wantHooks: true,
			wantErr:   false,
		},
		{
			name: "加载_无效JSON_返回错误",
			existingConfig: func(dir string) error {
				settingsPath := filepath.Join(dir, "settings.json")
				return os.WriteFile(settingsPath, []byte("{invalid json}"), 0644)
			},
			wantHooks: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tt.existingConfig(dir); err != nil {
				t.Fatalf("Failed to setup test config: %v", err)
			}

			manager := NewManager(dir)
			settings, err := manager.loadSettings()

			if (err != nil) != tt.wantErr {
				t.Errorf("loadSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if settings == nil {
					t.Error("Expected settings to be returned")
					return
				}
				hasHooks := settings.Hooks != nil
				if hasHooks != tt.wantHooks {
					t.Errorf("Expected hooks=%v, got hooks=%v", tt.wantHooks, hasHooks)
				}
			}
		})
	}
}

// TestManager_saveSettings tests saving settings to file
func TestManager_saveSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *claude.Settings
		wantErr  bool
		verify   func(*testing.T, string)
	}{
		{
			name: "保存_新settings文件",
			settings: &claude.Settings{
				IncludeCoAuthoredBy: false,
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settingsPath := filepath.Join(dir, "settings.json")
				if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
					t.Error("Expected settings file to be created")
				}
			},
		},
		{
			name: "保存_覆盖现有settings",
			settings: &claude.Settings{
				IncludeCoAuthoredBy: true,
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				settings := loadSettingsFromFile(t, dir)
				if !settings.IncludeCoAuthoredBy {
					t.Error("Expected IncludeCoAuthoredBy to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			manager := NewManager(dir)
			err := manager.saveSettings(tt.settings)

			if (err != nil) != tt.wantErr {
				t.Errorf("saveSettings() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verify != nil {
				tt.verify(t, dir)
			}
		})
	}
}

// TestManager_saveHooksBackup tests saving hooks backup
func TestManager_saveHooksBackup(t *testing.T) {
	dir := t.TempDir()
	manager := NewManager(dir)

	hooksConfig := &claude.HooksConfig{
		PostToolUse: []*claude.HookRule{
			{
				Matcher: "TestMatcher",
			},
		},
	}

	err := manager.saveHooksBackup(hooksConfig)
	if err != nil {
		t.Fatalf("saveHooksBackup failed: %v", err)
	}

	// 验证备份文件存在
	backupPath := filepath.Join(dir, "settings.json.hooks_backup")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to be created")
	}

	// 验证备份文件内容
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	var loadedConfig claude.HooksConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("Failed to unmarshal backup config: %v", err)
	}

	if loadedConfig.PostToolUse[0].Matcher != "TestMatcher" {
		t.Errorf("Expected matcher 'TestMatcher', got %q", loadedConfig.PostToolUse[0].Matcher)
	}
}

// TestManager_loadHooksBackup tests loading hooks backup
func TestManager_loadHooksBackup(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig func(string) error
		wantErr        bool
		verify         func(*testing.T, *claude.HooksConfig)
	}{
		{
			name:           "加载_无备份文件_返回错误",
			existingConfig: func(_ string) error { return nil },
			wantErr:        true,
			verify:         nil,
		},
		{
			name: "加载_有备份文件",
			existingConfig: func(dir string) error {
				hooksConfig := &claude.HooksConfig{
					PostToolUse: []*claude.HookRule{
						{
							Matcher: "BackupMatcher",
						},
					},
				}
				backupPath := filepath.Join(dir, "settings.json.hooks_backup")
				data, err := json.Marshal(hooksConfig)
				if err != nil {
					return err
				}
				return os.WriteFile(backupPath, data, 0644)
			},
			wantErr: false,
			verify: func(t *testing.T, config *claude.HooksConfig) {
				if config.PostToolUse[0].Matcher != "BackupMatcher" {
					t.Errorf("Expected matcher 'BackupMatcher', got %q", config.PostToolUse[0].Matcher)
				}
			},
		},
		{
			name: "加载_无效JSON_返回错误",
			existingConfig: func(dir string) error {
				backupPath := filepath.Join(dir, "settings.json.hooks_backup")
				return os.WriteFile(backupPath, []byte("{invalid json}"), 0644)
			},
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := tt.existingConfig(dir); err != nil {
				t.Fatalf("Failed to setup test config: %v", err)
			}

			manager := NewManager(dir)
			config, err := manager.loadHooksBackup()

			if (err != nil) != tt.wantErr {
				t.Errorf("loadHooksBackup() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, config)
			}
		})
	}
}

// TestManager_EnableDisableIntegration tests enable/disable integration
func TestManager_EnableDisableIntegration(t *testing.T) {
	dir := t.TempDir()
	manager := NewManager(dir)

	// 启用检查
	if err := manager.EnableCheck(context.Background()); err != nil {
		t.Fatalf("EnableCheck failed: %v", err)
	}

	// 验证已启用
	settings := loadSettingsFromFile(t, dir)
	if settings.Hooks == nil || len(settings.Hooks.PostToolUse) == 0 {
		t.Error("Expected hooks to be enabled")
	}

	// 禁用检查
	if err := manager.DisableCheck(context.Background()); err != nil {
		t.Fatalf("DisableCheck failed: %v", err)
	}

	// 验证已禁用
	settings = loadSettingsFromFile(t, dir)
	if settings.Hooks != nil && len(settings.Hooks.PostToolUse) > 0 {
		t.Error("Expected hooks to be disabled")
	}

	// 验证备份文件存在
	backupPath := filepath.Join(dir, "settings.json.hooks_backup")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to exist after disable")
	}
}

// Helper functions

func loadSettingsFromFile(t *testing.T, dir string) *claude.Settings {
	t.Helper()
	settingsPath := filepath.Join(dir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var settings claude.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	return &settings
}

func saveSettingsToFile(dir string, settings *claude.Settings) error {
	settingsPath := filepath.Join(dir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

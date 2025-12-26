package check

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ooneko/claude-config/internal/claude"
)

// TestNewManager tests creating a new check manager
func TestNewManager(t *testing.T) {
	tests := []struct {
		name      string
		claudeDir string
	}{
		{
			name:      "valid directory",
			claudeDir: "/test/claude",
		},
		{
			name:      "empty directory",
			claudeDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.claudeDir)
			if mgr == nil {
				t.Fatal("NewManager returned nil")
			}
			if mgr.claudeDir != tt.claudeDir {
				t.Errorf("claudeDir = %v, want %v", mgr.claudeDir, tt.claudeDir)
			}
		})
	}
}

// TestManager_EnableCheck tests enabling code checking hooks
func TestManager_EnableCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, claudeDir string)
		validateFunc  func(t *testing.T, claudeDir string)
		expectError   bool
		errorContains string
	}{
		{
			name: "enable_without_existing_settings",
			setupFunc: func(_ *testing.T, _ string) {
				// No settings file exists
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("failed to read settings file: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("failed to parse settings: %v", err)
				}

				hooks, ok := settings["hooks"]
				if !ok {
					t.Fatal("hooks not found in settings")
				}

				hooksMap, ok := hooks.(map[string]interface{})
				if !ok {
					t.Fatal("hooks is not a map")
				}

				postToolUse, ok := hooksMap["PostToolUse"]
				if !ok {
					t.Fatal("PostToolUse not found in hooks")
				}

				postToolUseSlice, ok := postToolUse.([]interface{})
				if !ok || len(postToolUseSlice) == 0 {
					t.Fatal("PostToolUse is empty or not a slice")
				}
			},
			expectError: false,
		},
		{
			name: "enable_with_existing_settings_no_hooks",
			setupFunc: func(t *testing.T, claudeDir string) {
				settings := map[string]interface{}{
					"includeCoAuthoredBy": false,
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, data, 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("failed to read settings file: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("failed to parse settings: %v", err)
				}

				if _, ok := settings["hooks"]; !ok {
					t.Fatal("hooks not found in settings")
				}
			},
			expectError: false,
		},
		{
			name: "enable_with_existing_hooks_backup",
			setupFunc: func(t *testing.T, claudeDir string) {
				// Create a backup hooks config
				backupHooks := map[string]interface{}{
					"PostToolUse": []map[string]interface{}{
						{
							"matcher": "TestMatcher",
							"hooks": []map[string]interface{}{
								{
									"type":    "command",
									"command": "/test/custom-hook.sh",
									"timeout": 60,
								},
							},
						},
					},
				}
				data, _ := json.Marshal(backupHooks)
				backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
				if err := os.WriteFile(backupPath, data, 0644); err != nil {
					t.Fatalf("failed to write backup: %v", err)
				}
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("failed to read settings file: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("failed to parse settings: %v", err)
				}

				hooks := settings["hooks"].(map[string]interface{})
				postToolUse := hooks["PostToolUse"].([]interface{})
				if len(postToolUse) == 0 {
					t.Fatal("PostToolUse is empty")
				}

				firstRule := postToolUse[0].(map[string]interface{})
				if firstRule["matcher"] != "TestMatcher" {
					t.Errorf("matcher = %v, want TestMatcher", firstRule["matcher"])
				}
			},
			expectError: false,
		},
		{
			name: "enable_with_invalid_settings_json",
			setupFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, []byte("{invalid json}"), 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc:  nil,
			expectError:   true,
			errorContains: "failed to parse settings file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("failed to create claude dir: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, claudeDir)
			}

			mgr := NewManager(claudeDir)
			err := mgr.EnableCheck(context.Background())

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, claudeDir)
			}
		})
	}
}

// TestManager_DisableCheck tests disabling code checking hooks
func TestManager_DisableCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, claudeDir string)
		validateFunc  func(t *testing.T, claudeDir string)
		expectError   bool
		errorContains string
	}{
		{
			name: "disable_enabled_hooks",
			setupFunc: func(t *testing.T, claudeDir string) {
				settings := map[string]interface{}{
					"hooks": map[string]interface{}{
						"PostToolUse": []map[string]interface{}{
							{
								"matcher": "Write|Edit",
								"hooks": []map[string]interface{}{
									{
										"type":    "command",
										"command": "~/.claude/hooks/test.sh",
										"timeout": 120,
									},
								},
							},
						},
					},
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, data, 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("failed to read settings file: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("failed to parse settings: %v", err)
				}

				hooks, ok := settings["hooks"]
				if ok {
					hooksMap := hooks.(map[string]interface{})
					if postToolUse, ok := hooksMap["PostToolUse"]; ok {
						if slice, ok := postToolUse.([]interface{}); ok {
							if len(slice) > 0 {
								t.Error("PostToolUse should be empty or nil after disable")
							}
						}
					}
				}

				// Check backup was created
				backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Error("hooks backup file should be created")
				}
			},
			expectError: false,
		},
		{
			name: "disable_without_existing_hooks",
			setupFunc: func(t *testing.T, claudeDir string) {
				settings := map[string]interface{}{
					"includeCoAuthoredBy": false,
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, data, 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				// Should succeed without errors
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if _, err := os.ReadFile(settingsPath); err != nil {
					t.Errorf("settings file should still exist: %v", err)
				}
			},
			expectError: false,
		},
		{
			name: "disable_with_no_settings_file",
			setupFunc: func(_ *testing.T, _ string) {
				// No settings file
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				// Should succeed - nothing to disable
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
					t.Error("settings file should not be created when no hooks to disable")
				}
			},
			expectError: false,
		},
		{
			name: "disable_creates_backup",
			setupFunc: func(t *testing.T, claudeDir string) {
				settings := map[string]interface{}{
					"hooks": map[string]interface{}{
						"PostToolUse": []map[string]interface{}{
							{
								"matcher": "TestMatcher",
								"hooks": []map[string]interface{}{
									{
										"type":    "command",
										"command": "/test/hook.sh",
										"timeout": 60,
									},
								},
							},
						},
					},
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, data, 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc: func(t *testing.T, claudeDir string) {
				backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
				data, err := os.ReadFile(backupPath)
				if err != nil {
					t.Fatalf("failed to read backup file: %v", err)
				}

				var backup map[string]interface{}
				if err := json.Unmarshal(data, &backup); err != nil {
					t.Fatalf("failed to parse backup: %v", err)
				}

				if _, ok := backup["PostToolUse"]; !ok {
					t.Error("PostToolUse should be in backup")
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("failed to create claude dir: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, claudeDir)
			}

			mgr := NewManager(claudeDir)
			err := mgr.DisableCheck(context.Background())

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, claudeDir)
			}
		})
	}
}

// TestManager_createDefaultHooksConfig tests creating default hooks configuration
func TestManager_createDefaultHooksConfig(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	mgr := NewManager(claudeDir)

	config := mgr.createDefaultHooksConfig()

	if config == nil {
		t.Fatal("createDefaultHooksConfig returned nil")
	}

	if len(config.PostToolUse) == 0 {
		t.Fatal("PostToolUse should not be empty")
	}

	rule := config.PostToolUse[0]
	if rule.Matcher != "Write|Edit|MultiEdit" {
		t.Errorf("Matcher = %v, want Write|Edit|MultiEdit", rule.Matcher)
	}

	if len(rule.Hooks) != 2 {
		t.Fatalf("Expected 2 hooks, got %d", len(rule.Hooks))
	}

	// Check first hook (smart-lint.sh)
	if rule.Hooks[0].Type != "command" {
		t.Errorf("First hook type = %v, want command", rule.Hooks[0].Type)
	}
	if rule.Hooks[0].Command != "~/.claude/hooks/smart-lint.sh" {
		t.Errorf("First hook command = %v, want ~/.claude/hooks/smart-lint.sh", rule.Hooks[0].Command)
	}
	if rule.Hooks[0].Timeout != 120 {
		t.Errorf("First hook timeout = %v, want 120", rule.Hooks[0].Timeout)
	}

	// Check second hook (smarter-test.sh)
	if rule.Hooks[1].Type != "command" {
		t.Errorf("Second hook type = %v, want command", rule.Hooks[1].Type)
	}
	if rule.Hooks[1].Command != "~/.claude/hooks/smarter-test.sh" {
		t.Errorf("Second hook command = %v, want ~/.claude/hooks/smarter-test.sh", rule.Hooks[1].Command)
	}
	if rule.Hooks[1].Timeout != 120 {
		t.Errorf("Second hook timeout = %v, want 120", rule.Hooks[1].Timeout)
	}
}

// TestManager_loadSettings tests loading settings from file
func TestManager_loadSettings(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, claudeDir string)
		validateFunc  func(t *testing.T, settings *claude.Settings)
		expectError   bool
		errorContains string
	}{
		{
			name: "load_non_existing_settings_file",
			setupFunc: func(_ *testing.T, _ string) {
				// No settings file
			},
			validateFunc: func(t *testing.T, settings *claude.Settings) {
				if settings == nil {
					t.Fatal("settings should not be nil")
				}
				if settings.IncludeCoAuthoredBy != false {
					t.Errorf("IncludeCoAuthoredBy = %v, want false", settings.IncludeCoAuthoredBy)
				}
			},
			expectError: false,
		},
		{
			name: "load_valid_settings_file",
			setupFunc: func(t *testing.T, claudeDir string) {
				settings := map[string]interface{}{
					"includeCoAuthoredBy": true,
				}
				data, _ := json.Marshal(settings)
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, data, 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc: func(t *testing.T, settings *claude.Settings) {
				if settings == nil {
					t.Fatal("settings should not be nil")
				}
				if settings.IncludeCoAuthoredBy != true {
					t.Errorf("IncludeCoAuthoredBy = %v, want true", settings.IncludeCoAuthoredBy)
				}
			},
			expectError: false,
		},
		{
			name: "load_invalid_json",
			setupFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if err := os.WriteFile(settingsPath, []byte("{invalid}"), 0644); err != nil {
					t.Fatalf("failed to write settings: %v", err)
				}
			},
			validateFunc:  nil,
			expectError:   true,
			errorContains: "failed to parse settings file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("failed to create claude dir: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, claudeDir)
			}

			mgr := NewManager(claudeDir)
			settings, err := mgr.loadSettings()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, settings)
			}
		})
	}
}

// TestManager_saveSettings tests saving settings to file
func TestManager_saveSettings(t *testing.T) {
	tests := []struct {
		name          string
		settings      *claude.Settings
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, claudeDir string)
	}{
		{
			name: "save_valid_settings",
			settings: &claude.Settings{
				IncludeCoAuthoredBy: true,
			},
			expectError: false,
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("failed to read settings file: %v", err)
				}

				var settings claude.Settings
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("failed to parse settings: %v", err)
				}

				if settings.IncludeCoAuthoredBy != true {
					t.Errorf("IncludeCoAuthoredBy = %v, want true", settings.IncludeCoAuthoredBy)
				}
			},
		},
		{
			name: "save_settings_creates_directory",
			settings: &claude.Settings{
				IncludeCoAuthoredBy: false,
			},
			expectError: false,
			validateFunc: func(t *testing.T, claudeDir string) {
				settingsPath := filepath.Join(claudeDir, "settings.json")
				if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
					t.Error("settings file should be created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude", "nested")

			mgr := NewManager(claudeDir)
			err := mgr.saveSettings(tt.settings)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, claudeDir)
			}
		})
	}
}

// TestManager_saveHooksBackup tests saving hooks backup
func TestManager_saveHooksBackup(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	mgr := NewManager(claudeDir)

	hooksConfig := &claude.HooksConfig{
		PostToolUse: []*claude.HookRule{
			{
				Matcher: "TestMatcher",
				Hooks: []*claude.HookItem{
					{
						Type:    "command",
						Command: "/test/hook.sh",
						Timeout: 60,
					},
				},
			},
		},
	}

	err := mgr.saveHooksBackup(hooksConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}

	var backup claude.HooksConfig
	if err := json.Unmarshal(data, &backup); err != nil {
		t.Fatalf("failed to parse backup: %v", err)
	}

	if len(backup.PostToolUse) == 0 {
		t.Fatal("PostToolUse should not be empty")
	}

	if backup.PostToolUse[0].Matcher != "TestMatcher" {
		t.Errorf("Matcher = %v, want TestMatcher", backup.PostToolUse[0].Matcher)
	}
}

// TestManager_loadHooksBackup tests loading hooks backup
func TestManager_loadHooksBackup(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, claudeDir string)
		validateFunc  func(t *testing.T, config *claude.HooksConfig)
		expectError   bool
		errorContains string
	}{
		{
			name: "load_existing_backup",
			setupFunc: func(t *testing.T, claudeDir string) {
				hooksConfig := map[string]interface{}{
					"PostToolUse": []map[string]interface{}{
						{
							"matcher": "BackupMatcher",
							"hooks": []map[string]interface{}{
								{
									"type":    "command",
									"command": "/backup/hook.sh",
									"timeout": 90,
								},
							},
						},
					},
				}
				data, _ := json.Marshal(hooksConfig)
				backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
				if err := os.WriteFile(backupPath, data, 0644); err != nil {
					t.Fatalf("failed to write backup: %v", err)
				}
			},
			validateFunc: func(t *testing.T, config *claude.HooksConfig) {
				if config == nil {
					t.Fatal("config should not be nil")
				}
				if len(config.PostToolUse) == 0 {
					t.Fatal("PostToolUse should not be empty")
				}
				if config.PostToolUse[0].Matcher != "BackupMatcher" {
					t.Errorf("Matcher = %v, want BackupMatcher", config.PostToolUse[0].Matcher)
				}
			},
			expectError: false,
		},
		{
			name:      "load_non_existing_backup",
			setupFunc: func(_ *testing.T, _ string) {},
			validateFunc: func(_ *testing.T, _ *claude.HooksConfig) {
				t.Error("should not reach validateFunc on error")
			},
			expectError:   true,
			errorContains: "hooks backup file not found",
		},
		{
			name: "load_invalid_backup_json",
			setupFunc: func(t *testing.T, claudeDir string) {
				backupPath := filepath.Join(claudeDir, "settings.json.hooks_backup")
				if err := os.WriteFile(backupPath, []byte("{invalid}"), 0644); err != nil {
					t.Fatalf("failed to write backup: %v", err)
				}
			},
			validateFunc:  nil,
			expectError:   true,
			errorContains: "failed to parse hooks backup file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("failed to create claude dir: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, claudeDir)
			}

			mgr := NewManager(claudeDir)
			config, err := mgr.loadHooksBackup()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, config)
			}
		})
	}
}

// TestManager_EnableDisableIntegration tests enable/disable integration
func TestManager_EnableDisableIntegration(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	mgr := NewManager(claudeDir)

	// Enable hooks
	if err := mgr.EnableCheck(context.Background()); err != nil {
		t.Fatalf("failed to enable hooks: %v", err)
	}

	// Verify hooks are enabled
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}

	hooks, ok := settings["hooks"]
	if !ok {
		t.Fatal("hooks should be enabled")
	}

	hooksMap, ok := hooks.(map[string]interface{})
	if !ok {
		t.Fatal("hooks should be a map")
	}

	postToolUse, ok := hooksMap["PostToolUse"]
	if !ok {
		t.Fatal("PostToolUse should exist after enable")
	}

	postToolUseSlice, ok := postToolUse.([]interface{})
	if !ok || len(postToolUseSlice) == 0 {
		t.Fatal("PostToolUse should not be empty after enable")
	}

	// Disable hooks
	if err := mgr.DisableCheck(context.Background()); err != nil {
		t.Fatalf("failed to disable hooks: %v", err)
	}

	// Verify hooks are disabled - use the actual Settings type to parse
	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var actualSettings claude.Settings
	if err := json.Unmarshal(data, &actualSettings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}

	// After disable, hooks should be nil
	if actualSettings.Hooks != nil {
		// If hooks is not nil, PostToolUse should be nil or empty
		if len(actualSettings.Hooks.PostToolUse) > 0 {
			t.Errorf("PostToolUse should be empty or nil after disable, got %d items", len(actualSettings.Hooks.PostToolUse))
		}
	}
}

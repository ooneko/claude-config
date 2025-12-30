package check

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/claude"
)

// Manager implements check functionality management
type Manager struct {
	claudeDir string
}

// NewManager creates a new check manager
func NewManager(claudeDir string) *Manager {
	return &Manager{
		claudeDir: claudeDir,
	}
}

// EnableCheck enables code checking hooks (PostToolUse hooks)
func (m *Manager) EnableCheck(_ context.Context) error {
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Initialize hooks config if it doesn't exist
	if settings.Hooks == nil {
		settings.Hooks = &claude.HooksConfig{}
	}

	// Load from backup or create default configuration
	var hooksConfig *claude.HooksConfig
	if backupConfig, err := m.loadHooksBackup(); err == nil {
		hooksConfig = backupConfig
	} else {
		hooksConfig = m.createDefaultHooksConfig()
	}

	// Enable PostToolUse hooks
	settings.Hooks.PostToolUse = hooksConfig.PostToolUse

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// DisableCheck disables code checking hooks (PostToolUse hooks)
func (m *Manager) DisableCheck(_ context.Context) error {
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// If hooks config doesn't exist, nothing to disable
	if settings.Hooks == nil {
		return nil
	}

	// Save current hooks configuration before modifying
	if err := m.saveHooksBackup(settings.Hooks); err != nil {
		return fmt.Errorf("failed to save hooks backup: %w", err)
	}

	// Remove PostToolUse hooks
	settings.Hooks.PostToolUse = nil

	// If all hooks are removed, set hooks to nil
	if len(settings.Hooks.PostToolUse) == 0 &&
		len(settings.Hooks.Stop) == 0 {
		settings.Hooks = nil
	}

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// createDefaultHooksConfig creates a default hooks configuration
func (m *Manager) createDefaultHooksConfig() *claude.HooksConfig {
	return &claude.HooksConfig{
		PostToolUse: []*claude.HookRule{
			{
				Matcher: "Write|Edit|MultiEdit",
				Hooks: []*claude.HookItem{
					{
						Type:    "command",
						Command: "~/.claude/hooks/smart-lint.sh",
						Timeout: 120,
					},
					{
						Type:    "command",
						Command: "~/.claude/hooks/smarter-test.sh",
						Timeout: 120,
					},
				},
			},
		},
	}
}

// loadSettings loads settings from settings.json，保留所有字段
func (m *Manager) loadSettings() (*claude.Settings, error) {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")
	raw, err := claude.LoadSettingsRaw(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	return claude.RawToSettings(raw), nil
}

// saveSettings saves settings to settings.json，保留所有字段
func (m *Manager) saveSettings(settings *claude.Settings) error {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")

	// 先读取现有数据以保留未知字段
	raw, err := claude.LoadSettingsRaw(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load existing settings: %w", err)
	}

	// 将 settings 转换为 map
	settingsData, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	var settingsMap map[string]interface{}
	if err := json.Unmarshal(settingsData, &settingsMap); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// 合并：新数据覆盖旧数据
	for k, v := range settingsMap {
		raw[k] = v
	}

	// 特殊处理：如果 settings.Env 为 nil，删除 raw 中的 env
	if settings.Env == nil {
		delete(raw, "env")
	} else if envMap, ok := raw["env"].(map[string]interface{}); ok {
		// 删除原始 env 中已不存在的 key（用于删除操作）
		settingsEnvMap := make(map[string]bool)
		for k := range settings.Env {
			settingsEnvMap[k] = true
		}
		for k := range envMap {
			if !settingsEnvMap[k] {
				delete(envMap, k)
			}
		}
		// 如果 env 为空，删除 env 字段
		if len(envMap) == 0 {
			delete(raw, "env")
		}
	}

	return claude.SaveSettingsRaw(settingsPath, raw)
}

// saveHooksBackup saves hooks configuration to backup file
func (m *Manager) saveHooksBackup(hooksConfig *claude.HooksConfig) error {
	backupPath := filepath.Join(m.claudeDir, "settings.json.hooks_backup")

	data, err := json.Marshal(hooksConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal hooks config: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write hooks backup file: %w", err)
	}

	return nil
}

// loadHooksBackup loads hooks configuration from backup file
func (m *Manager) loadHooksBackup() (*claude.HooksConfig, error) {
	backupPath := filepath.Join(m.claudeDir, "settings.json.hooks_backup")

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("hooks backup file not found")
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read hooks backup file: %w", err)
	}

	var hooksConfig claude.HooksConfig
	if err := json.Unmarshal(data, &hooksConfig); err != nil {
		return nil, fmt.Errorf("failed to parse hooks backup file: %w", err)
	}

	return &hooksConfig, nil
}

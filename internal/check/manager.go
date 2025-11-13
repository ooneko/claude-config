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

// loadSettings loads settings from settings.json
func (m *Manager) loadSettings() (*claude.Settings, error) {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")

	// If file doesn't exist, return default settings
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return &claude.Settings{
			IncludeCoAuthoredBy: false,
		}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings claude.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return &settings, nil
}

// saveSettings saves settings to settings.json
func (m *Manager) saveSettings(settings *claude.Settings) error {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")

	// Ensure directory exists
	if err := os.MkdirAll(m.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
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

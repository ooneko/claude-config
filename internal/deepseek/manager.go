package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/claude"
)

// Manager implements the DeepSeekManager interface
type Manager struct {
	claudeDir string
}

// NewManager creates a new DeepSeek manager
func NewManager(claudeDir string) *Manager {
	return &Manager{
		claudeDir: claudeDir,
	}
}

// Enable enables DeepSeek configuration with the given API key
func (m *Manager) Enable(ctx context.Context, apiKey string) error {
	// Save API key to secure file
	if err := m.saveAPIKey(apiKey); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Initialize env map if it doesn't exist
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Set DeepSeek configuration
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	settings.Env["ANTHROPIC_BASE_URL"] = "https://api.deepseek.com/anthropic"
	settings.Env["ANTHROPIC_MODEL"] = "deepseek-chat"
	settings.Env["ANTHROPIC_SMALL_FAST_MODEL"] = "deepseek-chat"

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Disable disables DeepSeek configuration (keeps API key)
func (m *Manager) Disable(ctx context.Context) error {
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env != nil {
		// Remove DeepSeek environment variables
		delete(settings.Env, "ANTHROPIC_AUTH_TOKEN")
		delete(settings.Env, "ANTHROPIC_BASE_URL")
		delete(settings.Env, "ANTHROPIC_MODEL")
		delete(settings.Env, "ANTHROPIC_SMALL_FAST_MODEL")

		// If env map is empty, set it to nil
		if len(settings.Env) == 0 {
			settings.Env = nil
		}
	}

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Reset removes the API key and disables DeepSeek
func (m *Manager) Reset(ctx context.Context) error {
	// First disable DeepSeek
	if err := m.Disable(ctx); err != nil {
		return fmt.Errorf("failed to disable DeepSeek: %w", err)
	}

	// Remove API key file
	apiKeyPath := filepath.Join(m.claudeDir, ".deepseek_api_key")
	if err := os.Remove(apiKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove API key file: %w", err)
	}

	return nil
}

// IsEnabled returns whether DeepSeek is currently enabled
func (m *Manager) IsEnabled(ctx context.Context) (bool, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return false, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env == nil {
		return false, nil
	}

	authToken := settings.Env["ANTHROPIC_AUTH_TOKEN"]
	baseURL := settings.Env["ANTHROPIC_BASE_URL"]

	return authToken != "" && baseURL != "", nil
}

// HasAPIKey returns whether an API key is stored
func (m *Manager) HasAPIKey(ctx context.Context) (bool, error) {
	apiKeyPath := filepath.Join(m.claudeDir, ".deepseek_api_key")
	_, err := os.Stat(apiKeyPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check API key file: %w", err)
	}
	return true, nil
}

// GetConfig returns current DeepSeek configuration
func (m *Manager) GetConfig(ctx context.Context) (*claude.DeepSeekConfig, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env == nil {
		return nil, nil
	}

	authToken := settings.Env["ANTHROPIC_AUTH_TOKEN"]
	baseURL := settings.Env["ANTHROPIC_BASE_URL"]

	if authToken == "" && baseURL == "" {
		return nil, nil
	}

	return &claude.DeepSeekConfig{
		AuthToken:      authToken,
		BaseURL:        baseURL,
		Model:          settings.Env["ANTHROPIC_MODEL"],
		SmallFastModel: settings.Env["ANTHROPIC_SMALL_FAST_MODEL"],
	}, nil
}

// saveAPIKey saves API key to a secure file with restricted permissions
func (m *Manager) saveAPIKey(apiKey string) error {
	apiKeyPath := filepath.Join(m.claudeDir, ".deepseek_api_key")

	// Ensure directory exists
	if err := os.MkdirAll(m.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	// Write API key with restricted permissions
	if err := os.WriteFile(apiKeyPath, []byte(apiKey), 0600); err != nil {
		return fmt.Errorf("failed to write API key file: %w", err)
	}

	return nil
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

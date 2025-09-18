package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/claude"
)

// Manager implements the ProxyManager interface
type Manager struct {
	claudeDir string
}

// NewManager creates a new proxy manager
func NewManager(claudeDir string) *Manager {
	return &Manager{
		claudeDir: claudeDir,
	}
}

// Enable enables proxy with the given configuration
func (m *Manager) Enable(ctx context.Context, config *claude.ProxyConfig) error {
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Initialize env map if it doesn't exist
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Set proxy configuration
	settings.Env["http_proxy"] = config.HTTPProxy
	settings.Env["https_proxy"] = config.HTTPSProxy

	// Save proxy configuration to .proxy_config file for future use
	if err := m.saveProxyConfig(config); err != nil {
		return fmt.Errorf("failed to save proxy config: %w", err)
	}

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Disable disables proxy
func (m *Manager) Disable(ctx context.Context) error {
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env != nil {
		delete(settings.Env, "http_proxy")
		delete(settings.Env, "https_proxy")

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

// Toggle toggles proxy state
func (m *Manager) Toggle(ctx context.Context) error {
	enabled, err := m.IsEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to check proxy status: %w", err)
	}

	if enabled {
		return m.Disable(ctx)
	}

	// Load saved proxy configuration
	config, err := m.loadProxyConfig()
	if err != nil {
		// Use default proxy configuration if no saved config
		config = &claude.ProxyConfig{
			HTTPProxy:  "http://127.0.0.1:7890",
			HTTPSProxy: "http://127.0.0.1:7890",
		}
	}

	return m.Enable(ctx, config)
}

// IsEnabled returns whether proxy is currently enabled
func (m *Manager) IsEnabled(ctx context.Context) (bool, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return false, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env == nil {
		return false, nil
	}

	httpProxy := settings.Env["http_proxy"]
	httpsProxy := settings.Env["https_proxy"]

	return httpProxy != "" && httpsProxy != "", nil
}

// GetConfig returns current proxy configuration
func (m *Manager) GetConfig(ctx context.Context) (*claude.ProxyConfig, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env == nil {
		return nil, nil
	}

	httpProxy := settings.Env["http_proxy"]
	httpsProxy := settings.Env["https_proxy"]

	if httpProxy == "" && httpsProxy == "" {
		return nil, nil
	}

	return &claude.ProxyConfig{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
	}, nil
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

// saveProxyConfig saves proxy configuration to .proxy_config file
func (m *Manager) saveProxyConfig(config *claude.ProxyConfig) error {
	configPath := filepath.Join(m.claudeDir, ".proxy_config")

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal proxy config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write proxy config file: %w", err)
	}

	return nil
}

// loadProxyConfig loads proxy configuration from .proxy_config file
func (m *Manager) loadProxyConfig() (*claude.ProxyConfig, error) {
	configPath := filepath.Join(m.claudeDir, ".proxy_config")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("proxy config file not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy config file: %w", err)
	}

	var config claude.ProxyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse proxy config file: %w", err)
	}

	return &config, nil
}

// LoadSavedConfig loads saved proxy configuration from file
func (m *Manager) LoadSavedConfig(ctx context.Context) (*claude.ProxyConfig, error) {
	return m.loadProxyConfig()
}

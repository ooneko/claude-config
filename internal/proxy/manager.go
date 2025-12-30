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
func (m *Manager) Enable(_ context.Context, config *claude.ProxyConfig) error {
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
func (m *Manager) Disable(_ context.Context) error {
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
func (m *Manager) IsEnabled(_ context.Context) (bool, error) {
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
func (m *Manager) GetConfig(_ context.Context) (*claude.ProxyConfig, error) {
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
func (m *Manager) LoadSavedConfig(_ context.Context) (*claude.ProxyConfig, error) {
	return m.loadProxyConfig()
}

// Reset removes saved proxy configuration and disables proxy
func (m *Manager) Reset(ctx context.Context) error {
	// First disable proxy
	if err := m.Disable(ctx); err != nil {
		return fmt.Errorf("failed to disable proxy: %w", err)
	}

	// Remove saved proxy configuration file
	configPath := filepath.Join(m.claudeDir, ".proxy_config")
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("failed to remove proxy config file: %w", err)
		}
	}

	return nil
}

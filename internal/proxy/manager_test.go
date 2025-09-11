package proxy

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

func TestProxyManager_Enable(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings without proxy
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Enable
	proxyConfig := &claude.ProxyConfig{
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
	}

	err = manager.Enable(ctx, proxyConfig)
	require.NoError(t, err)

	// Verify settings were updated
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Equal(t, "http://127.0.0.1:7890", updatedSettings.Env["http_proxy"])
	assert.Equal(t, "http://127.0.0.1:7890", updatedSettings.Env["https_proxy"])

	// Test IsEnabled
	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestProxyManager_Disable(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings with proxy
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"http_proxy":  "http://127.0.0.1:7890",
			"https_proxy": "http://127.0.0.1:7890",
			"OTHER_VAR":   "keep_this",
		},
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Disable
	err = manager.Disable(ctx)
	require.NoError(t, err)

	// Verify proxy settings were removed but other env vars remain
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Empty(t, updatedSettings.Env["http_proxy"])
	assert.Empty(t, updatedSettings.Env["https_proxy"])
	assert.Equal(t, "keep_this", updatedSettings.Env["OTHER_VAR"])

	// Test IsEnabled
	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestProxyManager_Toggle(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create proxy config file
	proxyConfigPath := filepath.Join(claudeDir, ".proxy_config")
	proxyConfig := &claude.ProxyConfig{
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
	}
	configData, err := json.Marshal(proxyConfig)
	require.NoError(t, err)
	err = os.WriteFile(proxyConfigPath, configData, 0644)
	require.NoError(t, err)

	// Create initial settings without proxy
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Toggle (should enable)
	err = manager.Toggle(ctx)
	require.NoError(t, err)

	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.True(t, enabled)

	// Test Toggle again (should disable)
	err = manager.Toggle(ctx)
	require.NoError(t, err)

	enabled, err = manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestProxyManager_GetConfig(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create settings with proxy
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"http_proxy":  "http://192.168.1.100:8080",
			"https_proxy": "http://192.168.1.100:8080",
		},
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test GetConfig
	config, err := manager.GetConfig(ctx)
	require.NoError(t, err)

	assert.Equal(t, "http://192.168.1.100:8080", config.HTTPProxy)
	assert.Equal(t, "http://192.168.1.100:8080", config.HTTPSProxy)
}

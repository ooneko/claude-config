package deepseek

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

func TestDeepSeekManager_Enable(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings without DeepSeek
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
	apiKey := "sk-test123456789"
	err = manager.Enable(ctx, apiKey)
	require.NoError(t, err)

	// Verify settings were updated
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Equal(t, apiKey, updatedSettings.Env["ANTHROPIC_AUTH_TOKEN"])
	assert.Equal(t, "https://api.deepseek.com/anthropic", updatedSettings.Env["ANTHROPIC_BASE_URL"])
	assert.Equal(t, "deepseek-chat", updatedSettings.Env["ANTHROPIC_MODEL"])
	assert.Equal(t, "deepseek-chat", updatedSettings.Env["ANTHROPIC_SMALL_FAST_MODEL"])

	// Verify API key file was created with correct permissions
	apiKeyPath := filepath.Join(claudeDir, ".deepseek_api_key")
	info, err := os.Stat(apiKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Test IsEnabled
	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.True(t, enabled)

	// Test HasAPIKey
	hasKey, err := manager.HasAPIKey(ctx)
	require.NoError(t, err)
	assert.True(t, hasKey)
}

func TestDeepSeekManager_Disable(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings with DeepSeek
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN":       "sk-test123",
			"ANTHROPIC_BASE_URL":         "https://api.deepseek.com/anthropic",
			"ANTHROPIC_MODEL":            "deepseek-chat",
			"ANTHROPIC_SMALL_FAST_MODEL": "deepseek-chat",
			"OTHER_VAR":                  "keep_this",
		},
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create API key file
	apiKeyPath := filepath.Join(claudeDir, ".deepseek_api_key")
	err = os.WriteFile(apiKeyPath, []byte("sk-test123"), 0600)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Disable
	err = manager.Disable(ctx)
	require.NoError(t, err)

	// Verify DeepSeek settings were removed but other env vars remain
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Empty(t, updatedSettings.Env["ANTHROPIC_AUTH_TOKEN"])
	assert.Empty(t, updatedSettings.Env["ANTHROPIC_BASE_URL"])
	assert.Empty(t, updatedSettings.Env["ANTHROPIC_MODEL"])
	assert.Empty(t, updatedSettings.Env["ANTHROPIC_SMALL_FAST_MODEL"])
	assert.Equal(t, "keep_this", updatedSettings.Env["OTHER_VAR"])

	// Verify API key file still exists (disable doesn't remove it)
	_, err = os.Stat(apiKeyPath)
	require.NoError(t, err)

	// Test IsEnabled
	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.False(t, enabled)

	// Test HasAPIKey (should still be true)
	hasKey, err := manager.HasAPIKey(ctx)
	require.NoError(t, err)
	assert.True(t, hasKey)
}

func TestDeepSeekManager_Reset(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings with DeepSeek
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "sk-test123",
			"ANTHROPIC_BASE_URL":   "https://api.deepseek.com/anthropic",
			"OTHER_VAR":            "keep_this",
		},
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create API key file
	apiKeyPath := filepath.Join(claudeDir, ".deepseek_api_key")
	err = os.WriteFile(apiKeyPath, []byte("sk-test123"), 0600)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Reset
	err = manager.Reset(ctx)
	require.NoError(t, err)

	// Verify DeepSeek settings were removed
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Empty(t, updatedSettings.Env["ANTHROPIC_AUTH_TOKEN"])
	assert.Empty(t, updatedSettings.Env["ANTHROPIC_BASE_URL"])
	assert.Equal(t, "keep_this", updatedSettings.Env["OTHER_VAR"])

	// Verify API key file was removed
	_, err = os.Stat(apiKeyPath)
	assert.True(t, os.IsNotExist(err))

	// Test IsEnabled
	enabled, err := manager.IsEnabled(ctx)
	require.NoError(t, err)
	assert.False(t, enabled)

	// Test HasAPIKey
	hasKey, err := manager.HasAPIKey(ctx)
	require.NoError(t, err)
	assert.False(t, hasKey)
}

func TestDeepSeekManager_GetConfig(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create settings with DeepSeek
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN":       "sk-test123456",
			"ANTHROPIC_BASE_URL":         "https://api.deepseek.com/anthropic",
			"ANTHROPIC_MODEL":            "deepseek-chat",
			"ANTHROPIC_SMALL_FAST_MODEL": "deepseek-chat",
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
	require.NotNil(t, config)

	assert.Equal(t, "sk-test123456", config.AuthToken)
	assert.Equal(t, "https://api.deepseek.com/anthropic", config.BaseURL)
	assert.Equal(t, "deepseek-chat", config.Model)
	assert.Equal(t, "deepseek-chat", config.SmallFastModel)
}

func TestDeepSeekManager_EnableWithExistingAPIKey(t *testing.T) {
	// Setup temp directory
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)

	// Create initial settings without DeepSeek
	settings := &claude.Settings{
		IncludeCoAuthoredBy: false,
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Create existing API key file
	apiKeyPath := filepath.Join(claudeDir, ".deepseek_api_key")
	oldAPIKey := "sk-old123"
	err = os.WriteFile(apiKeyPath, []byte(oldAPIKey), 0600)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(claudeDir)
	ctx := context.Background()

	// Test Enable with new API key (should overwrite)
	newAPIKey := "sk-new456"
	err = manager.Enable(ctx, newAPIKey)
	require.NoError(t, err)

	// Verify new API key is saved
	savedKey, err := os.ReadFile(apiKeyPath)
	require.NoError(t, err)
	assert.Equal(t, newAPIKey, string(savedKey))

	// Verify settings use new API key
	updatedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var updatedSettings claude.Settings
	err = json.Unmarshal(updatedData, &updatedSettings)
	require.NoError(t, err)

	assert.Equal(t, newAPIKey, updatedSettings.Env["ANTHROPIC_AUTH_TOKEN"])
}

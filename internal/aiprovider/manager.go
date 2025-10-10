package aiprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/claude"
)

// Manager implements the claude.AIProviderManager interface
type Manager struct {
	claudeDir string
	providers map[ProviderType]Provider
}

// NewManager creates a new AI provider manager
func NewManager(claudeDir string) claude.AIProviderManager {
	m := &Manager{
		claudeDir: claudeDir,
		providers: make(map[ProviderType]Provider),
	}

	// Register supported providers
	m.providers[ProviderDeepSeek] = &DeepSeekProvider{}
	m.providers[ProviderKimi] = &KimiProvider{}
	m.providers[ProviderZhiPu] = &ZhipuProvider{}

	return m
}

// Enable enables an AI provider with the given API key
func (m *Manager) Enable(_ context.Context, provider ProviderType, apiKey string) error {
	if !provider.IsValid() {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Save API key
	if err := m.saveAPIKey(provider, apiKey); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	// Get provider implementation
	providerImpl, exists := m.providers[provider]
	if !exists {
		return fmt.Errorf("provider implementation not found: %s", provider)
	}

	// Get default configuration
	config := providerImpl.GetDefaultConfig(apiKey)

	// Load current settings
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Initialize env map if it doesn't exist
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Set provider configuration
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = config.AuthToken
	settings.Env["ANTHROPIC_BASE_URL"] = config.BaseURL
	settings.Env["ANTHROPIC_MODEL"] = config.Model
	settings.Env["ANTHROPIC_SMALL_FAST_MODEL"] = config.SmallFastModel

	// Save settings
	if err := m.saveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Reset removes the API key and disables the provider
func (m *Manager) Reset(_ context.Context, provider ProviderType) error {
	// First disable the provider by clearing environment variables
	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env != nil {
		// Remove AI provider environment variables
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

	// Remove API key file
	apiKeyPath := m.getAPIKeyPath(provider)
	if err := os.Remove(apiKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove API key file: %w", err)
	}

	return nil
}

// Off disables all AI providers completely
func (m *Manager) Off(ctx context.Context) error {
	// First, save the current active provider for restoration
	if err := m.saveLastActiveProvider(ctx); err != nil {
		// Don't fail the off operation, just log it
		fmt.Printf("警告: 无法保存当前配置用于恢复: %v\n", err)
	}

	settings, err := m.loadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env != nil {
		// Remove all AI provider environment variables
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

// On restores the previously active AI provider
func (m *Manager) On(ctx context.Context) error {
	// Load the last active provider
	lastProvider, err := m.loadLastActiveProvider()
	if err != nil {
		return fmt.Errorf("failed to load last active provider: %w", err)
	}

	if lastProvider == ProviderNone {
		return fmt.Errorf("没有找到之前的AI提供商配置")
	}

	// Check if we have API key for this provider
	hasKey, err := m.HasAPIKey(ctx, lastProvider)
	if err != nil {
		return fmt.Errorf("failed to check API key: %w", err)
	}

	if !hasKey {
		return fmt.Errorf("提供商 %s 的API密钥已丢失，请重新启用", lastProvider)
	}

	// Load the API key
	apiKey, err := m.loadAPIKey(lastProvider)
	if err != nil {
		return fmt.Errorf("failed to load API key: %w", err)
	}

	// Re-enable the provider
	if err := m.Enable(ctx, lastProvider, apiKey); err != nil {
		return fmt.Errorf("failed to restore provider %s: %w", lastProvider, err)
	}

	return nil
}

// HasAPIKey returns whether an API key is stored for the provider
func (m *Manager) HasAPIKey(_ context.Context, provider ProviderType) (bool, error) {
	apiKeyPath := m.getAPIKeyPath(provider)
	_, err := os.Stat(apiKeyPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check API key file: %w", err)
	}
	return true, nil
}

// GetProviderConfig returns current configuration for a provider
func (m *Manager) GetProviderConfig(_ context.Context, provider ProviderType) (*ProviderConfig, error) {
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

	return &ProviderConfig{
		Type:           provider,
		AuthToken:      authToken,
		BaseURL:        baseURL,
		Model:          settings.Env["ANTHROPIC_MODEL"],
		SmallFastModel: settings.Env["ANTHROPIC_SMALL_FAST_MODEL"],
	}, nil
}

// GetActiveProvider returns the currently active provider
func (m *Manager) GetActiveProvider(_ context.Context) (ProviderType, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return ProviderNone, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.Env == nil {
		return ProviderNone, nil
	}

	baseURL := settings.Env["ANTHROPIC_BASE_URL"]

	// Determine provider based on base URL
	for providerType, provider := range m.providers {
		config := provider.GetDefaultConfig("")
		if config.BaseURL == baseURL {
			return providerType, nil
		}
	}

	return ProviderNone, nil
}

// ListSupportedProviders returns all supported provider types
func (m *Manager) ListSupportedProviders() []ProviderType {
	providers := make([]ProviderType, 0, len(m.providers))
	for providerType := range m.providers {
		providers = append(providers, providerType)
	}
	return providers
}

// getAPIKeyPath returns the API key file path for a provider
func (m *Manager) getAPIKeyPath(provider ProviderType) string {
	return filepath.Join(m.claudeDir, fmt.Sprintf(".%s_api_key", provider))
}

// saveAPIKey saves API key to a secure file with restricted permissions
func (m *Manager) saveAPIKey(provider ProviderType, apiKey string) error {
	apiKeyPath := m.getAPIKeyPath(provider)

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

// getLastActiveProviderPath returns the path for storing last active provider
func (m *Manager) getLastActiveProviderPath() string {
	return filepath.Join(m.claudeDir, ".last_active_provider")
}

// saveLastActiveProvider saves the currently active provider
func (m *Manager) saveLastActiveProvider(ctx context.Context) error {
	activeProvider, err := m.GetActiveProvider(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active provider: %w", err)
	}

	if activeProvider == ProviderNone {
		// No active provider to save
		return nil
	}

	lastProviderPath := m.getLastActiveProviderPath()

	// Ensure directory exists
	if err := os.MkdirAll(m.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	// Write last active provider
	if err := os.WriteFile(lastProviderPath, []byte(string(activeProvider)), 0644); err != nil {
		return fmt.Errorf("failed to write last active provider file: %w", err)
	}

	return nil
}

// loadLastActiveProvider loads the last active provider
func (m *Manager) loadLastActiveProvider() (ProviderType, error) {
	lastProviderPath := m.getLastActiveProviderPath()

	// If file doesn't exist, return ProviderNone
	if _, err := os.Stat(lastProviderPath); os.IsNotExist(err) {
		return ProviderNone, nil
	}

	data, err := os.ReadFile(lastProviderPath)
	if err != nil {
		return ProviderNone, fmt.Errorf("failed to read last active provider file: %w", err)
	}

	providerType := ProviderType(string(data))
	if !providerType.IsValid() {
		return ProviderNone, fmt.Errorf("invalid provider type: %s", providerType)
	}

	return providerType, nil
}

// loadAPIKey loads API key from file
func (m *Manager) loadAPIKey(provider ProviderType) (string, error) {
	apiKeyPath := m.getAPIKeyPath(provider)

	data, err := os.ReadFile(apiKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read API key file: %w", err)
	}

	return string(data), nil
}

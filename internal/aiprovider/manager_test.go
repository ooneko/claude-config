package aiprovider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooneko/claude-config/internal/claude"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name      string
		claudeDir string
		want      int // number of providers
	}{
		{
			name:      "create manager with valid directory",
			claudeDir: "/tmp/test-claude",
			want:      4, // DeepSeek, Kimi, ZhiPu, Doubao
		},
		{
			name:      "create manager with empty directory",
			claudeDir: "",
			want:      4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewManager(tt.claudeDir)

			if got == nil {
				t.Error("NewManager() returned nil")
				return
			}

			mgr, ok := got.(*Manager)
			if !ok {
				t.Error("NewManager() did not return *Manager")
				return
			}

			if mgr.claudeDir != tt.claudeDir {
				t.Errorf("NewManager() claudeDir = %v, want %v", mgr.claudeDir, tt.claudeDir)
			}

			if len(mgr.providers) != tt.want {
				t.Errorf("NewManager() providers count = %v, want %v", len(mgr.providers), tt.want)
			}

			// Check all expected providers are registered
			expectedProviders := []ProviderType{ProviderDeepSeek, ProviderKimi, ProviderGLM, ProviderDoubao}
			for _, provider := range expectedProviders {
				if _, exists := mgr.providers[provider]; !exists {
					t.Errorf("NewManager() missing provider %v", provider)
				}
			}
		})
	}
}

func TestManager_Enable(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "enable deepseek with valid key",
			provider: ProviderDeepSeek,
			apiKey:   "sk-deepseek-test-key",
			wantErr:  false,
		},
		{
			name:     "enable kimi with valid key",
			provider: ProviderKimi,
			apiKey:   "sk-kimi-test-key",
			wantErr:  false,
		},
		{
			name:     "enable glm with valid key",
			provider: ProviderGLM,
			apiKey:   "sk-glm-test-key",
			wantErr:  false,
		},
		{
			name:     "enable with empty api key",
			provider: ProviderDeepSeek,
			apiKey:   "",
			wantErr:  true,
		},
		{
			name:     "enable with invalid provider",
			provider: ProviderType("invalid"),
			apiKey:   "test-key",
			wantErr:  true,
		},
		{
			name:     "enable with none provider",
			provider: ProviderNone,
			apiKey:   "test-key",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for each test
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			err := mgr.Enable(ctx, tt.provider, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Enable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify API key was saved
				hasKey, err := mgr.HasAPIKey(ctx, tt.provider)
				if err != nil {
					t.Errorf("Manager.HasAPIKey() error = %v", err)
				}
				if !hasKey {
					t.Error("API key should be saved after successful enable")
				}

				// Verify provider is active
				activeProvider, err := mgr.GetActiveProvider(ctx)
				if err != nil {
					t.Errorf("Manager.GetActiveProvider() error = %v", err)
				}
				if activeProvider != tt.provider {
					t.Errorf("Active provider = %v, want %v", activeProvider, tt.provider)
				}

				// Verify configuration is set
				config, err := mgr.GetProviderConfig(ctx, tt.provider)
				if err != nil {
					t.Errorf("Manager.GetProviderConfig() error = %v", err)
				}
				if config == nil {
					t.Error("Provider config should not be nil after enable")
				} else if config.AuthToken != tt.apiKey {
					t.Errorf("Auth token = %v, want %v", config.AuthToken, tt.apiKey)
				}
			}
		})
	}
}

func TestManager_Reset(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		setup    bool // whether to enable provider first
		wantErr  bool
	}{
		{
			name:     "reset enabled provider",
			provider: ProviderDeepSeek,
			setup:    true,
			wantErr:  false,
		},
		{
			name:     "reset non-enabled provider",
			provider: ProviderKimi,
			setup:    false,
			wantErr:  false,
		},
		{
			name:     "reset invalid provider",
			provider: ProviderType("invalid"),
			setup:    false,
			wantErr:  false, // Reset doesn't validate provider type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			if tt.setup {
				err := mgr.Enable(ctx, tt.provider, "test-key")
				if err != nil {
					t.Fatalf("Setup enable failed: %v", err)
				}
			}

			err := mgr.Reset(ctx, tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Reset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.setup {
				// Verify API key was removed
				hasKey, err := mgr.HasAPIKey(ctx, tt.provider)
				if err != nil {
					t.Errorf("Manager.HasAPIKey() error = %v", err)
				}
				if hasKey {
					t.Error("API key should be removed after reset")
				}

				// Verify provider is no longer active
				activeProvider, err := mgr.GetActiveProvider(ctx)
				if err != nil {
					t.Errorf("Manager.GetActiveProvider() error = %v", err)
				}
				if activeProvider != ProviderNone {
					t.Errorf("Active provider should be None after reset, got %v", activeProvider)
				}

				// Verify configuration is cleared
				config, err := mgr.GetProviderConfig(ctx, tt.provider)
				if err != nil {
					t.Errorf("Manager.GetProviderConfig() error = %v", err)
				}
				if config != nil {
					t.Error("Provider config should be nil after reset")
				}
			}
		})
	}
}

func TestManager_Off(t *testing.T) {
	tests := []struct {
		name    string
		setup   []ProviderType // providers to enable before off
		wantErr bool
	}{
		{
			name:    "turn off with single provider enabled",
			setup:   []ProviderType{ProviderDeepSeek},
			wantErr: false,
		},
		{
			name:    "turn off with multiple providers enabled",
			setup:   []ProviderType{ProviderDeepSeek, ProviderKimi},
			wantErr: false,
		},
		{
			name:    "turn off with no providers enabled",
			setup:   []ProviderType{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			// Setup providers
			for i, provider := range tt.setup {
				err := mgr.Enable(ctx, provider, "test-key-"+string(rune(i+'0')))
				if err != nil {
					t.Fatalf("Setup enable failed for %v: %v", provider, err)
				}
			}

			err := mgr.Off(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Off() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify no provider is active
				activeProvider, err := mgr.GetActiveProvider(ctx)
				if err != nil {
					t.Errorf("Manager.GetActiveProvider() error = %v", err)
				}
				if activeProvider != ProviderNone {
					t.Errorf("Active provider should be None after off, got %v", activeProvider)
				}

				// Verify API keys are still preserved
				for _, provider := range tt.setup {
					hasKey, err := mgr.HasAPIKey(ctx, provider)
					if err != nil {
						t.Errorf("Manager.HasAPIKey() error = %v", err)
					}
					if !hasKey {
						t.Errorf("API key should be preserved after off for %v", provider)
					}
				}

				// Verify last active provider was saved if any provider was enabled
				if len(tt.setup) > 0 {
					lastProvider, err := mgr.loadLastActiveProvider()
					if err != nil {
						t.Errorf("loadLastActiveProvider() error = %v", err)
					}
					// Should save the last enabled provider
					expectedLast := tt.setup[len(tt.setup)-1]
					if lastProvider != expectedLast {
						t.Errorf("Last active provider = %v, want %v", lastProvider, expectedLast)
					}
				}
			}
		})
	}
}

func TestManager_On(t *testing.T) {
	tests := []struct {
		name         string
		setupOff     bool
		lastProvider ProviderType
		hasAPIKey    bool
		wantErr      bool
	}{
		{
			name:         "turn on with valid last provider",
			setupOff:     true,
			lastProvider: ProviderKimi,
			hasAPIKey:    true,
			wantErr:      false,
		},
		{
			name:         "turn on with no last provider",
			setupOff:     false,
			lastProvider: ProviderNone,
			hasAPIKey:    false,
			wantErr:      true,
		},
		{
			name:         "turn on with missing API key",
			setupOff:     true,
			lastProvider: ProviderGLM,
			hasAPIKey:    false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			if tt.setupOff && tt.lastProvider != ProviderNone {
				// Setup: enable provider, then turn off
				err := mgr.Enable(ctx, tt.lastProvider, "test-key")
				if err != nil {
					t.Fatalf("Setup enable failed: %v", err)
				}

				err = mgr.Off(ctx)
				if err != nil {
					t.Fatalf("Setup off failed: %v", err)
				}

				// Remove API key if test requires missing key
				if !tt.hasAPIKey {
					apiKeyPath := mgr.getAPIKeyPath(tt.lastProvider)
					os.Remove(apiKeyPath)
				}
			}

			err := mgr.On(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.On() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify provider is active again
				activeProvider, err := mgr.GetActiveProvider(ctx)
				if err != nil {
					t.Errorf("Manager.GetActiveProvider() error = %v", err)
				}
				if activeProvider != tt.lastProvider {
					t.Errorf("Active provider = %v, want %v", activeProvider, tt.lastProvider)
				}
			}
		})
	}
}

func TestManager_HasAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		setup    bool // whether to save API key first
		want     bool
		wantErr  bool
	}{
		{
			name:     "has api key for existing key",
			provider: ProviderDeepSeek,
			setup:    true,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "no api key for non-existing key",
			provider: ProviderKimi,
			setup:    false,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "check invalid provider",
			provider: ProviderType("invalid"),
			setup:    false,
			want:     false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			if tt.setup {
				err := mgr.saveAPIKey(tt.provider, "test-key")
				if err != nil {
					t.Fatalf("Setup saveAPIKey failed: %v", err)
				}
			}

			got, err := mgr.HasAPIKey(ctx, tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.HasAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.HasAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetProviderConfig(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		setup    bool // whether to enable provider first
		wantNil  bool
		wantErr  bool
	}{
		{
			name:     "get config for enabled provider",
			provider: ProviderDeepSeek,
			setup:    true,
			wantNil:  false,
			wantErr:  false,
		},
		{
			name:     "get config for non-enabled provider",
			provider: ProviderKimi,
			setup:    false,
			wantNil:  true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			if tt.setup {
				err := mgr.Enable(ctx, tt.provider, "test-key")
				if err != nil {
					t.Fatalf("Setup enable failed: %v", err)
				}
			}

			got, err := mgr.GetProviderConfig(ctx, tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetProviderConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNil && got != nil {
				t.Error("Manager.GetProviderConfig() should return nil for non-enabled provider")
			}
			if !tt.wantNil && got == nil {
				t.Error("Manager.GetProviderConfig() should not return nil for enabled provider")
			}

			if !tt.wantNil && got != nil {
				if got.Type != tt.provider {
					t.Errorf("Provider type = %v, want %v", got.Type, tt.provider)
				}
				if got.AuthToken != "test-key" {
					t.Errorf("Auth token = %v, want %v", got.AuthToken, "test-key")
				}
			}
		})
	}
}

func TestManager_GetActiveProvider(t *testing.T) {
	tests := []struct {
		name    string
		setup   []ProviderType // providers to enable (last one will be active)
		want    ProviderType
		wantErr bool
	}{
		{
			name:    "no active provider",
			setup:   []ProviderType{},
			want:    ProviderNone,
			wantErr: false,
		},
		{
			name:    "single active provider",
			setup:   []ProviderType{ProviderDeepSeek},
			want:    ProviderDeepSeek,
			wantErr: false,
		},
		{
			name:    "multiple providers - last one active",
			setup:   []ProviderType{ProviderDeepSeek, ProviderKimi},
			want:    ProviderKimi,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)
			ctx := context.Background()

			// Setup providers
			for i, provider := range tt.setup {
				err := mgr.Enable(ctx, provider, "test-key-"+string(rune(i+'0')))
				if err != nil {
					t.Fatalf("Setup enable failed for %v: %v", provider, err)
				}
			}

			got, err := mgr.GetActiveProvider(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetActiveProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.GetActiveProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_ListSupportedProviders(t *testing.T) {
	mgr := NewManager("/tmp/test").(*Manager)

	providers := mgr.ListSupportedProviders()

	expectedProviders := []ProviderType{ProviderDeepSeek, ProviderKimi, ProviderGLM, ProviderDoubao}
	if len(providers) != len(expectedProviders) {
		t.Errorf("ListSupportedProviders() returned %d providers, want %d", len(providers), len(expectedProviders))
	}

	// Check all expected providers are present
	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %v not found in list", expected)
		}
	}
}

func TestManager_loadSettings(t *testing.T) {
	tests := []struct {
		name         string
		setupFile    bool
		fileContent  string
		wantErr      bool
		wantDefaults bool
	}{
		{
			name:         "load non-existing settings file",
			setupFile:    false,
			wantErr:      false,
			wantDefaults: true,
		},
		{
			name:        "load valid settings file",
			setupFile:   true,
			fileContent: `{"includeCoAuthoredBy": true, "env": {"TEST": "value"}}`,
			wantErr:     false,
		},
		{
			name:        "load invalid json",
			setupFile:   true,
			fileContent: `{"invalid": json}`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)

			if tt.setupFile {
				settingsPath := filepath.Join(tmpDir, "settings.json")
				err := os.WriteFile(settingsPath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Setup file write failed: %v", err)
				}
			}

			got, err := mgr.loadSettings()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.loadSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("loadSettings() returned nil settings")
					return
				}

				if tt.wantDefaults {
					if got.IncludeCoAuthoredBy != false {
						t.Error("Default settings should have IncludeCoAuthoredBy = false")
					}
				}
			}
		})
	}
}

func TestManager_saveSettings(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir).(*Manager)

	settings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"TEST_KEY": "test_value",
		},
	}

	err := mgr.saveSettings(settings)
	if err != nil {
		t.Errorf("Manager.saveSettings() error = %v", err)
		return
	}

	// Verify file was created
	settingsPath := filepath.Join(tmpDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file was not created")
		return
	}

	// Verify content by loading it back
	loadedSettings, err := mgr.loadSettings()
	if err != nil {
		t.Errorf("Loading saved settings failed: %v", err)
		return
	}

	if loadedSettings.IncludeCoAuthoredBy != settings.IncludeCoAuthoredBy {
		t.Errorf("IncludeCoAuthoredBy = %v, want %v", loadedSettings.IncludeCoAuthoredBy, settings.IncludeCoAuthoredBy)
	}

	if loadedSettings.Env["TEST_KEY"] != settings.Env["TEST_KEY"] {
		t.Errorf("Env[TEST_KEY] = %v, want %v", loadedSettings.Env["TEST_KEY"], settings.Env["TEST_KEY"])
	}
}

func TestManager_saveAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "save valid api key",
			provider: ProviderDeepSeek,
			apiKey:   "sk-test-key-123",
			wantErr:  false,
		},
		{
			name:     "save empty api key",
			provider: ProviderKimi,
			apiKey:   "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)

			err := mgr.saveAPIKey(tt.provider, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.saveAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created with correct content
				apiKeyPath := mgr.getAPIKeyPath(tt.provider)
				content, err := os.ReadFile(apiKeyPath)
				if err != nil {
					t.Errorf("Reading saved API key failed: %v", err)
					return
				}

				if string(content) != tt.apiKey {
					t.Errorf("Saved API key = %v, want %v", string(content), tt.apiKey)
				}

				// Verify file permissions
				info, err := os.Stat(apiKeyPath)
				if err != nil {
					t.Errorf("Stat API key file failed: %v", err)
					return
				}

				if info.Mode().Perm() != 0600 {
					t.Errorf("API key file permissions = %v, want %v", info.Mode().Perm(), 0600)
				}
			}
		})
	}
}

func TestManager_loadAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		setup    bool
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "load existing api key",
			provider: ProviderDeepSeek,
			setup:    true,
			apiKey:   "sk-test-key",
			wantErr:  false,
		},
		{
			name:     "load non-existing api key",
			provider: ProviderKimi,
			setup:    false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mgr := NewManager(tmpDir).(*Manager)

			if tt.setup {
				err := mgr.saveAPIKey(tt.provider, tt.apiKey)
				if err != nil {
					t.Fatalf("Setup saveAPIKey failed: %v", err)
				}
			}

			got, err := mgr.loadAPIKey(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.loadAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.apiKey {
				t.Errorf("Manager.loadAPIKey() = %v, want %v", got, tt.apiKey)
			}
		})
	}
}

package aiprovider

import (
	"context"
	"testing"
)

func TestAIProviderManager_Enable(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "enable deepseek",
			provider: ProviderDeepSeek,
			apiKey:   "test-deepseek-key",
			wantErr:  false,
		},
		{
			name:     "enable kimi",
			provider: ProviderKimi,
			apiKey:   "test-kimi-key",
			wantErr:  false,
		},
		{
			name:     "enable glm4.5",
			provider: ProviderZhiPu,
			apiKey:   "test-glm-key",
			wantErr:  false,
		},
		{
			name:     "invalid provider",
			provider: ProviderType("invalid"),
			apiKey:   "test-key",
			wantErr:  true,
		},
		{
			name:     "empty api key",
			provider: ProviderDeepSeek,
			apiKey:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mgr := NewManager("/tmp/test-claude")

			err := mgr.Enable(ctx, tt.provider, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAIProviderManager_GetProviderConfig(t *testing.T) {
	ctx := context.Background()
	mgr := NewManager("/tmp/test-claude")

	// Test getting config for enabled provider
	err := mgr.Enable(ctx, ProviderKimi, "test-kimi-key")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	config, err := mgr.GetProviderConfig(ctx, ProviderKimi)
	if err != nil {
		t.Errorf("GetProviderConfig() error = %v", err)
	}

	if config == nil {
		t.Error("Config should not be nil for enabled provider")
		return
	}

	if config.AuthToken != "test-kimi-key" {
		t.Errorf("Expected auth token 'test-kimi-key', got '%s'", config.AuthToken)
	}
}

func TestAIProviderManager_GetActiveProvider(t *testing.T) {
	ctx := context.Background()
	mgr := NewManager("/tmp/test-claude-active")

	// Clean up any existing state
	for _, provider := range mgr.ListSupportedProviders() {
		_ = mgr.Reset(ctx, provider)
	}

	// No active provider initially
	provider, err := mgr.GetActiveProvider(ctx)
	if err != nil {
		t.Errorf("GetActiveProvider() error = %v", err)
	}
	if provider != ProviderNone {
		t.Errorf("Expected ProviderNone, got %v", provider)
	}

	// Enable a provider
	err = mgr.Enable(ctx, ProviderZhiPu, "test-glm-key")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	// Should return the active provider
	provider, err = mgr.GetActiveProvider(ctx)
	if err != nil {
		t.Errorf("GetActiveProvider() error = %v", err)
	}
	if provider != ProviderZhiPu {
		t.Errorf("Expected ProviderZhiPu, got %v", provider)
	}
}

func TestAIProviderManager_ListSupportedProviders(t *testing.T) {
	mgr := NewManager("/tmp/test-claude")

	providers := mgr.ListSupportedProviders()
	expectedProviders := []ProviderType{ProviderDeepSeek, ProviderKimi, ProviderZhiPu}

	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

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

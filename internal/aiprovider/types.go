package aiprovider

import (
	"context"

	"github.com/ooneko/claude-config/internal/claude"
)

// Type aliases for convenience
type ProviderType = claude.ProviderType
type ProviderConfig = claude.ProviderConfig

// Provider type constants
const (
	ProviderNone     = claude.ProviderNone
	ProviderDeepSeek = claude.ProviderDeepSeek
	ProviderKimi     = claude.ProviderKimi
	ProviderZhiPu    = claude.ProviderZhiPu
)

// ProviderManager defines the interface for managing AI providers
type ProviderManager interface {
	// Enable enables an AI provider with the given API key
	Enable(ctx context.Context, provider ProviderType, apiKey string) error

	// Disable disables an AI provider (keeps API key)
	Disable(ctx context.Context, provider ProviderType) error

	// Reset removes the API key and disables the provider
	Reset(ctx context.Context, provider ProviderType) error

	// IsEnabled returns whether a provider is currently enabled
	IsEnabled(ctx context.Context, provider ProviderType) (bool, error)

	// HasAPIKey returns whether an API key is stored for the provider
	HasAPIKey(ctx context.Context, provider ProviderType) (bool, error)

	// GetProviderConfig returns current configuration for a provider
	GetProviderConfig(ctx context.Context, provider ProviderType) (*ProviderConfig, error)

	// GetActiveProvider returns the currently active provider
	GetActiveProvider(ctx context.Context) (ProviderType, error)

	// ListSupportedProviders returns all supported provider types
	ListSupportedProviders() []ProviderType
}

// Provider defines the interface for individual AI provider implementations
type Provider interface {
	// GetType returns the provider type
	GetType() ProviderType

	// GetDefaultConfig returns the default configuration for this provider
	GetDefaultConfig(apiKey string) *ProviderConfig

	// ValidateConfig validates the provider configuration
	ValidateConfig(config *ProviderConfig) error
}

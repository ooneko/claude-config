package claude

import "context"

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads the current configuration from settings.json
	Load(ctx context.Context) (*Settings, error)

	// Save saves the configuration to settings.json
	Save(ctx context.Context, config *Settings) error

	// GetStatus returns current configuration status
	GetStatus(ctx context.Context) (*ConfigStatus, error)

	// Backup creates a backup of configuration
	Backup(ctx context.Context) (*BackupInfo, error)
}

// ProxyManager defines the interface for proxy management
type ProxyManager interface {
	// Enable enables proxy with the given configuration
	Enable(ctx context.Context, config *ProxyConfig) error

	// Disable disables proxy
	Disable(ctx context.Context) error

	// Toggle toggles proxy state
	Toggle(ctx context.Context) error

	// IsEnabled returns whether proxy is currently enabled
	IsEnabled(ctx context.Context) (bool, error)

	// GetConfig returns current proxy configuration
	GetConfig(ctx context.Context) (*ProxyConfig, error)

	// LoadSavedConfig loads saved proxy configuration from file
	LoadSavedConfig(ctx context.Context) (*ProxyConfig, error)
}

// AIProviderManager defines the interface for managing multiple AI providers
type AIProviderManager interface {
	// Enable enables an AI provider with the given API key
	Enable(ctx context.Context, provider ProviderType, apiKey string) error

	// Reset removes the API key and disables the provider
	Reset(ctx context.Context, provider ProviderType) error

	// Off disables all AI providers completely
	Off(ctx context.Context) error

	// On restores the previously active AI provider
	On(ctx context.Context) error

	// HasAPIKey returns whether an API key is stored for the provider
	HasAPIKey(ctx context.Context, provider ProviderType) (bool, error)

	// GetProviderConfig returns current configuration for a provider
	GetProviderConfig(ctx context.Context, provider ProviderType) (*ProviderConfig, error)

	// GetActiveProvider returns the currently active provider
	GetActiveProvider(ctx context.Context) (ProviderType, error)

	// ListSupportedProviders returns all supported provider types
	ListSupportedProviders() []ProviderType
}

// FileOperations defines the interface for file operations
type FileOperations interface {
	// Copy copies configuration files to Claude directory
	Copy(ctx context.Context, options *CopyOptions) error

	// Compare compares source and destination files
	Compare(ctx context.Context, sourcePath, destPath string) (*CompareResult, error)

	// MergeSettings intelligently merges settings.json files
	MergeSettings(ctx context.Context, source, dest *Settings) (*Settings, error)
}

// CopyOptions represents options for copy operations
type CopyOptions struct {
	Agents   bool `json:"agents"`
	Commands bool `json:"commands"`
	Hooks    bool `json:"hooks"`
	All      bool `json:"all"`
}

// CompareResult represents the result of file comparison
type CompareResult struct {
	Same        bool     `json:"same"`
	Differences []string `json:"differences,omitempty"`
}

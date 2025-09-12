package claude

import "context"

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads the current configuration from settings.json
	Load(ctx context.Context) (*Settings, error)

	// Save saves the configuration to settings.json
	Save(ctx context.Context, config *Settings) error

	// GetStatus returns the current configuration status
	GetStatus(ctx context.Context) (*ConfigStatus, error)

	// Backup creates a backup of the current configuration
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
}

// DeepSeekManager defines the interface for DeepSeek API management
type DeepSeekManager interface {
	// Enable enables DeepSeek configuration with the given API key
	Enable(ctx context.Context, apiKey string) error

	// Disable disables DeepSeek configuration (keeps API key)
	Disable(ctx context.Context) error

	// Reset removes the API key and disables DeepSeek
	Reset(ctx context.Context) error

	// IsEnabled returns whether DeepSeek is currently enabled
	IsEnabled(ctx context.Context) (bool, error)

	// HasAPIKey returns whether an API key is stored
	HasAPIKey(ctx context.Context) (bool, error)

	// GetConfig returns current DeepSeek configuration
	GetConfig(ctx context.Context) (*DeepSeekConfig, error)
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

// BackupManager defines the interface for backup operations
type BackupManager interface {
	// Create creates a backup of the Claude configuration directory
	Create(ctx context.Context) (*BackupInfo, error)

	// List lists available backup files
	List(ctx context.Context) ([]*BackupInfo, error)

	// Restore restores from a backup file
	Restore(ctx context.Context, backupPath string) error
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

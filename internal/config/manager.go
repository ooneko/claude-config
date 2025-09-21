package config

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ooneko/claude-config/internal/claude"
)

// Manager implements the ConfigManager interface
type Manager struct {
	claudeDir string
}

// NewManager creates a new configuration manager
func NewManager(claudeDir string) *Manager {
	return &Manager{
		claudeDir: claudeDir,
	}
}

// Load loads the current configuration from settings.json
func (m *Manager) Load(ctx context.Context) (*claude.Settings, error) {
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

// Save saves the configuration to settings.json
func (m *Manager) Save(ctx context.Context, config *claude.Settings) error {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")

	// Ensure directory exists
	if err := os.MkdirAll(m.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// GetStatus returns current configuration status
func (m *Manager) GetStatus(ctx context.Context) (*claude.ConfigStatus, error) {
	settingsPath := filepath.Join(m.claudeDir, "settings.json")

	status := &claude.ConfigStatus{
		ConfigPath: settingsPath,
	}

	// Check if config file exists
	if stat, err := os.Stat(settingsPath); err == nil {
		status.ConfigExists = true
		status.LastModified = stat.ModTime().Format(time.RFC3339)
	} else if os.IsNotExist(err) {
		status.ConfigExists = false
	} else {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// Check hooks configuration and other settings
	if status.ConfigExists {
		settings, err := m.Load(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}

		status.HooksConfigured = settings.Hooks != nil && len(settings.Hooks.PostToolUse) > 0
		status.HooksEnabled = status.HooksConfigured
		status.ProxyEnabled = settings.Env != nil && (settings.Env["http_proxy"] != "" || settings.Env["https_proxy"] != "")

		// Set proxy config if enabled
		if status.ProxyEnabled {
			status.ProxyConfig = &claude.ProxyConfig{
				HTTPProxy:  settings.Env["http_proxy"],
				HTTPSProxy: settings.Env["https_proxy"],
			}
		}

		// Check if DeepSeek is enabled by looking for DeepSeek-specific configuration
		status.DeepSeekEnabled = settings.Env["ANTHROPIC_AUTH_TOKEN"] != "" &&
			settings.Env["ANTHROPIC_BASE_URL"] != "" &&
			strings.Contains(settings.Env["ANTHROPIC_BASE_URL"], "deepseek")
	}

	return status, nil
}

// Backup creates a backup of configuration
func (m *Manager) Backup(ctx context.Context) (*claude.BackupInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("claude-config-backup-%s.tar.gz", timestamp)
	backupPath := filepath.Join(homeDir, filename)

	// Create tar.gz archive of claude directory
	if err := m.createTarGzArchive(m.claudeDir, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup archive: %w", err)
	}

	// Get file size
	stat, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file stats: %w", err)
	}

	return &claude.BackupInfo{
		Filename:    filename,
		FilePath:    backupPath,
		ContentType: "directory",
		Size:        stat.Size(),
		Timestamp:   time.Now(),
	}, nil
}

// createTarGzArchive creates a tar.gz archive of the source directory
func (m *Manager) createTarGzArchive(sourceDir, destPath string) error {
	// Create destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through source directory
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for tar header
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		// Skip if it's the source directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a regular file, copy its content
		if info.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}

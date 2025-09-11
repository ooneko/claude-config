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

// GetStatus returns the current configuration status
func (m *Manager) GetStatus(ctx context.Context) (*claude.ConfigStatus, error) {
	settings, err := m.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	status := &claude.ConfigStatus{}

	// Check proxy status
	if settings.Env != nil {
		httpProxy := settings.Env["http_proxy"]
		httpsProxy := settings.Env["https_proxy"]

		if httpProxy != "" && httpsProxy != "" {
			status.ProxyEnabled = true
			status.ProxyConfig = &claude.ProxyConfig{
				HTTPProxy:  httpProxy,
				HTTPSProxy: httpsProxy,
			}
		}

		// Check DeepSeek status
		authToken := settings.Env["ANTHROPIC_AUTH_TOKEN"]
		baseURL := settings.Env["ANTHROPIC_BASE_URL"]

		if authToken != "" && baseURL != "" {
			status.DeepSeekEnabled = true
			status.DeepSeekConfig = &claude.DeepSeekConfig{
				AuthToken:      authToken,
				BaseURL:        baseURL,
				Model:          settings.Env["ANTHROPIC_MODEL"],
				SmallFastModel: settings.Env["ANTHROPIC_SMALL_FAST_MODEL"],
			}
		}
	}

	// Check hooks status
	if settings.Hooks != nil && (len(settings.Hooks.PostToolUse) > 0 || len(settings.Hooks.Stop) > 0) {
		status.HooksEnabled = true
		status.HooksConfig = settings.Hooks
	}

	// Get backup files
	backupFiles, err := m.getBackupFiles()
	if err == nil {
		status.BackupFiles = backupFiles
	}

	return status, nil
}

// Backup creates a backup of the entire Claude configuration directory
func (m *Manager) Backup(ctx context.Context) (*claude.BackupInfo, error) {
	// Get home directory for backup location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable for testing
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			return nil, fmt.Errorf("failed to determine home directory: %w", err)
		}
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("claude-config-backup-%s.tar.gz", timestamp)
	backupPath := filepath.Join(homeDir, backupFilename)

	// Create tar.gz archive
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add all files from claude directory to tar archive
	err = m.addDirectoryToTar(tarWriter, m.claudeDir, ".claude")
	if err != nil {
		os.Remove(backupPath) // Clean up on error
		return nil, fmt.Errorf("failed to add directory to tar: %w", err)
	}

	// Close writers to ensure all data is written
	tarWriter.Close()
	gzipWriter.Close()
	file.Close()

	// Get file info
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return &claude.BackupInfo{
		Filename:    backupFilename,
		FilePath:    backupPath,
		Timestamp:   info.ModTime(),
		Size:        info.Size(),
		ContentType: "directory",
	}, nil
}

// addDirectoryToTar recursively adds directory contents to tar archive
func (m *Manager) addDirectoryToTar(tarWriter *tar.Writer, sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", filePath, err)
		}

		// Set the name in archive (relative path from targetDir)
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		header.Name = filepath.Join(targetDir, relPath)

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a regular file, write its content
		if info.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}
		}

		return nil
	})
}

// getBackupFiles returns a list of available backup files
func (m *Manager) getBackupFiles() ([]*claude.BackupInfo, error) {
	files, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read claude directory: %w", err)
	}

	var backupFiles []*claude.BackupInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, "settings.json.backup.") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			backupFiles = append(backupFiles, &claude.BackupInfo{
				Filename:  name,
				Timestamp: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}

	return backupFiles, nil
}

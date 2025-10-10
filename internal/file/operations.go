package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/claude"
)

// Operations implements the FileOperations interface
type Operations struct {
	sourceDir string
	claudeDir string
	merger    *SettingsJSONMerger
}

// NewOperations creates a new file operations manager
func NewOperations(sourceDir, claudeDir string) *Operations {
	return &Operations{
		sourceDir: sourceDir,
		claudeDir: claudeDir,
		merger:    NewSettingsJSONMerger(),
	}
}

// Copy copies configuration files to Claude directory based on options
func (o *Operations) Copy(ctx context.Context, options *claude.CopyOptions) error {
	if options == nil {
		options = &claude.CopyOptions{All: true}
	}

	// Ensure target directory exists
	if err := os.MkdirAll(o.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	var copyTargets []string

	if options.All {
		copyTargets = []string{"agents", "commands", "hooks", "output-styles", "settings.json", "CLAUDE.md.to.copy", "statusline.js"}
	} else {
		if options.Agents {
			copyTargets = append(copyTargets, "agents")
		}
		if options.Commands {
			copyTargets = append(copyTargets, "commands")
		}
		if options.Hooks {
			copyTargets = append(copyTargets, "hooks")
		}
	}

	// Always process settings.json specially
	if err := o.handleSettingsJSON(ctx); err != nil {
		return fmt.Errorf("failed to handle settings.json: %w", err)
	}

	// Copy other items
	for _, target := range copyTargets {
		if target == "settings.json" {
			continue // Already handled
		}

		sourcePath := filepath.Join(o.sourceDir, target)
		var destPath string

		// Special handling for CLAUDE.md.to.copy
		if target == "CLAUDE.md.to.copy" {
			destPath = filepath.Join(o.claudeDir, "CLAUDE.md")
		} else {
			destPath = filepath.Join(o.claudeDir, target)
		}

		if err := o.copyItem(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", target, err)
		}
	}

	return nil
}

// handleSettingsJSON handles intelligent merging of settings.json
func (o *Operations) handleSettingsJSON(_ context.Context) error {
	sourcePath := filepath.Join(o.sourceDir, "settings.json")
	destPath := filepath.Join(o.claudeDir, "settings.json")

	// Check if source settings exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil // No source settings to merge
	}

	// Load source settings
	sourceSettings, err := o.loadSettings(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load source settings: %w", err)
	}

	// Load destination settings (if exists)
	var destSettings *claude.Settings
	if _, err := os.Stat(destPath); err == nil {
		destSettings, err = o.loadSettings(destPath)
		if err != nil {
			return fmt.Errorf("failed to load destination settings: %w", err)
		}
	}

	// Merge settings
	mergedSettings, err := o.merger.MergeSettings(destSettings, sourceSettings)
	if err != nil {
		return fmt.Errorf("failed to merge settings: %w", err)
	}

	// Save merged settings
	if err := o.saveSettings(destPath, mergedSettings); err != nil {
		return fmt.Errorf("failed to save merged settings: %w", err)
	}

	return nil
}

// copyItem copies a file or directory recursively
func (o *Operations) copyItem(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Source doesn't exist, skip
		}
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if srcInfo.IsDir() {
		return o.copyDirectory(src, dest)
	}

	return o.copyFile(src, dest)
}

// copyFile copies a single file
func (o *Operations) copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	buffer := make([]byte, 32*1024)
	for {
		n, err := sourceFile.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("failed to read source file: %w", err)
		}

		_, err = destFile.Write(buffer[:n])
		if err != nil {
			return fmt.Errorf("failed to write destination file: %w", err)
		}

		if err != nil && err.Error() == "EOF" {
			break
		}
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	return os.Chmod(dest, srcInfo.Mode())
}

// copyDirectory copies a directory recursively
func (o *Operations) copyDirectory(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if err := o.copyItem(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// loadSettings loads settings from a JSON file
func (o *Operations) loadSettings(path string) (*claude.Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	settings := &claude.Settings{}
	if err := settings.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return settings, nil
}

// saveSettings saves settings to a JSON file
func (o *Operations) saveSettings(path string, settings *claude.Settings) error {
	data, err := settings.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// Compare compares source and destination files
func (o *Operations) Compare(_ context.Context, sourcePath, destPath string) (*claude.CompareResult, error) {
	// Check if both files exist
	sourceInfo, sourceErr := os.Stat(sourcePath)
	destInfo, destErr := os.Stat(destPath)

	if os.IsNotExist(sourceErr) && os.IsNotExist(destErr) {
		return &claude.CompareResult{Same: true}, nil
	}

	if os.IsNotExist(sourceErr) {
		return &claude.CompareResult{
			Same:        false,
			Differences: []string{"Source file does not exist"},
		}, nil
	}

	if os.IsNotExist(destErr) {
		return &claude.CompareResult{
			Same:        false,
			Differences: []string{"Destination file does not exist"},
		}, nil
	}

	// Compare file sizes
	if sourceInfo.Size() != destInfo.Size() {
		return &claude.CompareResult{
			Same: false,
			Differences: []string{fmt.Sprintf("File sizes differ: source=%d, dest=%d",
				sourceInfo.Size(), destInfo.Size())},
		}, nil
	}

	// Compare file contents
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	destData, err := os.ReadFile(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read destination file: %w", err)
	}

	// Simple byte comparison
	same := string(sourceData) == string(destData)
	if same {
		return &claude.CompareResult{Same: true}, nil
	}

	return &claude.CompareResult{
		Same:        false,
		Differences: []string{"File contents differ"},
	}, nil
}

// MergeSettings provides direct access to settings merging
func (o *Operations) MergeSettings(_ context.Context, source, dest *claude.Settings) (*claude.Settings, error) {
	return o.merger.MergeSettings(dest, source)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/aiprovider"
	"github.com/ooneko/claude-config/internal/check"
	"github.com/ooneko/claude-config/internal/claude"
	"github.com/ooneko/claude-config/internal/config"
	"github.com/ooneko/claude-config/internal/proxy"
)

var (
	claudeDir string

	// Managers
	configMgr     claude.ConfigManager
	proxyMgr      claude.ProxyManager
	checkMgr      *check.Manager
	aiProviderMgr claude.AIProviderManager
)

func init() {
	// Get default claude directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		claudeDir = ".claude"
	} else {
		claudeDir = filepath.Join(homeDir, ".claude")
	}

	// Initialize managers
	configMgr = config.NewManager(claudeDir)
	proxyMgr = proxy.NewManager(claudeDir)
	checkMgr = check.NewManager(claudeDir)
	aiProviderMgr = aiprovider.NewManager(claudeDir)
}

func main() {
	rootCmd := createRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

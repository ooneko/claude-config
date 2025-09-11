package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ooneko/claude-config/internal/claude"
)

// enableProxy enables proxy with default or custom configuration
func enableProxy() error {
	ctx := context.Background()

	// Use default proxy configuration
	proxyConfig := &claude.ProxyConfig{
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
	}

	err := proxyMgr.Enable(ctx, proxyConfig)
	if err != nil {
		return fmt.Errorf("启用代理失败: %w", err)
	}

	fmt.Printf("✅ 代理已启用：%s\n", proxyConfig.HTTPProxy)
	return nil
}

// createProxyCmd creates the proxy command and subcommands
func createProxyCmd() *cobra.Command {
	proxyCmd := &cobra.Command{
		Use:   "proxy <command>",
		Short: "代理管理",
		Long:  "管理 HTTP/HTTPS 代理设置 (127.0.0.1:7890)",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	proxyOnCmd := &cobra.Command{
		Use:   "on",
		Short: "启用代理",
		RunE: func(cmd *cobra.Command, args []string) error {
			return enableProxy()
		},
	}

	proxyOffCmd := &cobra.Command{
		Use:   "off",
		Short: "禁用代理",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return proxyMgr.Disable(ctx)
		},
	}

	proxyToggleCmd := &cobra.Command{
		Use:   "toggle",
		Short: "切换代理状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return proxyMgr.Toggle(ctx)
		},
	}

	proxyCmd.AddCommand(proxyOnCmd, proxyOffCmd, proxyToggleCmd)
	return proxyCmd
}

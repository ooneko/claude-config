package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ooneko/claude-config/internal/claude"
)

// enableProxy enables proxy with saved or user-input configuration
func enableProxy() error {
	ctx := context.Background()

	// Try to load saved proxy configuration first
	proxyConfig, err := proxyMgr.LoadSavedConfig(ctx)
	if err != nil {
		// No saved configuration, ask user for input
		proxyConfig, err = promptForProxyConfig()
		if err != nil {
			return fmt.Errorf("获取代理配置失败: %w", err)
		}
	}

	err = proxyMgr.Enable(ctx, proxyConfig)
	if err != nil {
		return fmt.Errorf("启用代理失败: %w", err)
	}

	fmt.Printf("✅ 代理已启用：%s\n", proxyConfig.HTTPProxy)
	return nil
}

// promptForProxyConfig prompts user for proxy configuration
func promptForProxyConfig() (*claude.ProxyConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入HTTP代理地址 (默认: http://127.0.0.1:7890): ")
	httpProxy, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取HTTP代理地址失败: %w", err)
	}
	httpProxy = strings.TrimSpace(httpProxy)
	if httpProxy == "" {
		httpProxy = "http://127.0.0.1:7890"
	}

	fmt.Print("请输入HTTPS代理地址 (默认: 与HTTP代理相同): ")
	httpsProxy, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取HTTPS代理地址失败: %w", err)
	}
	httpsProxy = strings.TrimSpace(httpsProxy)
	if httpsProxy == "" {
		httpsProxy = httpProxy
	}

	return &claude.ProxyConfig{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
	}, nil
}

// createProxyCmd creates the proxy command and subcommands
func createProxyCmd() *cobra.Command {
	proxyCmd := &cobra.Command{
		Use:   "proxy <command>",
		Short: "代理管理",
		Long:  "管理 HTTP/HTTPS 代理设置",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	proxyOnCmd := &cobra.Command{
		Use:   "on",
		Short: "启用代理",
		RunE: func(_ *cobra.Command, _ []string) error {
			return enableProxy()
		},
	}

	proxyOffCmd := &cobra.Command{
		Use:   "off",
		Short: "禁用代理",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			return proxyMgr.Disable(ctx)
		},
	}

	proxyToggleCmd := &cobra.Command{
		Use:   "toggle",
		Short: "切换代理状态",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			return proxyMgr.Toggle(ctx)
		},
	}

	proxyResetCmd := &cobra.Command{
		Use:   "reset",
		Short: "重置代理配置",
		Long:  "删除保存的代理配置文件并禁用代理",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			if err := proxyMgr.Reset(ctx); err != nil {
				return err
			}
			fmt.Println("✅ 代理配置已重置")
			return nil
		},
	}

	proxyStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "显示代理状态",
		Long:  "显示当前代理的启用状态和代理地址",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			isEnabled, err := proxyMgr.IsEnabled(ctx)
			if err != nil {
				return fmt.Errorf("获取代理状态失败: %w", err)
			}

			if isEnabled {
				config, err := proxyMgr.GetConfig(ctx)
				if err != nil {
					return fmt.Errorf("获取代理配置失败: %w", err)
				}
				fmt.Printf("🌐 代理状态: ✅ 已启用 (%s)\n", config.HTTPProxy)
			} else {
				fmt.Println("🌐 代理状态: ❌ 已禁用")
			}

			return nil
		},
	}

	proxyCmd.AddCommand(proxyOnCmd, proxyOffCmd, proxyToggleCmd, proxyResetCmd, proxyStatusCmd)
	return proxyCmd
}

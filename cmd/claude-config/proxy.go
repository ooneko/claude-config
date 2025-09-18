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
		Long:  "管理 HTTP/HTTPS 代理设置 (127.0.0.1:7890)",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
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

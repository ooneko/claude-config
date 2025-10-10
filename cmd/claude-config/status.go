package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// createStatusCmd creates the status command
func createStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "显示当前配置状态",
		Long:  `显示代理、检查功能和通知的当前状态`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return showStatus()
		},
	}

	return statusCmd
}

// showStatus displays the current status of all services
func showStatus() error {
	ctx := context.Background()

	fmt.Println("Claude 配置状态:")
	fmt.Println("================")
	fmt.Println()

	// Check proxy status
	if err := showProxyStatus(ctx); err != nil {
		fmt.Printf("❌ 代理状态检查失败: %v\n", err)
	}
	fmt.Println()

	// Check hooks/check status
	if err := showCheckStatus(ctx); err != nil {
		fmt.Printf("❌ 检查功能状态检查失败: %v\n", err)
	}
	fmt.Println()

	// Check notify status
	if err := showNotifyStatus(ctx); err != nil {
		fmt.Printf("❌ 通知状态检查失败: %v\n", err)
	}
	fmt.Println()

	// Check AI provider status
	showAIProviderStatus()

	return nil
}

// showProxyStatus shows the current proxy status
func showProxyStatus(ctx context.Context) error {
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
}

// showCheckStatus shows the current check/hooks status
func showCheckStatus(ctx context.Context) error {
	isEnabled, err := isCheckEnabled(ctx)
	if err != nil {
		return fmt.Errorf("获取检查功能状态失败: %w", err)
	}

	if isEnabled {
		fmt.Println("🔍 检查功能: ✅ 已启用 ")
	} else {
		fmt.Println("🔍 检查功能: ❌ 已禁用")
	}

	return nil
}

// showNotifyStatus shows the current notify status
func showNotifyStatus(ctx context.Context) error {
	isEnabled, ntfyTopic, err := isNotifyEnabled(ctx)
	if err != nil {
		return fmt.Errorf("获取通知状态失败: %w", err)
	}

	if isEnabled {
		if ntfyTopic != "" {
			fmt.Printf("📱 通知状态: ✅ 已启用 (Topic: %s)\n", ntfyTopic)
		} else {
			fmt.Println("📱 通知状态: ⚠️  hooks已启用但未配置NTFY_TOPIC")
		}
	} else {
		fmt.Println("📱 通知状态: ❌ 已禁用")
	}

	return nil
}

// isCheckEnabled checks if the check functionality is enabled
func isCheckEnabled(ctx context.Context) (bool, error) {
	settings, err := configMgr.Load(ctx)
	if err != nil {
		return false, fmt.Errorf("读取配置失败: %w", err)
	}

	// Check if PostToolUse hooks exist and contain smart-lint.sh or smart-test.sh
	if settings.Hooks == nil || settings.Hooks.PostToolUse == nil {
		return false, nil
	}

	smartLintCommand := "~/.claude/hooks/smart-lint.sh"
	smartTestCommand := "~/.claude/hooks/smart-test.sh"

	for _, rule := range settings.Hooks.PostToolUse {
		if rule.Matcher == "Write|Edit|MultiEdit" {
			hasSmartLint := false
			hasSmartTest := false

			for _, hook := range rule.Hooks {
				if hook.Command == smartLintCommand {
					hasSmartLint = true
				}
				if hook.Command == smartTestCommand {
					hasSmartTest = true
				}
			}

			// Consider enabled if we have at least one of the smart hooks
			if hasSmartLint || hasSmartTest {
				return true, nil
			}
		}
	}

	return false, nil
}

// isNotifyEnabled checks if the notify functionality is enabled
func isNotifyEnabled(ctx context.Context) (enabled bool, ntfyTopic string, err error) {
	settings, err := configMgr.Load(ctx)
	if err != nil {
		return false, "", fmt.Errorf("读取配置失败: %w", err)
	}

	// Get NTFY_TOPIC from env
	if settings.Env != nil {
		ntfyTopic = settings.Env["NTFY_TOPIC"]
	}

	// Check if NTFY hooks exist in Stop hooks
	if settings.Hooks == nil || settings.Hooks.Stop == nil {
		return false, ntfyTopic, nil
	}

	ntfyCommand := "~/.claude/hooks/ntfy-notifier.sh"

	for _, rule := range settings.Hooks.Stop {
		if rule.Matcher == "" { // Empty matcher for stop hooks
			for _, hook := range rule.Hooks {
				if hook.Command == ntfyCommand {
					return true, ntfyTopic, nil
				}
			}
		}
	}

	return false, ntfyTopic, nil
}

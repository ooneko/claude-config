package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// showStatus displays the current configuration status
func showStatus() error {
	ctx := context.Background()
	status, err := configMgr.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("获取状态失败: %w", err)
	}

	fmt.Println("📊 Claude 配置状态：")
	fmt.Println("====================")
	fmt.Println()

	// Proxy status
	fmt.Print("🌐 代理状态：")
	if status.ProxyEnabled {
		fmt.Println(" ✅ 已启用")
		if status.ProxyConfig != nil {
			fmt.Printf("   代理地址：%s\n", status.ProxyConfig.HTTPProxy)
		}
	} else {
		fmt.Println(" ❌ 已禁用")
	}
	fmt.Println()

	// DeepSeek status
	fmt.Print("🤖 DeepSeek 状态：")
	if status.DeepSeekEnabled {
		fmt.Println(" ✅ 已启用")
		if status.DeepSeekConfig != nil {
			// Mask API key for security
			maskedToken := maskAPIKey(status.DeepSeekConfig.AuthToken)
			fmt.Printf("   ANTHROPIC_AUTH_TOKEN: %s\n", maskedToken)
			fmt.Printf("   ANTHROPIC_BASE_URL: %s\n", status.DeepSeekConfig.BaseURL)
		}
	} else {
		fmt.Println(" ❌ 已禁用")
	}
	fmt.Println()

	// Hooks status
	fmt.Print("🪝 Hooks 总体状态：")
	if status.HooksEnabled {
		fmt.Println(" ✅ 已启用")
		fmt.Println()

		// Hook 类型控制
		if status.HooksConfig != nil {
			fmt.Println("Hook 类型控制：")
			if len(status.HooksConfig.PostToolUse) > 0 {
				fmt.Println("  check   : ✅ 检查 hooks (代码检查、测试)")
			} else {
				fmt.Println("  check   : ❌ 检查 hooks (代码检查、测试)")
			}
			if len(status.HooksConfig.Stop) > 0 {
				fmt.Println("  notify  : ✅ 通知 hooks (完成通知)")
			} else {
				fmt.Println("  notify  : ❌ 通知 hooks (完成通知)")
			}
			fmt.Println()
		}

	} else {
		fmt.Println(" ❌ 已禁用")
	}

	return nil
}

// createStatusCmd creates the status command
func createStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "显示当前配置状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStatus()
		},
	}
}

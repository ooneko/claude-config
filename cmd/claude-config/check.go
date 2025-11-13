package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// createCheckCmd creates the check command
func createCheckCmd() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check <on|off>",
		Short: "检查功能控制",
		Long: `检查功能控制 - 管理 lint 和 test 等代码检查 hooks

启用时会添加以下hooks到settings.json:
  - smart-lint.sh (智能代码检查)
  - smart-test.sh (智能测试)

这些hooks会在代码编辑后自动运行，确保代码质量。`,
		Example: `  claude-config check on   # 启用代码检查hooks
  claude-config check off  # 禁用代码检查hooks`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			action := args[0]
			return handleCheckCommand(cmd, action)
		},
		SilenceUsage:  true,  // 防止在错误时自动显示使用帮助
		SilenceErrors: false, // 确保错误被正确返回
	}

	return checkCmd
}

// handleCheckCommand handles the check command
func handleCheckCommand(cmd *cobra.Command, action string) error {
	ctx := context.Background()

	switch action {
	case "on", "enable":
		err := checkMgr.EnableCheck(ctx)
		if err != nil {
			return fmt.Errorf("启用代码检查功能失败: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✅ 代码检查功能已启用")
		fmt.Fprintln(cmd.OutOrStdout(), "   - smart-lint.sh (智能代码检查)")
		fmt.Fprintln(cmd.OutOrStdout(), "   - smart-test.sh (智能测试)")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "这些hooks将在代码编辑后自动运行，确保代码质量。")

	case "off", "disable":
		err := checkMgr.DisableCheck(ctx)
		if err != nil {
			return fmt.Errorf("禁用代码检查功能失败: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "❌ 代码检查功能已禁用")

	default:
		return fmt.Errorf("无效操作: %s\n\n支持的操作: on, off, enable, disable\n使用方法: claude-config check <on|off>", action)
	}

	return nil
}

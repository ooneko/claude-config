package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// enableDeepSeek enables DeepSeek with API key
func enableDeepSeek(args []string) error {
	ctx := context.Background()

	var apiKey string
	if len(args) > 0 {
		apiKey = args[0]
	} else {
		// Check if API key already exists
		hasKey, err := deepSeekMgr.HasAPIKey(ctx)
		if err != nil {
			return fmt.Errorf("检查 API 密钥失败: %w", err)
		}

		if hasKey {
			// Use existing API key
			fmt.Println("使用已保存的 API 密钥")
		} else {
			return fmt.Errorf("请提供 API 密钥: claude-config deepseek on <api-key>")
		}
	}

	err := deepSeekMgr.Enable(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("启用 DeepSeek 失败: %w", err)
	}

	fmt.Println("✅ DeepSeek 已启用")
	return nil
}

// createDeepSeekCmd creates the deepseek command and subcommands
func createDeepSeekCmd() *cobra.Command {
	deepseekCmd := &cobra.Command{
		Use:   "deepseek <command>",
		Short: "DeepSeek API 配置",
		Long:  "管理 DeepSeek API 配置和密钥，支持安全的密钥存储",
		Example: `  claude-config deepseek on <api-key>     # 启用 DeepSeek (提供密钥)
  claude-config deepseek on               # 启用 DeepSeek (使用已保存密钥)
  claude-config deepseek off              # 禁用 DeepSeek
  claude-config deepseek reset            # 重置并清除 API 密钥`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	deepseekOnCmd := &cobra.Command{
		Use:     "on [api-key]",
		Short:   "启用 DeepSeek",
		Long:    "启用 DeepSeek API 配置。如果提供 API 密钥则保存，否则使用已保存的密钥",
		Example: "  claude-config deepseek on sk-xxxxx     # 启用并保存新密钥",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return enableDeepSeek(args)
		},
	}

	deepseekOffCmd := &cobra.Command{
		Use:   "off",
		Short: "禁用 DeepSeek",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return deepSeekMgr.Disable(ctx)
		},
	}

	deepseekResetCmd := &cobra.Command{
		Use:   "reset",
		Short: "重置 DeepSeek (清除 API 密钥)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return deepSeekMgr.Reset(ctx)
		},
	}

	deepseekCmd.AddCommand(deepseekOnCmd, deepseekOffCmd, deepseekResetCmd)
	return deepseekCmd
}

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ooneko/claude-config/internal/aiprovider"
	"github.com/ooneko/claude-config/internal/claude"
	"github.com/spf13/cobra"
)

func createAIProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI提供商配置管理",
		Long:  `管理AI提供商配置，支持DeepSeek、Kimi、GLM4.5等多个提供商。`,
		Run: func(_ *cobra.Command, _ []string) {
			showAIProviderStatus()
		},
	}

	cmd.AddCommand(
		createAIProviderResetCmd(),
		createAIProviderOffCmd(),
		createAIProviderOnCmd(),
		createAIProviderListCmd(),
	)

	return cmd
}

func createAIProviderResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset <provider>",
		Short: "重置AI提供商",
		Long:  `重置指定的AI提供商（删除API密钥和配置）。支持的提供商：deepseek, kimi, zhipu`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			provider := claude.NormalizeProviderName(args[0])

			if provider == claude.ProviderNone {
				fmt.Printf("❌ 不支持的提供商: %s\n", args[0])
				fmt.Println("支持的提供商: deepseek, kimi, zhipu")
				return
			}

			ctx := context.Background()
			err := aiProviderMgr.Reset(ctx, provider)
			if err != nil {
				fmt.Printf("❌ 重置AI提供商失败: %v\n", err)
				return
			}

			fmt.Printf("✅ 成功重置 %s\n", provider)
		},
	}
}

func createAIProviderOffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "off",
		Short: "关闭所有AI提供商",
		Long:  `完全关闭所有AI提供商功能（保留所有API密钥）。`,
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()
			err := aiProviderMgr.Off(ctx)
			if err != nil {
				fmt.Printf("❌ 关闭AI提供商失败: %v\n", err)
				return
			}

			fmt.Println("✅ 已关闭所有AI提供商")
		},
	}
}

func createAIProviderOnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "on [provider]",
		Short: "启用AI提供商",
		Long:  `启用指定的AI提供商，如果未指定则恢复最后一次关闭前配置的AI提供商。支持的提供商：deepseek, kimi, zhipu`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			ctx := context.Background()

			if len(args) == 0 {
				// 恢复之前的配置
				err := aiProviderMgr.On(ctx)
				if err != nil {
					fmt.Printf("❌ 恢复AI提供商失败: %v\n", err)
					return
				}
				fmt.Println("✅ 已恢复AI提供商配置")
				return
			}

			// 启用指定的提供商
			provider := claude.NormalizeProviderName(args[0])

			if provider == claude.ProviderNone {
				fmt.Printf("❌ 不支持的提供商: %s\n", args[0])
				fmt.Println("支持的提供商: deepseek, kimi, zhipu")
				return
			}

			// 检查是否有保存的API密钥
			hasKey, err := aiProviderMgr.HasAPIKey(ctx, provider)
			if err != nil {
				fmt.Printf("❌ 检查API密钥失败: %v\n", err)
				return
			}

			if !hasKey {
				fmt.Printf("⚠️  提供商 %s 的API密钥未配置\n", provider)
				fmt.Printf("请使用以下命令配置API密钥:\n")
				fmt.Printf("  echo 'your-api-key' | claude-config ai on %s\n", provider)
				fmt.Printf("或者:\n")
				fmt.Printf("  claude-config ai on %s\n", provider)
				fmt.Printf("然后输入您的API密钥\n")

				// 尝试从标准输入读取API密钥
				fmt.Printf("\n请输入 %s 的API密钥: ", provider)
				var apiKey string
				if _, err := fmt.Scanln(&apiKey); err != nil {
					fmt.Printf("❌ 读取API密钥失败: %v\n", err)
					return
				}

				if apiKey == "" {
					fmt.Println("❌ API密钥不能为空")
					return
				}

				// 启用提供商
				err = aiProviderMgr.Enable(ctx, provider, apiKey)
				if err != nil {
					fmt.Printf("❌ 启用AI提供商失败: %v\n", err)
					return
				}

				fmt.Printf("✅ 成功配置并启用 %s\n", provider)
				return
			}

			// 有API密钥，直接启用
			// 首先获取保存的API密钥
			apiKey, err := getAPIKeyForProvider(provider)
			if err != nil {
				fmt.Printf("❌ 加载API密钥失败: %v\n", err)
				return
			}

			err = aiProviderMgr.Enable(ctx, provider, apiKey)
			if err != nil {
				fmt.Printf("❌ 启用AI提供商失败: %v\n", err)
				return
			}

			fmt.Printf("✅ 成功启用 %s\n", provider)
		},
	}
}

func createAIProviderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出支持的AI提供商",
		Long:  `列出所有支持的AI提供商及其状态。`,
		Run: func(_ *cobra.Command, _ []string) {
			showAIProviderList()
		},
	}
}

func showAIProviderStatus() {
	ctx := context.Background()

	fmt.Println("🤖 AI提供商状态")
	fmt.Println("================")

	// 获取当前活跃的提供商
	activeProvider, err := aiProviderMgr.GetActiveProvider(ctx)
	if err != nil {
		fmt.Printf("❌ 获取活跃提供商失败: %v\n", err)
		return
	}

	if activeProvider == aiprovider.ProviderNone {
		fmt.Println("📍 当前状态: 未启用任何AI提供商")
	} else {
		fmt.Printf("📍 当前活跃提供商: %s\n", activeProvider)

		// 获取配置信息
		config, err := aiProviderMgr.GetProviderConfig(ctx, activeProvider)
		if err != nil {
			fmt.Printf("❌ 获取配置失败: %v\n", err)
			return
		}

		if config != nil {
			fmt.Printf("   📡 基础URL: %s\n", config.BaseURL)
			fmt.Printf("   🧠 模型: %s\n", config.Model)
			fmt.Printf("   ⚡ 快速模型: %s\n", config.SmallFastModel)
		}
	}

	fmt.Println()
}

// getAPIKeyForProvider 获取指定提供商的API密钥
func getAPIKeyForProvider(provider aiprovider.ProviderType) (string, error) {
	// 通过manager的内部方法获取API密钥，但manager的loadAPIKey是私有的
	// 我们需要通过文件系统直接读取
	claudeDir := getClaudeDir()
	apiKeyPath := filepath.Join(claudeDir, fmt.Sprintf(".%s_api_key", provider))

	data, err := os.ReadFile(apiKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read API key file: %w", err)
	}

	return string(data), nil
}

// getClaudeDir 获取Claude配置目录路径
func getClaudeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude")
}

func showAIProviderList() {
	ctx := context.Background()

	fmt.Println("🤖 支持的AI提供商")
	fmt.Println("==================")

	providers := aiProviderMgr.ListSupportedProviders()
	activeProvider, _ := aiProviderMgr.GetActiveProvider(ctx)

	for _, provider := range providers {
		status := "⚪"
		if provider == activeProvider {
			status = "🟢"
		}

		hasKey, _ := aiProviderMgr.HasAPIKey(ctx, provider)
		keyStatus := ""
		if hasKey {
			keyStatus = " (已保存API密钥)"
		}

		fmt.Printf("%s %s%s\n", status, provider, keyStatus)
	}

	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("🟢 当前活跃提供商")
	fmt.Println("⚪ 可用提供商")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  claude-config ai on [provider]")
	fmt.Println("  claude-config ai reset <provider>")
	fmt.Println("  claude-config ai off")
	fmt.Println("  claude-config ai list")
}

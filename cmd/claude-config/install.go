package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ooneko/claude-config/internal/install"
)

// runInstall executes the install command
func runInstall(cmd *cobra.Command) error {
	ctx := context.Background()

	// 解析命令行参数
	options := install.InstallOptions{}

	allFlag, _ := cmd.Flags().GetBool("all")
	agentsFlag, _ := cmd.Flags().GetBool("agents")
	commandsFlag, _ := cmd.Flags().GetBool("commands")
	hooksFlag, _ := cmd.Flags().GetBool("hooks")
	outputStylesFlag, _ := cmd.Flags().GetBool("output-styles")
	settingsFlag, _ := cmd.Flags().GetBool("settings")
	claudeFlag, _ := cmd.Flags().GetBool("claude")

	// 如果没有指定任何选项，默认安装所有
	if !allFlag && !agentsFlag && !commandsFlag && !hooksFlag &&
		!outputStylesFlag && !settingsFlag && !claudeFlag {
		options.All = true
	} else {
		options.All = allFlag
		options.Agents = agentsFlag
		options.Commands = commandsFlag
		options.Hooks = hooksFlag
		options.OutputStyles = outputStylesFlag
		options.Settings = settingsFlag
		options.Claude = claudeFlag
	}

	// 验证选项
	if err := options.Validate(); err != nil {
		return fmt.Errorf("无效的安装选项: %w", err)
	}

	// 创建安装管理器并执行安装
	installMgr := install.NewManager(claudeDir)

	fmt.Println("🚀 开始安装Claude配置文件...")
	if err := installMgr.Install(ctx, options); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	fmt.Println("✅ 安装完成！")
	fmt.Printf("配置目录：%s\n", claudeDir)

	return nil
}

// createInstallCmd creates the install command
func createInstallCmd() *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "安装配置文件",
		Long: `安装Claude Code配置文件到 ~/.claude 目录

支持选择性安装特定组件:
  --all          安装所有配置文件 (默认)
  --agents       仅安装 agents 
  --commands     仅安装 commands
  --hooks        仅安装 hooks
  --output-styles 仅安装 output-styles
  --settings     仅安装 settings.json
  --claude       仅安装 CLAUDE.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd)
		},
	}

	// Install command flags
	installCmd.Flags().Bool("all", false, "安装所有配置文件")
	installCmd.Flags().Bool("agents", false, "仅安装agents")
	installCmd.Flags().Bool("commands", false, "仅安装commands")
	installCmd.Flags().Bool("hooks", false, "仅安装hooks")
	installCmd.Flags().Bool("output-styles", false, "仅安装output-styles")
	installCmd.Flags().Bool("settings", false, "仅安装settings.json")
	installCmd.Flags().Bool("claude", false, "仅安装CLAUDE.md")

	return installCmd
}

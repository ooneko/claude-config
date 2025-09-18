package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "claude-config",
		Short: "Claude 配置管理工具",
		Long:  `Claude Configuration Tool 是一个统一配置管理工具，整合了配置管理和文件复制功能。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 没有子命令时显示帮助信息
			fmt.Println("欢迎使用 Claude 配置管理工具！")
			fmt.Println()
			_ = cmd.Help()
		},
	}

	initCommands(rootCmd)
	return rootCmd
}

func initCommands(rootCmd *cobra.Command) {
	// 添加所有子命令
	rootCmd.AddCommand(
		createStatusCmd(),
		createProxyCmd(),
		createCheckCmd(),
		createDeepSeekCmd(),
		createNotifyCmd(),
		createInstallCmd(),
		createBackupCmd(),
	)
}

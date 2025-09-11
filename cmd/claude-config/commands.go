package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "claude-config",
		Short: "Claude 配置管理工具",
		Long: `Claude Configuration Tool 是一个统一配置管理工具，整合了配置管理和文件复制功能。

提供以下功能：
  • 代理配置管理 (HTTP/HTTPS)
  • Hooks 系统控制 (智能检查和通知)
  • NTFY 通知配置 (启用/禁用)
  • DeepSeek API 集成
  • 配置文件安装和备份`,
		Run: func(cmd *cobra.Command, args []string) {
			// 没有子命令时显示帮助信息
			fmt.Println("欢迎使用 Claude 配置管理工具！")
			fmt.Println()
			cmd.Help()
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

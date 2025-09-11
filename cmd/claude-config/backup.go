package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// createBackupCmd creates the backup command
func createBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "备份配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			backupInfo, err := configMgr.Backup(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("✅ 配置已备份到：%s\n", backupInfo.FilePath)
			fmt.Printf("   大小：%s\n", formatBytes(backupInfo.Size))
			fmt.Printf("   时间：%s\n", backupInfo.Timestamp.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}

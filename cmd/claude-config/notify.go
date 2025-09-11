package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ooneko/claude-config/internal/claude"
	"github.com/spf13/cobra"
)

// createNotifyCmd creates the notify command
func createNotifyCmd() *cobra.Command {
	notifyCmd := &cobra.Command{
		Use:   "notify",
		Short: "NTFY通知配置管理",
		Long:  `管理NTFY通知配置，支持启用/禁用通知功能`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("使用 'claude-config notify on' 启用通知或 'claude-config notify off' 禁用通知")
			cmd.Help()
		},
	}

	// 添加子命令
	notifyCmd.AddCommand(createNotifyOnCmd())
	notifyCmd.AddCommand(createNotifyOffCmd())

	return notifyCmd
}

// createNotifyOnCmd creates the notify on command
func createNotifyOnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "on",
		Short: "启用NTFY通知",
		Long:  `启用NTFY通知功能，如果未配置NTFY_TOPIC则提示用户输入，并添加通知hooks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return enableNTFY()
		},
	}
}

// createNotifyOffCmd creates the notify off command
func createNotifyOffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "off",
		Short: "禁用NTFY通知",
		Long:  `禁用NTFY通知功能，保留NTFY_TOPIC但移除通知hooks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return disableNTFY()
		},
	}
}

// enableNTFY 启用NTFY通知功能
func enableNTFY() error {
	ctx := context.Background()

	// 读取当前配置
	settings, err := configMgr.Load(ctx)
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 确保env部分存在
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// 检查是否已有NTFY_TOPIC配置，如果没有则提示用户输入
	ntfyTopic := settings.Env["NTFY_TOPIC"]
	if ntfyTopic == "" {
		fmt.Print("请输入NTFY Topic: ")
		fmt.Scanln(&ntfyTopic)

		ntfyTopic = strings.TrimSpace(ntfyTopic)
		if ntfyTopic == "" {
			return fmt.Errorf("NTFY Topic不能为空")
		}

		// 更新配置
		settings.Env["NTFY_TOPIC"] = ntfyTopic
	}

	// 确保hooks配置存在
	if settings.Hooks == nil {
		settings.Hooks = &claude.HooksConfig{}
	}

	// 检查Stop hooks中是否已存在ntfy-notifier.sh
	ntfyCommand := "~/.claude/hooks/ntfy-notifier.sh"
	ntfyExists := false

	for _, rule := range settings.Hooks.Stop {
		if rule.Matcher == "" {
			for _, hook := range rule.Hooks {
				if hook.Command == ntfyCommand {
					ntfyExists = true
					break
				}
			}
		}
		if ntfyExists {
			break
		}
	}

	// 如果ntfy hook不存在，添加它
	if !ntfyExists {
		// 查找空matcher的rule，如果不存在则创建
		var targetRule *claude.HookRule
		for _, rule := range settings.Hooks.Stop {
			if rule.Matcher == "" {
				targetRule = rule
				break
			}
		}

		if targetRule == nil {
			targetRule = &claude.HookRule{
				Matcher: "",
				Hooks:   []*claude.HookItem{},
			}
			settings.Hooks.Stop = append(settings.Hooks.Stop, targetRule)
		}

		// 添加ntfy hook
		ntfyHook := &claude.HookItem{
			Type:    "command",
			Command: ntfyCommand,
		}
		targetRule.Hooks = append(targetRule.Hooks, ntfyHook)
	}

	// 保存配置
	if err := configMgr.Save(ctx, settings); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("✅ NTFY通知已启用！Topic: %s\n", ntfyTopic)
	return nil
}

// disableNTFY 禁用NTFY通知功能
func disableNTFY() error {
	ctx := context.Background()

	// 读取当前配置
	settings, err := configMgr.Load(ctx)
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 检查hooks配置是否存在
	if settings.Hooks == nil || settings.Hooks.Stop == nil {
		fmt.Println("✅ NTFY通知已经是禁用状态")
		return nil
	}

	// 查找并移除ntfy-notifier.sh hook
	ntfyCommand := "~/.claude/hooks/ntfy-notifier.sh"
	removed := false

	for i, rule := range settings.Hooks.Stop {
		if rule.Matcher == "" {
			// 在该rule的hooks中查找并移除ntfy hook
			newHooks := []*claude.HookItem{}
			for _, hook := range rule.Hooks {
				if hook.Command != ntfyCommand {
					newHooks = append(newHooks, hook)
				} else {
					removed = true
				}
			}

			// 如果该rule没有hooks了，移除整个rule
			if len(newHooks) == 0 {
				settings.Hooks.Stop = append(settings.Hooks.Stop[:i], settings.Hooks.Stop[i+1:]...)
			} else {
				rule.Hooks = newHooks
			}
			break
		}
	}

	if !removed {
		fmt.Println("✅ NTFY通知已经是禁用状态")
		return nil
	}

	// 保存配置
	if err := configMgr.Save(ctx, settings); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Println("✅ NTFY通知已禁用（保留NTFY_TOPIC配置）")
	return nil
}

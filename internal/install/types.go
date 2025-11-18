package install

import "fmt"

// Options 安装选项配置
type Options struct {
	All          bool // 安装所有配置文件
	Agents       bool // 仅安装agents
	Commands     bool // 仅安装commands
	Hooks        bool // 仅安装hooks
	OutputStyles bool // 仅安装output-styles
	Settings     bool // 仅安装settings.json
	Claude       bool // 仅安装CLAUDE.md
	Statusline   bool // 仅安装statusline.js
	Force        bool // 强制覆盖已存在的文件
	Delete       bool // 删除目标目录中不在源资源中的文件（需要与Force配合使用）
}

// Validate 验证安装选项
func (opts Options) Validate() error {
	if !opts.All && !opts.Agents && !opts.Commands && !opts.Hooks &&
		!opts.OutputStyles && !opts.Settings && !opts.Claude && !opts.Statusline {
		return fmt.Errorf("必须至少选择一个安装选项")
	}
	return nil
}

// GetSelectedComponents 获取选中的组件列表
func (opts Options) GetSelectedComponents() []string {
	var components []string

	if opts.All {
		return []string{"agents", "commands", "hooks", "output-styles", "settings.json", "CLAUDE.md.template", "statusline.js"}
	}

	if opts.Agents {
		components = append(components, "agents")
	}
	if opts.Commands {
		components = append(components, "commands")
	}
	if opts.Hooks {
		components = append(components, "hooks")
	}
	if opts.OutputStyles {
		components = append(components, "output-styles")
	}
	if opts.Settings {
		components = append(components, "settings.json")
	}
	if opts.Claude {
		components = append(components, "CLAUDE.md.template")
	}
	if opts.Statusline {
		components = append(components, "statusline.js")
	}

	return components
}

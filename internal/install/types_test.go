package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options InstallOptions
		wantErr bool
	}{
		{
			name:    "有效选项 - All",
			options: InstallOptions{All: true},
			wantErr: false,
		},
		{
			name:    "有效选项 - 特定组件",
			options: InstallOptions{Agents: true, Commands: true},
			wantErr: false,
		},
		{
			name:    "有效选项 - 单个组件",
			options: InstallOptions{Settings: true},
			wantErr: false,
		},
		{
			name:    "无效选项 - 全部为false",
			options: InstallOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstallOptions_GetSelectedComponents(t *testing.T) {
	tests := []struct {
		name     string
		options  InstallOptions
		expected []string
	}{
		{
			name:    "All选项",
			options: InstallOptions{All: true},
			expected: []string{
				"agents", "commands", "hooks", "output-styles",
				"settings.json", "CLAUDE.md.template", "statusline.js",
			},
		},
		{
			name:     "仅agents",
			options:  InstallOptions{Agents: true},
			expected: []string{"agents"},
		},
		{
			name:     "agents和commands",
			options:  InstallOptions{Agents: true, Commands: true},
			expected: []string{"agents", "commands"},
		},
		{
			name: "所有单独选项",
			options: InstallOptions{
				Agents: true, Commands: true, Hooks: true,
				OutputStyles: true, Settings: true, Claude: true, Statusline: true,
			},
			expected: []string{
				"agents", "commands", "hooks", "output-styles",
				"settings.json", "CLAUDE.md.template", "statusline.js",
			},
		},
		{
			name:     "仅settings",
			options:  InstallOptions{Settings: true},
			expected: []string{"settings.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.GetSelectedComponents()
			assert.Equal(t, tt.expected, result)
		})
	}
}

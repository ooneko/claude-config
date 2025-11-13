package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name:    "有效选项 - All",
			options: Options{All: true},
			wantErr: false,
		},
		{
			name:    "有效选项 - 特定组件",
			options: Options{Agents: true, Commands: true},
			wantErr: false,
		},
		{
			name:    "有效选项 - 单个组件",
			options: Options{Settings: true},
			wantErr: false,
		},
		{
			name:    "无效选项 - 全部为false",
			options: Options{},
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

func TestOptions_GetSelectedComponents(t *testing.T) {
	tests := []struct {
		name     string
		options  Options
		expected []string
	}{
		{
			name:    "All选项",
			options: Options{All: true},
			expected: []string{
				"agents", "commands", "output-styles",
				"settings.json", "CLAUDE.md.template", "statusline.js",
			},
		},
		{
			name:     "仅agents",
			options:  Options{Agents: true},
			expected: []string{"agents"},
		},
		{
			name:     "agents和commands",
			options:  Options{Agents: true, Commands: true},
			expected: []string{"agents", "commands"},
		},
		{
			name: "所有单独选项",
			options: Options{
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
			options:  Options{Settings: true},
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

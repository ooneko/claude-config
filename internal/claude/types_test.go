package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettings_MarshalJSON(t *testing.T) {
	settings := &Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://127.0.0.1:7890",
			"https_proxy": "http://127.0.0.1:7890",
		},
		Hooks: &HooksConfig{
			PostToolUse: []*HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh",
						},
					},
				},
			},
		},
	}

	data, err := settings.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// 验证JSON格式化正确
	assert.Contains(t, string(data), "{\n")
	assert.Contains(t, string(data), "  \"includeCoAuthoredBy\": true")
}

func TestSettings_UnmarshalJSON(t *testing.T) {
	jsonData := `{
  "includeCoAuthoredBy": true,
  "env": {
    "http_proxy": "http://127.0.0.1:7890",
    "https_proxy": "http://127.0.0.1:7890"
  },
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/smart-lint.sh"
          }
        ]
      }
    ]
  }
}`

	var settings Settings
	err := settings.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	assert.True(t, settings.IncludeCoAuthoredBy)
	assert.Equal(t, "http://127.0.0.1:7890", settings.Env["http_proxy"])
	assert.Equal(t, "http://127.0.0.1:7890", settings.Env["https_proxy"])

	require.NotNil(t, settings.Hooks)
	require.Len(t, settings.Hooks.PostToolUse, 1)
	assert.Equal(t, "Write|Edit", settings.Hooks.PostToolUse[0].Matcher)
	require.Len(t, settings.Hooks.PostToolUse[0].Hooks, 1)
	assert.Equal(t, "command", settings.Hooks.PostToolUse[0].Hooks[0].Type)
	assert.Equal(t, "~/.claude/hooks/smart-lint.sh", settings.Hooks.PostToolUse[0].Hooks[0].Command)
}

func TestNormalizeProviderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProviderType
	}{
		// Case-insensitive matches
		{
			name:     "deepseek lowercase",
			input:    "deepseek",
			expected: ProviderDeepSeek,
		},
		{
			name:     "deepseek uppercase",
			input:    "DEEPSEEK",
			expected: ProviderDeepSeek,
		},
		{
			name:     "deepseek mixed case",
			input:    "DeepSeek",
			expected: ProviderDeepSeek,
		},
		{
			name:     "kimi lowercase",
			input:    "kimi",
			expected: ProviderKimi,
		},
		{
			name:     "kimi uppercase",
			input:    "KIMI",
			expected: ProviderKimi,
		},
		{
			name:     "zhipu lowercase",
			input:    "zhipu",
			expected: ProviderZhiPu,
		},
		{
			name:     "zhipu uppercase",
			input:    "ZHIPU",
			expected: ProviderZhiPu,
		},
		{
			name:     "zhipu with hyphen",
			input:    "zhipu-ai",
			expected: ProviderZhiPu,
		},
		{
			name:     "zhipu with hyphen uppercase",
			input:    "ZHIPU-AI",
			expected: ProviderZhiPu,
		},
		// Backwards compatibility
		{
			name:     "exact match ZhiPu",
			input:    "ZhiPu",
			expected: ProviderZhiPu,
		},
		// Invalid cases
		{
			name:     "invalid provider",
			input:    "invalid",
			expected: ProviderNone,
		},
		{
			name:     "empty string",
			input:    "",
			expected: ProviderNone,
		},
		{
			name:     "partial match",
			input:    "deep",
			expected: ProviderNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeProviderName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProviderType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		expected bool
	}{
		{
			name:     "valid deepseek",
			provider: ProviderDeepSeek,
			expected: true,
		},
		{
			name:     "valid kimi",
			provider: ProviderKimi,
			expected: true,
		},
		{
			name:     "valid zhipu",
			provider: ProviderZhiPu,
			expected: true,
		},
		{
			name:     "invalid none",
			provider: ProviderNone,
			expected: false,
		},
		{
			name:     "invalid custom",
			provider: ProviderType("custom"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

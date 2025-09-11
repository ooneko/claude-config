package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ooneko/claude-config/internal/claude"
)

func TestSettingsJsonMerger_MergeSettings_ProxyProtection(t *testing.T) {
	merger := NewSettingsJsonMerger()

	// User's current settings with proxy
	dest := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"http_proxy":              "http://127.0.0.1:7890",
			"https_proxy":             "http://127.0.0.1:7890",
			"CLAUDE_HOOKS_GO_ENABLED": "false",
		},
	}

	// Source settings with different proxy
	source := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://192.168.1.100:8080",
			"https_proxy": "http://192.168.1.100:8080",
		},
	}

	// Test merge
	result, err := merger.MergeSettings(dest, source)
	require.NoError(t, err)

	// Should use source's IncludeCoAuthoredBy
	assert.True(t, result.IncludeCoAuthoredBy)

	// Should preserve user's proxy settings (proxy protection)
	assert.Equal(t, "http://127.0.0.1:7890", result.Env["http_proxy"])
	assert.Equal(t, "http://127.0.0.1:7890", result.Env["https_proxy"])

	// Should preserve user's other env vars
	assert.Equal(t, "false", result.Env["CLAUDE_HOOKS_GO_ENABLED"])
}

func TestSettingsJsonMerger_MergeSettings_HooksIntelligentMerge(t *testing.T) {
	merger := NewSettingsJsonMerger()

	// User's current settings with custom hook
	dest := &claude.Settings{
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/custom-lint.sh",
						},
					},
				},
			},
		},
	}

	// Source settings with broader matcher and additional hooks
	source := &claude.Settings{
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh",
						},
					},
				},
			},
			Stop: []*claude.HookRule{
				{
					Matcher: "",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh",
						},
					},
				},
			},
		},
	}

	// Test merge
	result, err := merger.MergeSettings(dest, source)
	require.NoError(t, err)

	// Should have merged PostToolUse hooks
	require.NotNil(t, result.Hooks)
	require.Len(t, result.Hooks.PostToolUse, 1)

	postToolUse := result.Hooks.PostToolUse[0]

	// Should use the broader matcher
	assert.Equal(t, "Write|Edit|MultiEdit", postToolUse.Matcher)

	// Should include both hooks without duplicates
	require.Len(t, postToolUse.Hooks, 2)

	// User's custom hook should come first
	assert.Equal(t, "~/.claude/hooks/custom-lint.sh", postToolUse.Hooks[0].Command)
	assert.Equal(t, "~/.claude/hooks/smart-lint.sh", postToolUse.Hooks[1].Command)

	// Should add new Stop hooks
	require.Len(t, result.Hooks.Stop, 1)
	assert.Equal(t, "~/.claude/hooks/ntfy-notifier.sh", result.Hooks.Stop[0].Hooks[0].Command)
}

func TestSettingsJsonMerger_MergeSettings_CompleteScenario(t *testing.T) {
	merger := NewSettingsJsonMerger()

	// This test covers the exact scenario from DESIGN.md

	// User's current settings
	dest := &claude.Settings{
		IncludeCoAuthoredBy: false,
		Env: map[string]string{
			"http_proxy":              "http://127.0.0.1:7890",
			"https_proxy":             "http://127.0.0.1:7890",
			"CLAUDE_HOOKS_GO_ENABLED": "false",
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/custom-lint.sh",
						},
					},
				},
			},
		},
	}

	// Source file settings
	source := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":  "http://192.168.1.100:8080",
			"https_proxy": "http://192.168.1.100:8080",
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh",
						},
					},
				},
			},
			Stop: []*claude.HookRule{
				{
					Matcher: "",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh",
						},
					},
				},
			},
		},
	}

	// Expected result matching DESIGN.md
	expected := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"http_proxy":              "http://127.0.0.1:7890", // User's proxy preserved
			"https_proxy":             "http://127.0.0.1:7890", // User's proxy preserved
			"CLAUDE_HOOKS_GO_ENABLED": "false",                 // User's setting preserved
		},
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit|MultiEdit", // Broader matcher chosen
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/custom-lint.sh", // User's hook first
						},
						{
							Type:    "command",
							Command: "~/.claude/hooks/smart-lint.sh", // Source hook added
						},
					},
				},
			},
			Stop: []*claude.HookRule{
				{
					Matcher: "",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh", // New hook added
						},
					},
				},
			},
		},
	}

	// Test merge
	result, err := merger.MergeSettings(dest, source)
	require.NoError(t, err)

	// Verify the result matches expected
	assert.Equal(t, expected.IncludeCoAuthoredBy, result.IncludeCoAuthoredBy)
	assert.Equal(t, expected.Env, result.Env)

	// Verify hooks structure
	require.NotNil(t, result.Hooks)
	require.Len(t, result.Hooks.PostToolUse, 1)
	require.Len(t, result.Hooks.Stop, 1)

	// Verify PostToolUse hooks
	postToolUse := result.Hooks.PostToolUse[0]
	assert.Equal(t, expected.Hooks.PostToolUse[0].Matcher, postToolUse.Matcher)
	assert.Len(t, postToolUse.Hooks, 2)
	assert.Equal(t, expected.Hooks.PostToolUse[0].Hooks[0].Command, postToolUse.Hooks[0].Command)
	assert.Equal(t, expected.Hooks.PostToolUse[0].Hooks[1].Command, postToolUse.Hooks[1].Command)

	// Verify Stop hooks
	stop := result.Hooks.Stop[0]
	assert.Equal(t, expected.Hooks.Stop[0].Matcher, stop.Matcher)
	assert.Len(t, stop.Hooks, 1)
	assert.Equal(t, expected.Hooks.Stop[0].Hooks[0].Command, stop.Hooks[0].Command)
}

func TestSettingsJsonMerger_MergeSettings_BadCases(t *testing.T) {
	merger := NewSettingsJsonMerger()

	// Test case 1: Prevent duplicate hooks with same command
	dest := &claude.Settings{
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "smart-lint.sh",
						},
					},
				},
			},
		},
	}

	source := &claude.Settings{
		Hooks: &claude.HooksConfig{
			PostToolUse: []*claude.HookRule{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []*claude.HookItem{
						{
							Type:    "command",
							Command: "smart-lint.sh", // Same command
						},
					},
				},
			},
		},
	}

	result, err := merger.MergeSettings(dest, source)
	require.NoError(t, err)

	// Should not have duplicate commands
	require.Len(t, result.Hooks.PostToolUse, 1)
	assert.Len(t, result.Hooks.PostToolUse[0].Hooks, 1)
	assert.Equal(t, "smart-lint.sh", result.Hooks.PostToolUse[0].Hooks[0].Command)
	assert.Equal(t, "Write|Edit|MultiEdit", result.Hooks.PostToolUse[0].Matcher)
}

func TestSettingsJsonMerger_NormalizeMatcherPattern(t *testing.T) {
	merger := NewSettingsJsonMerger()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple pattern",
			input:    "Write|Edit",
			expected: "Edit|Write",
		},
		{
			name:     "with spaces",
			input:    "Write | Edit | MultiEdit",
			expected: "Edit|MultiEdit|Write",
		},
		{
			name:     "with duplicates",
			input:    "Write|Edit|Write",
			expected: "Edit|Write",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single pattern",
			input:    "Write",
			expected: "Write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merger.normalizeMatcherPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSettingsJsonMerger_IsProxyVar(t *testing.T) {
	merger := NewSettingsJsonMerger()

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{
			name:     "http_proxy",
			variable: "http_proxy",
			expected: true,
		},
		{
			name:     "https_proxy",
			variable: "https_proxy",
			expected: true,
		},
		{
			name:     "other variable",
			variable: "CLAUDE_HOOKS_GO_ENABLED",
			expected: false,
		},
		{
			name:     "similar but not proxy",
			variable: "http_proxy_backup",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merger.isProxyVar(tt.variable)
			assert.Equal(t, tt.expected, result)
		})
	}
}

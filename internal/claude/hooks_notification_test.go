package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHooksConfig_NotificationStructure tests that notification configuration
// should be under hooks.Notification, not as a top-level Notification field
func TestHooksConfig_NotificationStructure(t *testing.T) {
	// RED Test: This test defines the CORRECT structure
	// Notification rules should be under hooks.Notification
	correctSettings := &Settings{
		IncludeCoAuthoredBy: false,
		Hooks: &HooksConfig{
			Notification: []*HookRule{
				{
					Matcher: "permission_prompt",
					Hooks: []*HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh notification permission_prompt",
						},
					},
				},
			},
		},
	}

	// Test marshaling to JSON
	data, err := json.MarshalIndent(correctSettings, "", "  ")
	require.NoError(t, err)

	// Verify the JSON contains the correct structure
	var unmarshaled Settings
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	require.NotNil(t, unmarshaled.Hooks)
	assert.Len(t, unmarshaled.Hooks.Notification, 1)
	assert.Equal(t, "permission_prompt", unmarshaled.Hooks.Notification[0].Matcher)
}

// TestHooksConfig_NotificationFileRoundTrip tests saving and loading
// settings with notification under hooks structure
func TestHooksConfig_NotificationFileRoundTrip(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "claude-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create settings with notification under hooks
	originalSettings := &Settings{
		IncludeCoAuthoredBy: false,
		Hooks: &HooksConfig{
			Notification: []*HookRule{
				{
					Matcher: "permission_prompt",
					Hooks: []*HookItem{
						{
							Type:    "command",
							Command: "~/.claude/hooks/ntfy-notifier.sh notification permission_prompt",
						},
					},
				},
			},
		},
	}

	// Save to file
	settingsPath := filepath.Join(tempDir, "settings.json")
	data, err := json.MarshalIndent(originalSettings, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Load from file
	loadedData, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var loadedSettings Settings
	err = json.Unmarshal(loadedData, &loadedSettings)
	require.NoError(t, err)

	// Verify loaded settings match original
	require.NotNil(t, loadedSettings.Hooks)
	assert.Len(t, loadedSettings.Hooks.Notification, 1)
	assert.False(t, loadedSettings.IncludeCoAuthoredBy)

	// Verify notification rules are preserved
	permissionRule := loadedSettings.Hooks.Notification[0]
	assert.Equal(t, "permission_prompt", permissionRule.Matcher)
	assert.Len(t, permissionRule.Hooks, 1)
	assert.Contains(t, permissionRule.Hooks[0].Command, "permission_prompt")
}

// TestHooksConfig_FindNotificationRuleByMatcher tests finding notification rules
func TestHooksConfig_FindNotificationRuleByMatcher(t *testing.T) {
	settings := &Settings{
		Hooks: &HooksConfig{
			Notification: []*HookRule{
				{
					Matcher: "permission_prompt",
					Hooks: []*HookItem{
						{Type: "command", Command: "command1"},
					},
				},
			},
		},
	}

	// Test finding existing rules
	permissionRule := FindHookRuleByMatcher(settings.Hooks.Notification, "permission_prompt")
	require.NotNil(t, permissionRule)
	assert.Equal(t, "permission_prompt", permissionRule.Matcher)

	// Test finding non-existing rule
	nonExistingRule := FindHookRuleByMatcher(settings.Hooks.Notification, "non_existing")
	assert.Nil(t, nonExistingRule)
}

// FindHookRuleByMatcher is a helper function to find hook rules by matcher
func FindHookRuleByMatcher(rules []*HookRule, matcher string) *HookRule {
	for _, rule := range rules {
		if rule.Matcher == matcher {
			return rule
		}
	}
	return nil
}

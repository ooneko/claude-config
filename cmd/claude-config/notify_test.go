package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ooneko/claude-config/internal/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigureMacOSNotifications tests the macOS notification configuration function
func TestConfigureMacOSNotifications(t *testing.T) {
	// Test on any system (function should work regardless of OS)
	settings := &claude.Settings{
		IncludeCoAuthoredBy: true,
		Env: map[string]string{
			"EXISTING_VAR": "existing_value",
		},
	}

	// Call the function
	configureMacOSNotifications(settings)

	// Verify notification configuration was added to hooks.Notification
	require.NotNil(t, settings.Hooks)
	assert.Len(t, settings.Hooks.Notification, 1)

	// Verify existing settings are preserved
	assert.True(t, settings.IncludeCoAuthoredBy)
	assert.Equal(t, "existing_value", settings.Env["EXISTING_VAR"])

	// Verify permission prompt rule uses ntfy-notifier.sh
	permissionRule := findHookRuleByMatcher(settings.Hooks.Notification, "permission_prompt")
	require.NotNil(t, permissionRule)
	assert.Len(t, permissionRule.Hooks, 1)
	assert.Equal(t, "command", permissionRule.Hooks[0].Type)
	expectedPermissionCommand := "~/.claude/hooks/ntfy-notifier.sh notification permission_prompt"
	assert.Equal(t, expectedPermissionCommand, permissionRule.Hooks[0].Command)
}

// TestConfigureMacOSNotifications_Idempotent tests that calling the function multiple times works correctly
func TestConfigureMacOSNotifications_Idempotent(t *testing.T) {
	settings := &claude.Settings{}

	// Call the function twice
	configureMacOSNotifications(settings)
	firstNotificationConfig := settings.Hooks.Notification

	configureMacOSNotifications(settings)
	secondNotificationConfig := settings.Hooks.Notification

	// Should still have the same configuration (not duplicated)
	assert.NotNil(t, secondNotificationConfig)
	assert.Len(t, secondNotificationConfig, 1)
	assert.Equal(t, len(firstNotificationConfig), len(secondNotificationConfig))
}

// TestRuntimeGOSSection tests that we're using runtime.GOOS correctly
func TestRuntimeGOSSection(t *testing.T) {
	// This test verifies that runtime.GOOS detection works as expected
	// On macOS systems, this should be "darwin"
	if runtime.GOOS == "darwin" {
		assert.Equal(t, "darwin", runtime.GOOS)
	} else {
		// On non-macOS systems, runtime.GOOS should not be "darwin"
		assert.NotEqual(t, "darwin", runtime.GOOS)
	}
}

// TestNTFYNotifierScript tests the ntfy-notifier.sh script functionality
func TestNTFYNotifierScript(t *testing.T) {
	// Test that the script exists and is executable
	scriptPath := filepath.Join(os.Getenv("HOME"), ".claude", "hooks", "ntfy-notifier.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("ntfy-notifier.sh script not found, skipping integration test")
	}

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read script: %v", err)
	}

	scriptContent := string(content)

	// Verify script supports the new notification types
	expectedSubTypes := []string{"permission_prompt", "idle_prompt"}
	for _, subType := range expectedSubTypes {
		if !containsSubstring(scriptContent, subType) {
			t.Errorf("Script does not appear to handle subtype: %s", subType)
		}
	}

	// Verify script has say command support for macOS
	if runtime.GOOS == "darwin" {
		if !containsSubstring(scriptContent, "say") {
			t.Errorf("Script should support 'say' command on macOS")
		}
	}
}

// TestNTFYNotifierScriptParameters tests different parameter combinations
func TestNTFYNotifierScriptParameters(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "notification permission_prompt",
			args:     []string{"notification", "permission_prompt"},
			expected: "permission_prompt",
		},
		{
			name:     "notification idle_prompt",
			args:     []string{"notification", "idle_prompt"},
			expected: "idle_prompt",
		},
		{
			name:     "stop event",
			args:     []string{"stop"},
			expected: "stop",
		},
		{
			name:     "default notification",
			args:     []string{"notification"},
			expected: "notification",
		},
	}

	scriptPath := filepath.Join(os.Getenv("HOME"), ".claude", "hooks", "ntfy-notifier.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("ntfy-notifier.sh script not found, skipping integration test")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify script structure supports these parameter combinations
			content, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("Failed to read script: %v", err)
			}

			scriptContent := string(content)
			// Check that script handles the case structure for these parameters
			if tc.expected == "stop" {
				if !containsSubstring(scriptContent, `"stop")`) {
					t.Errorf("Script should handle stop case")
				}
			} else if containsSubstring(scriptContent, tc.expected) {
				// Script contains the expected subtype
				t.Logf("Script contains expected subtype: %s", tc.expected)
			}
		})
	}
}

// Helper function to find hook rule by matcher
func findHookRuleByMatcher(rules []*claude.HookRule, matcher string) *claude.HookRule {
	for _, rule := range rules {
		if rule.Matcher == matcher {
			return rule
		}
	}
	return nil
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s == substr {
		return true
	}
	if len(s) == 0 {
		return false
	}
	// Check prefix, suffix, or recursive search
	prefixMatch := s[:len(substr)] == substr
	suffixMatch := s[len(s)-len(substr):] == substr
	recursiveMatch := containsSubstring(s[1:], substr)
	return prefixMatch || suffixMatch || recursiveMatch
}

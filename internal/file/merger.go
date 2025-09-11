package file

import (
	"fmt"
	"strings"

	"github.com/ooneko/claude-config/internal/claude"
)

// SettingsJsonMerger implements intelligent merging of settings.json files
type SettingsJsonMerger struct{}

// NewSettingsJsonMerger creates a new settings merger
func NewSettingsJsonMerger() *SettingsJsonMerger {
	return &SettingsJsonMerger{}
}

// MergeSettings intelligently merges source settings into destination settings
// Following the design rules:
// 1. Proxy configuration protection: user's proxy settings have priority
// 2. Environment variable merging: preserve existing settings
// 3. Hooks intelligent merging: merge by matcher, avoid duplicates
func (m *SettingsJsonMerger) MergeSettings(dest, source *claude.Settings) (*claude.Settings, error) {
	if dest == nil {
		dest = &claude.Settings{}
	}
	if source == nil {
		return dest, nil
	}

	// Create result by copying destination
	result := &claude.Settings{
		IncludeCoAuthoredBy: source.IncludeCoAuthoredBy, // Use source value
		StatusLine:          dest.StatusLine,            // Keep destination status line
	}

	// Merge environment variables with proxy protection
	result.Env = m.mergeEnvironmentVariables(dest.Env, source.Env)

	// Merge hooks intelligently
	var err error
	result.Hooks, err = m.mergeHooksConfig(dest.Hooks, source.Hooks)
	if err != nil {
		return nil, fmt.Errorf("failed to merge hooks: %w", err)
	}

	return result, nil
}

// mergeEnvironmentVariables merges env vars with proxy configuration protection
func (m *SettingsJsonMerger) mergeEnvironmentVariables(destEnv, sourceEnv map[string]string) map[string]string {
	if destEnv == nil && sourceEnv == nil {
		return nil
	}

	result := make(map[string]string)

	// First, add all destination variables
	for key, value := range destEnv {
		result[key] = value
	}

	// Then, add source variables, but protect proxy settings
	for key, value := range sourceEnv {
		// Proxy configuration protection: skip proxy vars if they exist in destination
		if m.isProxyVar(key) && destEnv != nil && destEnv[key] != "" {
			continue // Keep destination proxy settings
		}
		result[key] = value
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// mergeHooksConfig intelligently merges hooks configurations
func (m *SettingsJsonMerger) mergeHooksConfig(destHooks, sourceHooks *claude.HooksConfig) (*claude.HooksConfig, error) {
	if destHooks == nil && sourceHooks == nil {
		return nil, nil
	}
	if destHooks == nil {
		return sourceHooks, nil
	}
	if sourceHooks == nil {
		return destHooks, nil
	}

	result := &claude.HooksConfig{}

	// Merge PostToolUse hooks
	var err error
	result.PostToolUse, err = m.mergeHookRules(destHooks.PostToolUse, sourceHooks.PostToolUse)
	if err != nil {
		return nil, fmt.Errorf("failed to merge PostToolUse hooks: %w", err)
	}

	// Merge Stop hooks
	result.Stop, err = m.mergeHookRules(destHooks.Stop, sourceHooks.Stop)
	if err != nil {
		return nil, fmt.Errorf("failed to merge Stop hooks: %w", err)
	}

	return result, nil
}

// mergeHookRules merges hook rules by matcher, avoiding duplicates
func (m *SettingsJsonMerger) mergeHookRules(destRules, sourceRules []*claude.HookRule) ([]*claude.HookRule, error) {
	if len(destRules) == 0 && len(sourceRules) == 0 {
		return nil, nil
	}
	if len(destRules) == 0 {
		return sourceRules, nil
	}
	if len(sourceRules) == 0 {
		return destRules, nil
	}

	// Check for overlapping matchers and merge intelligently
	var result []*claude.HookRule
	processedDestRules := make([]bool, len(destRules))

	// Process each source rule
	for _, sourceRule := range sourceRules {
		merged := false
		sourceMatcher := m.normalizeMatcherPattern(sourceRule.Matcher)

		// Look for overlapping destination rules
		for i, destRule := range destRules {
			if processedDestRules[i] {
				continue
			}

			destMatcher := m.normalizeMatcherPattern(destRule.Matcher)

			// Check if matchers overlap
			if m.matchersOverlap(destMatcher, sourceMatcher) {
				// Merge these two rules
				mergedRule, err := m.mergeHookRule(destRule, sourceRule)
				if err != nil {
					return nil, err
				}
				result = append(result, mergedRule)
				processedDestRules[i] = true
				merged = true
				break
			}
		}

		// If no overlap found, add source rule as-is
		if !merged {
			result = append(result, sourceRule)
		}
	}

	// Add any remaining destination rules that weren't merged
	for i, destRule := range destRules {
		if !processedDestRules[i] {
			result = append(result, destRule)
		}
	}

	return result, nil
}

// matchersOverlap checks if two normalized matcher patterns overlap
func (m *SettingsJsonMerger) matchersOverlap(matcher1, matcher2 string) bool {
	if matcher1 == matcher2 {
		return true
	}

	// Check if one is a subset of the other
	parts1 := strings.Split(matcher1, "|")
	parts2 := strings.Split(matcher2, "|")

	// Create sets for easier comparison
	set1 := make(map[string]bool)
	for _, part := range parts1 {
		set1[part] = true
	}

	set2 := make(map[string]bool)
	for _, part := range parts2 {
		set2[part] = true
	}

	// Check if there's any overlap
	for part := range set1 {
		if set2[part] {
			return true
		}
	}

	return false
}

// mergeHookRule merges two hook rules with the same matcher
func (m *SettingsJsonMerger) mergeHookRule(destRule, sourceRule *claude.HookRule) (*claude.HookRule, error) {
	// Use the more comprehensive matcher pattern
	matcher := m.choosePreferredMatcher(destRule.Matcher, sourceRule.Matcher)

	// Merge hooks, avoiding duplicates
	mergedHooks := make([]*claude.HookItem, 0)
	commandsSeen := make(map[string]bool)

	// Add destination hooks first
	for _, hook := range destRule.Hooks {
		if !commandsSeen[hook.Command] {
			mergedHooks = append(mergedHooks, hook)
			commandsSeen[hook.Command] = true
		}
	}

	// Add source hooks, avoiding duplicates
	for _, hook := range sourceRule.Hooks {
		if !commandsSeen[hook.Command] {
			mergedHooks = append(mergedHooks, hook)
			commandsSeen[hook.Command] = true
		}
	}

	return &claude.HookRule{
		Matcher: matcher,
		Hooks:   mergedHooks,
	}, nil
}

// normalizeMatcherPattern normalizes matcher patterns for comparison
func (m *SettingsJsonMerger) normalizeMatcherPattern(matcher string) string {
	if matcher == "" {
		return ""
	}

	// Split by | and sort for consistent comparison
	parts := strings.Split(matcher, "|")

	// Remove duplicates and sort
	seen := make(map[string]bool)
	var unique []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && !seen[part] {
			unique = append(unique, part)
			seen[part] = true
		}
	}

	// Sort for consistent comparison
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			if unique[i] > unique[j] {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	return strings.Join(unique, "|")
}

// choosePreferredMatcher chooses the more comprehensive matcher pattern
func (m *SettingsJsonMerger) choosePreferredMatcher(matcher1, matcher2 string) string {
	norm1 := m.normalizeMatcherPattern(matcher1)
	norm2 := m.normalizeMatcherPattern(matcher2)

	parts1 := strings.Split(norm1, "|")
	parts2 := strings.Split(norm2, "|")

	// Choose the one with more parts (more comprehensive)
	if len(parts2) > len(parts1) {
		return matcher2
	}
	if len(parts1) > len(parts2) {
		return matcher1
	}

	// Same length, choose alphabetically last (arbitrary but consistent)
	if norm2 > norm1 {
		return matcher2
	}
	return matcher1
}

// isProxyVar checks if a variable is a proxy-related variable
func (m *SettingsJsonMerger) isProxyVar(key string) bool {
	return key == "http_proxy" || key == "https_proxy"
}

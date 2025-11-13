package claude

import (
	"encoding/json"
	"strings"
	"time"
)

// Settings represents the main configuration structure for Claude Code
type Settings struct {
	IncludeCoAuthoredBy bool              `json:"includeCoAuthoredBy"`
	Env                 map[string]string `json:"env,omitempty"`
	Hooks               *HooksConfig      `json:"hooks,omitempty"`
	StatusLine          *StatusLineConfig `json:"statusLine,omitempty"`
}

// HooksConfig represents the hooks configuration
type HooksConfig struct {
	PreToolUse   []*HookRule `json:"preToolUse,omitempty"`
	PostToolUse  []*HookRule `json:"PostToolUse,omitempty"`
	Stop         []*HookRule `json:"Stop,omitempty"`
	Notification []*HookRule `json:"Notification,omitempty"`
}

// HookRule represents a single hook rule with matcher and hooks
type HookRule struct {
	Matcher string      `json:"matcher"`
	Hooks   []*HookItem `json:"hooks"`
}

// HookItem represents a single hook command
type HookItem struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"` // Timeout in seconds, 0 means no timeout
}

// StatusLineConfig represents status line configuration
type StatusLineConfig map[string]interface{}

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderNone     ProviderType = ""
	ProviderDeepSeek ProviderType = "deepseek"
	ProviderKimi     ProviderType = "kimi"
	ProviderGLM      ProviderType = "GLM"
	ProviderDoubao   ProviderType = "doubao"
)

// String returns the string representation of ProviderType
func (p ProviderType) String() string {
	return string(p)
}

// IsValid checks if the provider type is valid
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderDeepSeek, ProviderKimi, ProviderGLM, ProviderDoubao:
		return true
	default:
		return false
	}
}

// NormalizeProviderName converts user input to the correct ProviderType
// This allows case-insensitive provider names for better user experience
func NormalizeProviderName(input string) ProviderType {
	switch strings.ToLower(input) {
	case "deepseek":
		return ProviderDeepSeek
	case "kimi":
		return ProviderKimi
	case "glm":
		return ProviderGLM
	case "zhipu", "zhipu-ai": // Backwards compatibility
		return ProviderGLM
	case "doubao":
		return ProviderDoubao
	default:
		// If exact match, return as-is for backwards compatibility
		p := ProviderType(input)
		if p.IsValid() {
			return p
		}
		return ProviderNone
	}
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Type           ProviderType `json:"type"`
	AuthToken      string       `json:"auth_token"`
	BaseURL        string       `json:"base_url"`
	Model          string       `json:"model"`
	SmallFastModel string       `json:"small_fast_model"`
}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
}

// ConfigStatus represents configuration status information
type ConfigStatus struct {
	ConfigExists    bool         `json:"config_exists"`
	ConfigPath      string       `json:"config_path"`
	LastModified    string       `json:"last_modified,omitempty"`
	HooksConfigured bool         `json:"hooks_configured"`
	HooksEnabled    bool         `json:"hooks_enabled"`
	ProxyEnabled    bool         `json:"proxy_enabled"`
	ProxyConfig     *ProxyConfig `json:"proxy_config,omitempty"`
	DeepSeekEnabled bool         `json:"deepseek_enabled"`
}

// BackupInfo represents backup operation result
type BackupInfo struct {
	Filename    string    `json:"filename"`
	FilePath    string    `json:"file_path"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Timestamp   time.Time `json:"timestamp"`
}

// MarshalJSON implements json.Marshaler for Settings
func (s *Settings) MarshalJSON() ([]byte, error) {
	type alias Settings
	return json.MarshalIndent((*alias)(s), "", "  ")
}

// UnmarshalJSON implements json.Unmarshaler for Settings
func (s *Settings) UnmarshalJSON(data []byte) error {
	type alias Settings
	return json.Unmarshal(data, (*alias)(s))
}

package claude

import (
	"encoding/json"
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
	PostToolUse []*HookRule `json:"PostToolUse,omitempty"`
	Stop        []*HookRule `json:"Stop,omitempty"`
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
}

// StatusLineConfig represents status line configuration
type StatusLineConfig map[string]interface{}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
}

// DeepSeekConfig represents DeepSeek API configuration
type DeepSeekConfig struct {
	AuthToken      string `json:"auth_token"`
	BaseURL        string `json:"base_url"`
	Model          string `json:"model"`
	SmallFastModel string `json:"small_fast_model"`
}

// BackupInfo represents backup file information
type BackupInfo struct {
	Filename    string    `json:"filename"`
	FilePath    string    `json:"filepath"` // 完整的备份文件路径
	Timestamp   time.Time `json:"timestamp"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"` // "directory" 或 "settings"
}

// ConfigStatus represents the current configuration status
type ConfigStatus struct {
	ProxyEnabled    bool            `json:"proxy_enabled"`
	ProxyConfig     *ProxyConfig    `json:"proxy_config,omitempty"`
	DeepSeekEnabled bool            `json:"deepseek_enabled"`
	DeepSeekConfig  *DeepSeekConfig `json:"deepseek_config,omitempty"`
	HooksEnabled    bool            `json:"hooks_enabled"`
	HooksConfig     *HooksConfig    `json:"hooks_config,omitempty"`
	BackupFiles     []*BackupInfo   `json:"backup_files,omitempty"`
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

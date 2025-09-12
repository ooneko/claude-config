package claude

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupInfo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		backup   *BackupInfo
		expected *BackupInfo
	}{
		{
			name: "complete backup info",
			backup: &BackupInfo{
				Filename:    "claude-config-backup-20250110_143052.tar.gz",
				FilePath:    "/Users/test/claude-config-backup-20250110_143052.tar.gz",
				Timestamp:   now,
				Size:        1024,
				ContentType: "directory",
			},
			expected: &BackupInfo{
				Filename:    "claude-config-backup-20250110_143052.tar.gz",
				FilePath:    "/Users/test/claude-config-backup-20250110_143052.tar.gz",
				Timestamp:   now,
				Size:        1024,
				ContentType: "directory",
			},
		},
		{
			name: "settings backup info",
			backup: &BackupInfo{
				Filename:    "settings.json.backup.20250110_143052",
				FilePath:    "/Users/test/.claude/settings.json.backup.20250110_143052",
				Timestamp:   now,
				Size:        512,
				ContentType: "settings",
			},
			expected: &BackupInfo{
				Filename:    "settings.json.backup.20250110_143052",
				FilePath:    "/Users/test/.claude/settings.json.backup.20250110_143052",
				Timestamp:   now,
				Size:        512,
				ContentType: "settings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.Filename, tt.backup.Filename)
			assert.Equal(t, tt.expected.FilePath, tt.backup.FilePath)
			assert.Equal(t, tt.expected.Timestamp, tt.backup.Timestamp)
			assert.Equal(t, tt.expected.Size, tt.backup.Size)
			assert.Equal(t, tt.expected.ContentType, tt.backup.ContentType)
		})
	}
}

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

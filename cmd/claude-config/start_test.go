package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStartCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		wantErr  bool
		errMsg   string
		setup    func() func()
		validate func(t *testing.T, output string)
	}{
		{
			name:    "valid DeepSeek provider",
			args:    []string{"start", "deepseek"},
			wantErr: false,
			setup: func() func() {
				// Mock API key file
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.deepseek_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-test123"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:    "valid Kimi provider with custom model",
			args:    []string{"start", "kimi", "--model", "kimi-plus"},
			wantErr: false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.kimi_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-kimi456"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:    "valid GLM provider with API key",
			args:    []string{"start", "GLM", "--api-key", "sk-glm789"},
			wantErr: false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:    "start without provider argument (native Claude Code)",
			args:    []string{"start"},
			wantErr: false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				// 创建带有现有配置的 settings.json
				settingsContent := `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "old-token",
    "ANTHROPIC_BASE_URL": "old-url",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "old-model"
  }
}`
				err = os.WriteFile(claudeDir+"/settings.json", []byte(settingsContent), 0644)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:    "unsupported provider",
			args:    []string{"start", "unknown"},
			wantErr: true,
			errMsg:  "unsupported provider: unknown",
		},
		{
			name:    "missing API key",
			args:    []string{"start", "deepseek"},
			wantErr: true,
			errMsg:  "API key not found",
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup()
				defer cleanup()
			}

			// 设置环境变量
			for key, value := range tt.envVars {
				originalValue := os.Getenv(key)
				os.Setenv(key, value)
				defer os.Setenv(key, originalValue)
			}

			// 设置 mock 命令用于测试
			os.Setenv("CLAUDE_MOCK", "echo")
			defer os.Unsetenv("CLAUDE_MOCK")

			cmd := createStartCmd()

			// 捕获输出
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// 设置参数，排除 "start" 命令本身
			args := tt.args
			if len(args) > 0 && args[0] == "start" {
				args = args[1:]
			}
			cmd.SetArgs(args)

			// 执行命令（使用 context 支持超时）
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd.SetContext(ctx)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, buf.String())
				}
			}
		})
	}
}

func TestStartNativeClaude(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig string
		wantErr        bool
		validate       func(t *testing.T, tempDir string)
	}{
		{
			name: "cleans existing ANTHROPIC env config",
			existingConfig: `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "old-token",
    "ANTHROPIC_BASE_URL": "old-url",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "old-haiku",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "old-sonnet",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "old-opus"
  }
}`,
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				// 验证 settings.json 中的 env 配置被清理
				settingsPath := tempDir + "/.claude/settings.json"
				content, err := os.ReadFile(settingsPath)
				require.NoError(t, err)

				// 验证不再包含 ANTHROPIC_ 配置
				assert.NotContains(t, string(content), "ANTHROPIC_AUTH_TOKEN")
				assert.NotContains(t, string(content), "ANTHROPIC_BASE_URL")
				assert.NotContains(t, string(content), "ANTHROPIC_DEFAULT_HAIKU_MODEL")
				assert.NotContains(t, string(content), "ANTHROPIC_DEFAULT_SONNET_MODEL")
				assert.NotContains(t, string(content), "ANTHROPIC_DEFAULT_OPUS_MODEL")
			},
		},
		{
			name: "preserves non-ANTHROPIC config",
			existingConfig: `{
  "includeCoAuthoredBy": true,
  "hooks": {
    "pre-commit": "echo test"
  },
  "env": {
    "OTHER_VAR": "other-value",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "old-model"
  }
}`,
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				settingsPath := tempDir + "/.claude/settings.json"
				content, err := os.ReadFile(settingsPath)
				require.NoError(t, err)

				// 验证非 ANTHROPIC 配置保留
				assert.Contains(t, string(content), "includeCoAuthoredBy")
				assert.Contains(t, string(content), "OTHER_VAR")
				assert.Contains(t, string(content), "hooks")

				// 验证 ANTHROPIC 配置被清理
				assert.NotContains(t, string(content), "ANTHROPIC_DEFAULT_HAIKU_MODEL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer func() {
				os.Setenv("HOME", originalHome)
			}()

			claudeDir := tempDir + "/.claude"
			err := os.MkdirAll(claudeDir, 0755)
			require.NoError(t, err)

			// 创建现有的配置文件
			if tt.existingConfig != "" {
				err = os.WriteFile(claudeDir+"/settings.json", []byte(tt.existingConfig), 0644)
				require.NoError(t, err)
			}

			// 设置 mock 命令
			os.Setenv("CLAUDE_MOCK", "echo")
			defer os.Unsetenv("CLAUDE_MOCK")

			cmd := createStartCmd()
			cmd.SetArgs([]string{}) // 无参数调用

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd.SetContext(ctx)

			err = cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tempDir)
				}
			}
		})
	}
}

func TestParseStartArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantProvider string
		wantAPIKey   string
		wantModel    string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid args with provider only",
			args:         []string{"deepseek"},
			wantProvider: "deepseek",
			wantErr:      false,
		},
		{
			name:         "valid args with API key",
			args:         []string{"kimi", "--api-key", "sk-test123"},
			wantProvider: "kimi",
			wantAPIKey:   "sk-test123",
			wantErr:      false,
		},
		{
			name:         "valid args with model",
			args:         []string{"GLM", "--model", "glm-4-plus"},
			wantProvider: "GLM",
			wantModel:    "glm-4-plus",
			wantErr:      false,
		},
		{
			name:         "valid args with all options",
			args:         []string{"doubao", "--api-key", "sk-doubao", "--model", "doubao-v2"},
			wantProvider: "doubao",
			wantAPIKey:   "sk-doubao",
			wantModel:    "doubao-v2",
			wantErr:      false,
		},
		{
			name:         "empty args (native Claude Code)",
			args:         []string{},
			wantProvider: "",
			wantAPIKey:   "",
			wantModel:    "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个临时的 cobra 命令来测试参数解析
			cmd := createStartCmd()
			cmd.SetArgs(tt.args)

			// 执行 flag 解析
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("ParseFlags() unexpected error = %v", err)
				}
				return
			}

			// 获取解析后的参数
			provider, apiKey, model, err := parseStartArgs(cmd)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProvider, provider)
				assert.Equal(t, tt.wantAPIKey, apiKey)
				assert.Equal(t, tt.wantModel, model)
			}
		})
	}
}

// TestPassthroughArgs 测试透传参数功能
func TestPassthroughArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantArgs []string
		wantErr  bool
		errMsg   string
		setup    func() func()
	}{
		{
			name:     "pass through single argument",
			args:     []string{"deepseek", "--", "--dangerously-skip-permissions"},
			wantArgs: []string{"--dangerously-skip-permissions"},
			wantErr:  false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.deepseek_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-test123"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:     "pass through multiple arguments",
			args:     []string{"kimi", "--", "--verbose", "--debug"},
			wantArgs: []string{"--verbose", "--debug"},
			wantErr:  false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.kimi_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-kimi456"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:     "pass through with mixed flags and arguments",
			args:     []string{"GLM", "--model", "glm-4-plus", "--", "--config", "/path/to/config"},
			wantArgs: []string{"--config", "/path/to/config"},
			wantErr:  false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.GLM_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-glm789"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:     "native claude with passthrough args",
			args:     []string{"--", "--dangerously-skip-permissions"},
			wantArgs: []string{"--dangerously-skip-permissions"},
			wantErr:  false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
		{
			name:     "no passthrough args",
			args:     []string{"deepseek"},
			wantArgs: []string{},
			wantErr:  false,
			setup: func() func() {
				tempDir := t.TempDir()
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				claudeDir := tempDir + "/.claude"
				err := os.MkdirAll(claudeDir, 0755)
				require.NoError(t, err)

				apiKeyPath := claudeDir + "/.deepseek_api_key"
				err = os.WriteFile(apiKeyPath, []byte("sk-test123"), 0600)
				require.NoError(t, err)

				return func() {
					os.Setenv("HOME", originalHome)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup()
				defer cleanup()
			}

			// 设置 mock 命令来验证透传的参数
			os.Setenv("CLAUDE_MOCK", "echo")
			defer os.Unsetenv("CLAUDE_MOCK")

			cmd := createStartCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd.SetContext(ctx)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// 验证透传的参数被正确处理
				// 这里我们检查环境变量 CLAUDE_PASSTHROUGH_ARGS 是否包含期望的参数
				passthroughArgs := os.Getenv("CLAUDE_PASSTHROUGH_ARGS")
				if len(tt.wantArgs) > 0 {
					for _, arg := range tt.wantArgs {
						assert.Contains(t, passthroughArgs, arg)
					}
				}
			}
		})
	}
}

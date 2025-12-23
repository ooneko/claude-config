package provider

import (
	"testing"

	"github.com/ooneko/claude-config/internal/claude"
)

func TestEnvMapper_MapToEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		provider claude.ProviderType
		config   *claude.ProviderConfig
		apiKey   string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "DeepSeek provider config",
			provider: claude.ProviderDeepSeek,
			config: &claude.ProviderConfig{
				BaseURL:        "https://api.deepseek.com/anthropic",
				Model:          "deepseek-chat",
				SmallFastModel: "deepseek-chat",
			},
			apiKey: "sk-test123",
			want: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":       "sk-test123",
				"ANTHROPIC_BASE_URL":         "https://api.deepseek.com/anthropic",
				"ANTHROPIC_MODEL":            "deepseek-chat",
				"ANTHROPIC_SMALL_FAST_MODEL": "deepseek-chat",
			},
			wantErr: false,
		},
		{
			name:     "Kimi provider config",
			provider: claude.ProviderKimi,
			config: &claude.ProviderConfig{
				BaseURL:        "https://api.kimi.com/coding/",
				Model:          "kimi-for-coding",
				SmallFastModel: "kimi-for-coding",
			},
			apiKey: "sk-kimi456",
			want: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":       "sk-kimi456",
				"ANTHROPIC_BASE_URL":         "https://api.kimi.com/coding/",
				"ANTHROPIC_MODEL":            "kimi-for-coding",
				"ANTHROPIC_SMALL_FAST_MODEL": "kimi-for-coding",
			},
			wantErr: false,
		},
		{
			name:     "GLM provider config",
			provider: claude.ProviderGLM,
			config: &claude.ProviderConfig{
				BaseURL:        "https://open.bigmodel.cn/api/anthropic",
				Model:          "glm-4.7",
				SmallFastModel: "glm-4.7",
			},
			apiKey: "sk-glm789",
			want: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":       "sk-glm789",
				"ANTHROPIC_BASE_URL":         "https://open.bigmodel.cn/api/anthropic",
				"ANTHROPIC_MODEL":            "glm-4.7",
				"ANTHROPIC_SMALL_FAST_MODEL": "glm-4.7",
			},
			wantErr: false,
		},
		{
			name:     "Doubao provider config",
			provider: claude.ProviderDoubao,
			config: &claude.ProviderConfig{
				BaseURL:        "https://ark.cn-beijing.volces.com/api/coding",
				Model:          "doubao-seed-code-preview-latest",
				SmallFastModel: "doubao-seed-code-preview-latest",
			},
			apiKey: "sk-doubao012",
			want: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":       "sk-doubao012",
				"ANTHROPIC_BASE_URL":         "https://ark.cn-beijing.volces.com/api/coding",
				"ANTHROPIC_MODEL":            "doubao-seed-code-preview-latest",
				"ANTHROPIC_SMALL_FAST_MODEL": "doubao-seed-code-preview-latest",
			},
			wantErr: false,
		},
		{
			name:     "unknown provider",
			provider: claude.ProviderType("unknown"),
			config:   &claude.ProviderConfig{},
			apiKey:   "sk-test",
			wantErr:  true,
		},
		{
			name:     "nil API key",
			provider: claude.ProviderDeepSeek,
			config:   &claude.ProviderConfig{},
			apiKey:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewEnvMapper()
			got, err := mapper.MapToEnvironment(tt.provider, tt.config, tt.apiKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("MapToEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("MapToEnvironment() got %d env vars, want %d", len(got), len(tt.want))
				}

				for key, wantVal := range tt.want {
					if gotVal, exists := got[key]; !exists {
						t.Errorf("MapToEnvironment() missing environment variable %s", key)
					} else if gotVal != wantVal {
						t.Errorf("MapToEnvironment() %s = %v, want %v", key, gotVal, wantVal)
					}
				}
			}
		})
	}
}

func TestEnvMapper_ValidateProviderConfig(t *testing.T) {
	tests := []struct {
		name     string
		provider claude.ProviderType
		config   *claude.ProviderConfig
		apiKey   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid DeepSeek config",
			provider: claude.ProviderDeepSeek,
			config: &claude.ProviderConfig{
				BaseURL: "https://api.deepseek.com/anthropic",
				Model:   "deepseek-chat",
			},
			apiKey:  "sk-valid123",
			wantErr: false,
		},
		{
			name:     "empty API key",
			provider: claude.ProviderDeepSeek,
			config:   &claude.ProviderConfig{},
			apiKey:   "",
			wantErr:  true,
			errMsg:   "API key is required",
		},
		{
			name:     "empty base URL",
			provider: claude.ProviderDeepSeek,
			config: &claude.ProviderConfig{
				Model: "deepseek-chat",
			},
			apiKey:  "sk-valid123",
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name:     "empty model",
			provider: claude.ProviderDeepSeek,
			config: &claude.ProviderConfig{
				BaseURL: "https://api.deepseek.com/anthropic",
			},
			apiKey:  "sk-valid123",
			wantErr: true,
			errMsg:  "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewEnvMapper()
			err := mapper.ValidateProviderConfig(tt.provider, tt.config, tt.apiKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviderConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ValidateProviderConfig() error = %v, wantMsg %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

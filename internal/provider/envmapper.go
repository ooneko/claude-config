package provider

import (
	"fmt"

	"github.com/ooneko/claude-config/internal/claude"
)

// EnvMapper 将 provider 配置映射到环境变量
type EnvMapper struct{}

// NewEnvMapper 创建新的环境变量映射器
func NewEnvMapper() *EnvMapper {
	return &EnvMapper{}
}

// MapToEnvironment 将 provider 配置映射为 ANTHROPIC_* 环境变量
func (m *EnvMapper) MapToEnvironment(provider claude.ProviderType, config *claude.ProviderConfig, apiKey string) (map[string]string, error) {
	if err := m.ValidateProviderConfig(provider, config, apiKey); err != nil {
		return nil, err
	}

	envVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": apiKey,
		"ANTHROPIC_BASE_URL":   config.BaseURL,
	}

	// 添加默认模型环境变量
	m.addDefaultModelEnvVars(envVars, provider, config)

	return envVars, nil
}

// addDefaultModelEnvVars 添加默认模型环境变量
func (m *EnvMapper) addDefaultModelEnvVars(envVars map[string]string, provider claude.ProviderType, config *claude.ProviderConfig) {
	var haikuModel, sonnetModel, opusModel string

	switch provider {
	case claude.ProviderDeepSeek:
		haikuModel = config.Model
		sonnetModel = config.Model
		opusModel = config.Model
	case claude.ProviderKimi:
		haikuModel = config.Model
		sonnetModel = config.Model
		opusModel = config.Model
	case claude.ProviderGLM:
		haikuModel = "glm-4.7"
		sonnetModel = "glm-4.7"
		opusModel = "glm-4.7"
	case claude.ProviderDoubao:
		haikuModel = config.Model
		sonnetModel = config.Model
		opusModel = config.Model
	}

	envVars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = haikuModel
	envVars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = sonnetModel
	envVars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = opusModel
}

// ValidateProviderConfig 验证 provider 配置是否完整
func (m *EnvMapper) ValidateProviderConfig(provider claude.ProviderType, config *claude.ProviderConfig, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	if config == nil {
		return fmt.Errorf("provider config is required")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// 验证 provider 是否支持
	switch provider {
	case claude.ProviderDeepSeek, claude.ProviderKimi, claude.ProviderGLM, claude.ProviderDoubao:
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

package aiprovider

import "fmt"

// DeepSeekProvider implements the Provider interface for DeepSeek
type DeepSeekProvider struct{}

// GetType returns the provider type
func (p *DeepSeekProvider) GetType() ProviderType {
	return ProviderDeepSeek
}

// GetDefaultConfig returns the default configuration for DeepSeek
func (p *DeepSeekProvider) GetDefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		Type:           ProviderDeepSeek,
		AuthToken:      apiKey,
		BaseURL:        "https://api.deepseek.com/anthropic",
		Model:          "deepseek-chat",
		SmallFastModel: "deepseek-chat",
	}
}

// ValidateConfig validates the DeepSeek configuration
func (p *DeepSeekProvider) ValidateConfig(config *ProviderConfig) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required for DeepSeek")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required for DeepSeek")
	}
	return nil
}

// KimiProvider implements the Provider interface for Kimi
type KimiProvider struct{}

// GetType returns the provider type
func (p *KimiProvider) GetType() ProviderType {
	return ProviderKimi
}

// GetDefaultConfig returns the default configuration for Kimi
func (p *KimiProvider) GetDefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		Type:           ProviderKimi,
		AuthToken:      apiKey,
		BaseURL:        "https://api.kimi.com/coding/",
		Model:          "kimi-for-coding",
		SmallFastModel: "kimi-for-coding",
	}
}

// ValidateConfig validates the Kimi configuration
func (p *KimiProvider) ValidateConfig(config *ProviderConfig) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required for Kimi")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required for Kimi")
	}
	return nil
}

// GLMProvider implements the Provider interface for GLM
type GLMProvider struct{}

// GetType returns the provider type
func (p *GLMProvider) GetType() ProviderType {
	return ProviderGLM
}

// GetDefaultConfig returns the default configuration for GLM
func (p *GLMProvider) GetDefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		Type:           ProviderGLM,
		AuthToken:      apiKey,
		BaseURL:        "https://open.bigmodel.cn/api/anthropic",
		Model:          "glm-4.6",
		SmallFastModel: "glm-4.6",
	}
}

// ValidateConfig validates the GLM configuration
func (p *GLMProvider) ValidateConfig(config *ProviderConfig) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required for GLM")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required for GLM")
	}
	return nil
}

// DoubaoProvider implements the Provider interface for Doubao
type DoubaoProvider struct{}

// GetType returns the provider type
func (p *DoubaoProvider) GetType() ProviderType {
	return ProviderDoubao
}

// GetDefaultConfig returns the default configuration for Doubao
func (p *DoubaoProvider) GetDefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		Type:           ProviderDoubao,
		AuthToken:      apiKey,
		BaseURL:        "https://ark.cn-beijing.volces.com/api/coding",
		Model:          "doubao-seed-code-preview-latest",
		SmallFastModel: "doubao-seed-code-preview-latest",
	}
}

// ValidateConfig validates the Doubao configuration
func (p *DoubaoProvider) ValidateConfig(config *ProviderConfig) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required for Doubao")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required for Doubao")
	}
	return nil
}

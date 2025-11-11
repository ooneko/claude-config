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

// ZhipuProvider implements the Provider interface for Zhipu
type ZhipuProvider struct{}

// GetType returns the provider type
func (p *ZhipuProvider) GetType() ProviderType {
	return ProviderZhiPu
}

// GetDefaultConfig returns the default configuration for Zhipu
func (p *ZhipuProvider) GetDefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		Type:           ProviderZhiPu,
		AuthToken:      apiKey,
		BaseURL:        "https://open.bigmodel.cn/api/anthropic",
		Model:          "glm-4.6",
		SmallFastModel: "glm-4.6",
	}
}

// ValidateConfig validates the Zhipu configuration
func (p *ZhipuProvider) ValidateConfig(config *ProviderConfig) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required for Zhipu")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required for Zhipu")
	}
	return nil
}

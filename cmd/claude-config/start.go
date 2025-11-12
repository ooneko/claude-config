package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ooneko/claude-config/internal/aiprovider"
	"github.com/ooneko/claude-config/internal/claude"
	"github.com/ooneko/claude-config/internal/provider"
	"github.com/spf13/cobra"
)

// anthropicEnvVars 需要清理的 ANTHROPIC 相关环境变量
var anthropicEnvVars = []string{
	"ANTHROPIC_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_MODEL",
	"ANTHROPIC_SMALL_FAST_MODEL",
}

type startOptions struct {
	apiKey string
	model  string
}

func createStartCmd() *cobra.Command {
	opts := &startOptions{}

	cmd := &cobra.Command{
		Use:   "start [provider]",
		Short: "启动 Claude Code，可指定 AI provider",
		Long: `启动 Claude Code，可选择指定 AI provider 通过环境变量设置配置。

无参数时启动原生 Claude Code（清理现有配置）。
支持以下 provider:
- deepseek: DeepSeek API
- kimi: Kimi API
- GLM: 智谱 GLM API
- doubao: 豆包 API

示例:
  claude-config start              # 启动原生 Claude Code
  claude-config start deepseek
  claude-config start kimi --model kimi-plus
  claude-config start GLM --api-key sk-xxxxxxxx`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd, args, opts)
		},
	}

	cmd.Flags().StringVar(&opts.apiKey, "api-key", "", "API 密钥 (可选，优先使用存储的密钥)")
	cmd.Flags().StringVar(&opts.model, "model", "", "指定模型 (可选，使用 provider 默认模型)")

	return cmd
}

func runStart(_ *cobra.Command, args []string, opts *startOptions) error {
	// 获取 home 目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")

	// 无参数：启动原生 Claude Code
	if len(args) == 0 {
		return startNativeClaude(claudeDir)
	}

	// 有参数：启动指定 provider
	return startWithProvider(claudeDir, args[0], opts)
}

func parseProviderFromArg(arg string) (claude.ProviderType, error) {
	providerType := claude.NormalizeProviderName(arg)

	if providerType == claude.ProviderNone {
		return "", fmt.Errorf("unsupported provider: %s", arg)
	}

	return providerType, nil
}

func loadStoredAPIKey(claudeDir string, providerType claude.ProviderType) (string, error) {
	apiKeyPath := filepath.Join(claudeDir, "."+string(providerType)+"_api_key")

	data, err := os.ReadFile(apiKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("API key not found for provider %s, please provide --api-key or configure first", providerType)
		}
		return "", fmt.Errorf("failed to read API key file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func getProvider(providerType claude.ProviderType) aiprovider.Provider {
	switch providerType {
	case claude.ProviderDeepSeek:
		return &aiprovider.DeepSeekProvider{}
	case claude.ProviderKimi:
		return &aiprovider.KimiProvider{}
	case claude.ProviderGLM:
		return &aiprovider.GLMProvider{}
	case claude.ProviderDoubao:
		return &aiprovider.DoubaoProvider{}
	default:
		return nil
	}
}

func startClaudeCode(envVars map[string]string) error {
	// 设置环境变量
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// 检查是否存在 CLAUDE_MOCK 环境变量（用于测试）
	if mockCmd := os.Getenv("CLAUDE_MOCK"); mockCmd != "" {
		cmd := exec.Command(mockCmd)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// 启动 Claude Code (假设 claude 命令在 PATH 中)
	cmd := exec.Command("claude")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// startNativeClaude 启动原生 Claude Code（清理配置）
func startNativeClaude(claudeDir string) error {
	if err := cleanAnthropicConfig(claudeDir); err != nil {
		fmt.Printf("Warning: failed to clean existing config: %v\n", err)
	}

	// 启动原生 Claude Code（无环境变量）
	return startClaudeCode(map[string]string{})
}

// cleanAnthropicConfig 清理 settings.json 和环境变量中的 ANTHROPIC 配置
func cleanAnthropicConfig(claudeDir string) error {
	// 清理 settings.json 中的配置
	manager := aiprovider.NewManager(claudeDir)
	if err := manager.Off(context.Background()); err != nil {
		return fmt.Errorf("failed to clean settings.json: %w", err)
	}

	// 清理环境变量
	for _, envVar := range anthropicEnvVars {
		os.Unsetenv(envVar)
	}

	return nil
}

// startWithProvider 启动指定 provider 的 Claude Code
func startWithProvider(claudeDir string, providerArg string, opts *startOptions) error {
	providerType, err := parseProviderFromArg(providerArg)
	if err != nil {
		return err
	}

	// 获取 API 密钥
	apiKey, err := getAPIKey(claudeDir, providerType, opts.apiKey)
	if err != nil {
		return err
	}

	// 获取环境变量配置
	envVars, err := buildProviderEnvVars(providerType, apiKey, opts.model)
	if err != nil {
		return err
	}

	// 启动 Claude Code
	return startClaudeCode(envVars)
}

// getAPIKey 获取 API 密钥，优先使用命令行参数，其次使用存储的密钥
func getAPIKey(claudeDir string, providerType claude.ProviderType, cmdAPIKey string) (string, error) {
	if cmdAPIKey != "" {
		return cmdAPIKey, nil
	}

	return loadStoredAPIKey(claudeDir, providerType)
}

// buildProviderEnvVars 构建 provider 的环境变量配置
func buildProviderEnvVars(providerType claude.ProviderType, apiKey, model string) (map[string]string, error) {
	// 获取 provider 配置
	prov := getProvider(providerType)
	providerConfig := prov.GetDefaultConfig(apiKey)

	// 应用命令行参数覆盖
	if model != "" {
		providerConfig.Model = model
		providerConfig.SmallFastModel = model
	}

	// 映射到环境变量
	mapper := provider.NewEnvMapper()
	return mapper.MapToEnvironment(providerType, providerConfig, apiKey)
}

// parseStartArgs 解析启动命令参数
func parseStartArgs(cmd *cobra.Command) (string, string, string, error) {
	// 获取位置参数（非 flag 参数）
	args := cmd.Flags().Args()
	if len(args) > 1 {
		return "", "", "", fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	// 支持无参数调用
	if len(args) == 0 {
		return "", "", "", nil
	}

	provider := args[0]

	apiKey, _ := cmd.Flags().GetString("api-key")
	model, _ := cmd.Flags().GetString("model")

	return provider, apiKey, model, nil
}

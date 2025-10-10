package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSettingsJSONMerger(t *testing.T) {
	merger := NewSettingsJSONMerger()
	assert.NotNil(t, merger)
}

func TestSettingsJsonMerger_ShouldPreserveProxyConfig(t *testing.T) {
	merger := NewSettingsJSONMerger()

	tests := []struct {
		name       string
		targetData map[string]interface{}
		expected   bool
	}{
		{
			name: "有HTTP代理",
			targetData: map[string]interface{}{
				"env": map[string]interface{}{
					"http_proxy": "http://127.0.0.1:7890",
				},
			},
			expected: true,
		},
		{
			name: "有HTTPS代理",
			targetData: map[string]interface{}{
				"env": map[string]interface{}{
					"https_proxy": "http://127.0.0.1:7890",
				},
			},
			expected: true,
		},
		{
			name: "有HTTP和HTTPS代理",
			targetData: map[string]interface{}{
				"env": map[string]interface{}{
					"http_proxy":  "http://127.0.0.1:7890",
					"https_proxy": "http://127.0.0.1:7890",
				},
			},
			expected: true,
		},
		{
			name: "没有代理配置",
			targetData: map[string]interface{}{
				"env": map[string]interface{}{
					"OTHER_VAR": "value",
				},
			},
			expected: false,
		},
		{
			name:       "没有env字段",
			targetData: map[string]interface{}{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merger.ShouldPreserveProxyConfig(tt.targetData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSettingsJsonMerger_FilterProxyFromSource(t *testing.T) {
	merger := NewSettingsJSONMerger()

	sourceData := map[string]interface{}{
		"includeCoAuthoredBy": true,
		"env": map[string]interface{}{
			"http_proxy":  "http://proxy.example.com:8080",
			"https_proxy": "http://proxy.example.com:8080",
			"OTHER_VAR":   "keep_this",
		},
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Write|Edit",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "lint.sh",
						},
					},
				},
			},
		},
	}

	result := merger.FilterProxyFromSource(sourceData)

	// 检查代理配置被移除
	env, ok := result["env"].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, env, "http_proxy")
	assert.NotContains(t, env, "https_proxy")
	assert.Contains(t, env, "OTHER_VAR")

	// 检查其他配置保留
	assert.Equal(t, true, result["includeCoAuthoredBy"])
	assert.Contains(t, result, "hooks")
}

func TestSettingsJsonMerger_DeepMergeDict(t *testing.T) {
	merger := NewSettingsJSONMerger()

	target := map[string]interface{}{
		"includeCoAuthoredBy": false,
		"env": map[string]interface{}{
			"http_proxy":              "http://127.0.0.1:7890",
			"CLAUDE_HOOKS_GO_ENABLED": "false",
		},
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Write|Edit",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "custom-lint.sh",
						},
					},
				},
			},
		},
	}

	source := map[string]interface{}{
		"includeCoAuthoredBy": true,
		"env": map[string]interface{}{
			"NEW_VAR": "new_value",
		},
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Write|Edit|MultiEdit",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "smart-lint.sh",
						},
					},
				},
			},
			"Stop": []interface{}{
				map[string]interface{}{
					"matcher": "",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "notifier.sh",
						},
					},
				},
			},
		},
	}

	result := merger.DeepMergeDict(target, source)

	// 检查基本字段被覆盖
	assert.Equal(t, true, result["includeCoAuthoredBy"])

	// 检查env被合并
	env, ok := result["env"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:7890", env["http_proxy"])
	assert.Equal(t, "false", env["CLAUDE_HOOKS_GO_ENABLED"])
	assert.Equal(t, "new_value", env["NEW_VAR"])

	// 检查hooks被智能合并
	hooks, ok := result["hooks"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, hooks, "PostToolUse")
	assert.Contains(t, hooks, "Stop")
}

func TestSettingsJsonMerger_MergeSettings(t *testing.T) {
	merger := NewSettingsJSONMerger()
	tempDir := t.TempDir()

	// 创建源文件
	sourceFile := filepath.Join(tempDir, "source.json")
	sourceData := map[string]interface{}{
		"includeCoAuthoredBy": true,
		"env": map[string]interface{}{
			"http_proxy": "http://proxy.example.com:8080",
			"NEW_VAR":    "new_value",
		},
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Write|Edit",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "smart-lint.sh",
						},
					},
				},
			},
		},
	}

	sourceJSON, _ := json.MarshalIndent(sourceData, "", "  ")
	err := os.WriteFile(sourceFile, sourceJSON, 0644)
	require.NoError(t, err)

	// 创建目标文件
	targetFile := filepath.Join(tempDir, "target.json")
	targetData := map[string]interface{}{
		"includeCoAuthoredBy": false,
		"env": map[string]interface{}{
			"http_proxy":              "http://127.0.0.1:7890",
			"CLAUDE_HOOKS_GO_ENABLED": "false",
		},
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Write|Edit",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "custom-lint.sh",
						},
					},
				},
			},
		},
	}

	targetJSON, _ := json.MarshalIndent(targetData, "", "  ")
	err = os.WriteFile(targetFile, targetJSON, 0644)
	require.NoError(t, err)

	// 执行合并
	err = merger.MergeSettings(targetFile, sourceFile)
	require.NoError(t, err)

	// 验证合并结果
	mergedData, err := merger.readJSONFile(targetFile)
	require.NoError(t, err)

	// 检查代理配置被保留
	env, ok := mergedData["env"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:7890", env["http_proxy"])

	// 检查新变量被添加
	assert.Equal(t, "new_value", env["NEW_VAR"])
	assert.Equal(t, "false", env["CLAUDE_HOOKS_GO_ENABLED"])

	// 检查includeCoAuthoredBy被更新
	assert.Equal(t, true, mergedData["includeCoAuthoredBy"])
}

func TestSettingsJsonMerger_MergeSettings_NoTargetFile(t *testing.T) {
	merger := NewSettingsJSONMerger()
	tempDir := t.TempDir()

	// 创建源文件（包含代理配置）
	sourceFile := filepath.Join(tempDir, "source.json")
	sourceData := map[string]interface{}{
		"includeCoAuthoredBy": true,
		"env": map[string]interface{}{
			"http_proxy": "http://proxy.example.com:8080",
			"NEW_VAR":    "new_value",
		},
	}

	sourceJSON, _ := json.MarshalIndent(sourceData, "", "  ")
	err := os.WriteFile(sourceFile, sourceJSON, 0644)
	require.NoError(t, err)

	// 目标文件不存在
	targetFile := filepath.Join(tempDir, "target.json")

	// 执行合并
	err = merger.MergeSettings(targetFile, sourceFile)
	require.NoError(t, err)

	// 验证结果
	resultData, err := merger.readJSONFile(targetFile)
	require.NoError(t, err)

	// 检查代理配置被移除
	env, ok := resultData["env"].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, env, "http_proxy")
	assert.Equal(t, "new_value", env["NEW_VAR"])

	// 检查其他配置保留
	assert.Equal(t, true, resultData["includeCoAuthoredBy"])
}

func TestSettingsJsonMerger_MergeHooks(t *testing.T) {
	merger := NewSettingsJSONMerger()

	targetHooks := map[string]interface{}{
		"PostToolUse": []interface{}{
			map[string]interface{}{
				"matcher": "Write|Edit",
				"hooks": []interface{}{
					map[string]interface{}{
						"type":    "command",
						"command": "custom-lint.sh",
					},
				},
			},
		},
	}

	sourceHooks := map[string]interface{}{
		"PostToolUse": []interface{}{
			map[string]interface{}{
				"matcher": "Write|Edit|MultiEdit",
				"hooks": []interface{}{
					map[string]interface{}{
						"type":    "command",
						"command": "smart-lint.sh",
					},
				},
			},
		},
		"Stop": []interface{}{
			map[string]interface{}{
				"matcher": "",
				"hooks": []interface{}{
					map[string]interface{}{
						"type":    "command",
						"command": "notifier.sh",
					},
				},
			},
		},
	}

	result := merger.MergeHooks(targetHooks, sourceHooks)

	// 检查PostToolUse被合并
	assert.Contains(t, result, "PostToolUse")
	assert.Contains(t, result, "Stop")

	postToolUse, ok := result["PostToolUse"].([]interface{})
	require.True(t, ok)
	assert.Len(t, postToolUse, 2) // 两个不同的matcher
}

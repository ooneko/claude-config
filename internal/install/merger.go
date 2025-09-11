package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SettingsJsonMerger settings.json智能合并器
type SettingsJsonMerger struct{}

// NewSettingsJsonMerger 创建新的settings.json合并器
func NewSettingsJsonMerger() *SettingsJsonMerger {
	return &SettingsJsonMerger{}
}

// ShouldPreserveProxyConfig 检查是否应该保留目标文件中的代理配置
func (m *SettingsJsonMerger) ShouldPreserveProxyConfig(targetData map[string]interface{}) bool {
	env, ok := targetData["env"].(map[string]interface{})
	if !ok {
		return false
	}

	_, hasHTTP := env["http_proxy"]
	_, hasHTTPS := env["https_proxy"]
	return hasHTTP || hasHTTPS
}

// FilterProxyFromSource 从源数据中移除代理配置
func (m *SettingsJsonMerger) FilterProxyFromSource(sourceData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 深度复制
	for k, v := range sourceData {
		result[k] = m.deepCopyValue(v)
	}

	if env, ok := result["env"].(map[string]interface{}); ok {
		delete(env, "http_proxy")
		delete(env, "https_proxy")

		// 如果env为空，删除env字段
		if len(env) == 0 {
			delete(result, "env")
		}
	}

	return result
}

// deepCopyValue 深度复制值
func (m *SettingsJsonMerger) deepCopyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = m.deepCopyValue(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = m.deepCopyValue(val)
		}
		return result
	default:
		return v
	}
}

// DeepMergeDict 深度合并字典，source覆盖target
func (m *SettingsJsonMerger) DeepMergeDict(target, source map[string]interface{}) map[string]interface{} {
	result := m.deepCopyValue(target).(map[string]interface{})

	for key, value := range source {
		if existing, exists := result[key]; exists {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					// 特殊处理hooks字典
					if key == "hooks" {
						result[key] = m.MergeHooks(existingMap, sourceMap)
					} else {
						result[key] = m.DeepMergeDict(existingMap, sourceMap)
					}
					continue
				}
			}
			if existingSlice, ok := existing.([]interface{}); ok {
				if sourceSlice, ok := value.([]interface{}); ok {
					// 数组合并去重（只处理基本类型）
					combined := append(existingSlice, sourceSlice...)
					result[key] = m.uniqueSlice(combined)
					continue
				}
			}
		}
		result[key] = m.deepCopyValue(value)
	}

	return result
}

// uniqueSlice 数组去重
func (m *SettingsJsonMerger) uniqueSlice(slice []interface{}) []interface{} {
	seen := make(map[string]bool)
	var result []interface{}

	for _, item := range slice {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// 字典类型不能直接hash，直接添加
			result = append(result, itemMap)
		} else {
			// 基本类型可以去重
			key := fmt.Sprintf("%v", item)
			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
	}

	return result
}

// MergeHooks 智能合并hooks配置
func (m *SettingsJsonMerger) MergeHooks(targetHooks, sourceHooks map[string]interface{}) map[string]interface{} {
	result := m.deepCopyValue(targetHooks).(map[string]interface{})

	for eventType, sourceConfigs := range sourceHooks {
		if existing, exists := result[eventType]; exists {
			// 合并同一事件类型的hooks
			if existingSlice, ok := existing.([]interface{}); ok {
				if sourceSlice, ok := sourceConfigs.([]interface{}); ok {
					result[eventType] = m.mergeHookConfigs(existingSlice, sourceSlice)
				}
			}
		} else {
			result[eventType] = m.deepCopyValue(sourceConfigs)
		}
	}

	return result
}

// mergeHookConfigs 合并同一事件类型的hook配置
func (m *SettingsJsonMerger) mergeHookConfigs(existing, source []interface{}) []interface{} {
	// 按matcher建立映射
	existingMatchers := make(map[string]int)
	for i, config := range existing {
		if configMap, ok := config.(map[string]interface{}); ok {
			if matcher, ok := configMap["matcher"].(string); ok {
				existingMatchers[matcher] = i
			}
		}
	}

	result := m.deepCopyValue(existing).([]interface{})

	for _, config := range source {
		configMap, ok := config.(map[string]interface{})
		if !ok {
			continue
		}

		matcher, ok := configMap["matcher"].(string)
		if !ok {
			matcher = ""
		}

		if existingIndex, exists := existingMatchers[matcher]; exists {
			// 相同matcher，合并hooks命令
			if resultMap, ok := result[existingIndex].(map[string]interface{}); ok {
				existingHooks, _ := resultMap["hooks"].([]interface{})
				newHooks, _ := configMap["hooks"].([]interface{})

				mergedHooks := m.mergeHookCommands(existingHooks, newHooks)
				resultMap["hooks"] = mergedHooks
			}
		} else {
			// 新的matcher，直接添加
			result = append(result, m.deepCopyValue(config))
		}
	}

	return result
}

// mergeHookCommands 按command去重合并hook命令
func (m *SettingsJsonMerger) mergeHookCommands(existing, new []interface{}) []interface{} {
	existingCommands := make(map[string]bool)

	// 记录现有命令
	for _, hook := range existing {
		if hookMap, ok := hook.(map[string]interface{}); ok {
			if command, ok := hookMap["command"].(string); ok {
				existingCommands[command] = true
			}
		}
	}

	result := m.deepCopyValue(existing).([]interface{})

	// 添加新命令（去重）
	for _, hook := range new {
		if hookMap, ok := hook.(map[string]interface{}); ok {
			if command, ok := hookMap["command"].(string); ok {
				if !existingCommands[command] {
					result = append(result, m.deepCopyValue(hook))
				}
			}
		}
	}

	return result
}

// MergeSettings 合并settings.json文件
func (m *SettingsJsonMerger) MergeSettings(targetFile, sourceFile string) error {
	// 读取源文件
	sourceData, err := m.readJSONFile(sourceFile)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		// 目标文件不存在，检查源文件是否包含代理配置
		if env, ok := sourceData["env"].(map[string]interface{}); ok {
			if _, hasHTTP := env["http_proxy"]; hasHTTP {
				fmt.Println("⚠️  源文件包含代理配置，但将被跳过")
				fmt.Println("   请使用 claude-config proxy on 来配置代理")
				sourceData = m.FilterProxyFromSource(sourceData)
			}
			if _, hasHTTPS := env["https_proxy"]; hasHTTPS {
				fmt.Println("⚠️  源文件包含代理配置，但将被跳过")
				fmt.Println("   请使用 claude-config proxy on 来配置代理")
				sourceData = m.FilterProxyFromSource(sourceData)
			}
		}

		return m.writeJSONFile(targetFile, sourceData)
	}

	// 读取目标文件
	targetData, err := m.readJSONFile(targetFile)
	if err != nil {
		return fmt.Errorf("读取目标文件失败: %w", err)
	}

	// 检查是否需要保留代理配置
	preserveProxy := m.ShouldPreserveProxyConfig(targetData)

	if preserveProxy {
		fmt.Println("📡 检测到现有代理配置，将保留用户代理设置")
		sourceData = m.FilterProxyFromSource(sourceData)
	}

	// 深度合并
	mergedData := m.DeepMergeDict(targetData, sourceData)

	// 检查是否有变化
	if !m.isEqual(mergedData, targetData) {
		fmt.Println("🔄 检测到settings.json配置变化")
		fmt.Println("将进行智能合并，保留您的个人配置")
		if preserveProxy {
			fmt.Println("   - 保留现有代理配置")
		}

		return m.writeJSONFile(targetFile, mergedData)
	}

	fmt.Println("settings.json配置无变化，跳过")
	return nil
}

// readJSONFile 读取JSON文件
func (m *SettingsJsonMerger) readJSONFile(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// writeJSONFile 写入JSON文件
func (m *SettingsJsonMerger) writeJSONFile(filename string, data map[string]interface{}) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// isEqual 简单比较两个map是否相等
func (m *SettingsJsonMerger) isEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)

	return string(aJSON) == string(bJSON)
}

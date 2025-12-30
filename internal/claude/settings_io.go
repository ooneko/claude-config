package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LoadSettingsRaw 读取 settings.json，保留所有字段（包括未知字段）
// 返回 map[string]interface{} 以确保不会丢失任何字段
func LoadSettingsRaw(settingsPath string) (map[string]interface{}, error) {
	// 如果文件不存在，返回默认设置
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"includeCoAuthoredBy": false,
		}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SaveSettingsRaw 保存 settings.json，保留所有字段
func SaveSettingsRaw(settingsPath string, data map[string]interface{}) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, jsonData, 0644)
}

// RawToSettings 将 map 转换为 *Settings
// 用于将原始数据转换为类型安全的 Settings 结构体
func RawToSettings(raw map[string]interface{}) *Settings {
	data, _ := json.Marshal(raw)
	settings := &Settings{}
	json.Unmarshal(data, settings)
	return settings
}

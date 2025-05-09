package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 命令行工具配置结构
type Config struct {
	// 默认输出格式
	DefaultFormat string `json:"default_format"`
	
	// 默认语言
	DefaultLanguage string `json:"default_language"`
	
	// 是否启用调试模式
	Debug bool `json:"debug"`
	
	// 用户自定义配置
	Custom map[string]interface{} `json:"custom"`
}

// Load 从配置文件加载配置
func Load() (*Config, error) {
	// 设置默认配置
	cfg := &Config{
		DefaultFormat: "text",
		DefaultLanguage: "zh",
		Debug: false,
		Custom: make(map[string]interface{}),
	}

	// 检查环境变量中是否指定了配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 默认查找用户目录下的配置文件
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(homeDir, ".gocmd.json")
		} else {
			// 如果无法获取用户目录，则使用当前目录
			configPath = "config.json"
		}
	}

	// 尝试读取配置文件
	if _, err := os.Stat(configPath); err == nil {
		file, err := os.Open(configPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

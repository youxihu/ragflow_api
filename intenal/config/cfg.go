package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"ragflow_api/intenal/str"
)

func LoadConfig(path string) (str.RagConf, error) {
	var cfg str.RagConf

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return cfg, nil
}

package str

type RagConf struct {
	RagFlow struct {
		DatasetID []string `yaml:"dataset_id"`
		APIKey    string   `yaml:"api_key"`
		Address   string   `yaml:"address"`
		Port      int      `yaml:"port"`
	} `yaml:"ragflow"`

	// 可以在这里添加 DirPath、Timeout 的配置项（可选）
	DirPath string `yaml:"dirPath,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

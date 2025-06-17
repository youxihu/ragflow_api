package main

import (
	"log"
	"ragflow_api/internal/config"
	pd "ragflow_api/internal/fileManagement/parsedocs"
)

// ParseAll
func main() {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/online.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	parser := pd.NewDocParser(cfg)
	if err := parser.ParseZeroChunkDocs(); err != nil {
		log.Fatalf("自动解析失败: %v", err)
	}
}

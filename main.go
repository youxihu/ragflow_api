package main

import (
	"log"
	"ragflow_api/intenal/config"
	pd "ragflow_api/intenal/fileManagement/parsedocs"
)

func main() {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	parser := pd.NewDocParser(cfg)
	if err := parser.ParseZeroChunkDocs(); err != nil {
		log.Fatalf("自动解析失败: %v", err)
	}
}

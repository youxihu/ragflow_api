package parsedocs

import (
	"log"
	"ragflow_api/internal/config"
	"testing"
)

func TestParse(t *testing.T) {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/local.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	parser := NewDocParser(cfg)
	if err := parser.ParseZeroChunkDocs(); err != nil {
		log.Fatalf("自动解析失败: %v", err)
	}
}

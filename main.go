package main

import (
	"log"
	"ragflow_api/internal/config"
	"ragflow_api/internal/fileManagement/listdocs"
)

// ParseAll
//func main() {
//	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/local.auth.yaml")
//	if err != nil {
//		log.Fatalf("加载配置失败: %v", err)
//	}
//
//	parser := pd.NewDocParser(cfg)
//	if err := parser.ParseZeroChunkDocs(); err != nil {
//		log.Fatalf("自动解析失败: %v", err)
//	}
//}

// AreAllDone
func main() {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/local.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	lister := listdocs.NewDocLister(cfg)

	_, err = lister.AreAllDatasetsDone()
	if err != nil {
		log.Fatalf("检查过程中出错: %v", err)
	}
}

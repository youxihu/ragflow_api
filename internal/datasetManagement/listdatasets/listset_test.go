package listdatasets

import (
	"log"
	"ragflow_api/internal/config"
	"testing"
)

func TestListSets(t *testing.T) {
	// 读取配置文件
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	lister := NewDatasetLister(cfg)
	if err := lister.Start(); err != nil {
		log.Fatalf("上传失败: %v", err)
	}
}

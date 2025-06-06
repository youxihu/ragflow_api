package listdocs

import (
	"log"
	"ragflow_api/internal/config"
	"testing"
)

func TestList(t *testing.T) {
	// 读取配置文件
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/online.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	lister := NewDocLister(cfg)
	err = lister.Start(
		1,             // page
		20,            // page_size
		"create_time", // orderby
		true,          // desc
		"",            // keywords
		"",            // document_id
		"",            // document_name
	)
	if err != nil {
		log.Fatalf("获取文档列表失败: %v", err)
	}
}

func TestAreAllDatasetsDone(t *testing.T) {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/local.auth.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	lister := NewDocLister(cfg)

	allDone, err := lister.AreAllDatasetsDone()
	if err != nil {
		t.Errorf("检查过程中出错: %v", err)
	}

	if allDone {
		t.Log("✅ 所有 dataset 的文档都已完成")
	}
}

package listdocs

import (
	"log"
	"ragflow_api/intenal/config"
	"testing"
)

func TestList(t *testing.T) {
	// 读取配置文件
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/auth.yaml")
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

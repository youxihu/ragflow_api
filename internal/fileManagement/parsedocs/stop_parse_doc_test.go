package parsedocs

import (
	"log"
	"ragflow_api/internal/config"
	"testing"
)

func TestStopParsing(t *testing.T) {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/youxihu.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	stopper := NewDocParsStopper(cfg)
	if err := stopper.StopRunningDocuments(); err != nil {
		t.Errorf("🛑 停止解析失败: %v", err)
	} else {
		t.Log("🎉 所有正在解析的文档已成功停止！")
	}
}

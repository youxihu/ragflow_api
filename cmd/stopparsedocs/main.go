package main

import (
	"log"
	"ragflow_api/internal/config"
	spd "ragflow_api/internal/fileManagement/parsedocs"
)

func main() {
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/online.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	stopper := spd.NewDocParsStopper(cfg)
	if err := stopper.StopRunningDocuments(); err != nil {
		log.Fatalf("🛑 停止解析失败: %v", err)
	} else {
		log.Fatalf("🎉 所有正在解析的文档已成功停止！")
	}
}

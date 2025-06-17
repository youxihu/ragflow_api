package main

import (
	"log"
	"ragflow_api/internal/config"
	up "ragflow_api/internal/fileManagement/uploadocs"
)

func main() {
	// 读取配置文件
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/online.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	uploader := up.NewDocUploader(cfg, "files", "60s")
	if err := uploader.Start(); err != nil {
		log.Fatalf("上传失败: %v", err)
	}
}

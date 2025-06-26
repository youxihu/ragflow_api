// main.go
package main

import (
	"log"
	"ragflow_api/internal/config"
	"ragflow_api/pkg/mcpServer"
)

func main() {
	// 从配置文件中读取配置
	cfg, err := config.LoadConfig("/home/youxihu/secret/aiops/rag_api/local.auth.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 启动 MCP 服务
	mcpServer.StartMCPServer(cfg)
}

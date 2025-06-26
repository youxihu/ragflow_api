package mcpServer

import (
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"log"
	"net/http"
	lsdata "ragflow_api/internal/datasetManagement/listdatasets"
	lsdoc "ragflow_api/internal/fileManagement/listdocs"
	"ragflow_api/internal/str"
)

func StartMCPServer(cfg str.RagConf) {
	listDataEr := lsdata.NewDatasetLister(cfg)
	listDocsEr := lsdoc.NewDocLister(cfg)

	sseTransport, handler, err := transport.NewSSEServerTransportAndHandler("/mcp/message")
	if err != nil {
		log.Fatalf("init sse transport fail: %v", err)
	}

	// 创建 MCP Server
	mcpServer, err := server.NewServer(sseTransport)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// 注册工具
	if err := lsdata.RegisterDatasetListTool(mcpServer, listDataEr); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}
	if err := lsdoc.RegisterDocListTool(mcpServer, listDocsEr); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}
	if err := lsdoc.RegisterAreAllDoneTool(mcpServer, listDocsEr); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/mcp/sse", handler.HandleSSE())         // SSE 数据流
	mux.Handle("/mcp/message", handler.HandleMessage()) // 工具消息

	log.Println("✅ MCP 服务已启动，访问地址: http://0.0.0.0:8081/mcp")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

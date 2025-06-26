package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

func main() {
	// 创建 Transport 层
	transportClient, err := transport.NewSSEClientTransport("http://localhost:8081/mcp/sse")

	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	// 创建客户端
	mcpClient, err := client.NewClient(transportClient)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer mcpClient.Close()

	// 获取工具列表
	toolsResult, err := mcpClient.ListTools(context.Background())
	if err != nil {
		log.Fatalf("获取工具失败: %v", err)
	}

	fmt.Println("🔍 可用工具:")
	for _, tool := range toolsResult.Tools {
		fmt.Printf(" - %s: %s\n", tool.Name, tool.Description)
	}

	// 🔹 工具1: list_all_datasets
	callTool(mcpClient, "list_all_datasets", map[string]interface{}{
		"page":      1,
		"page_size": 10,
		"orderby":   "create_time",
		"desc":      true,
		"name":      "",
		"id":        "",
	})

	// 🔹 工具2: list_all_documents
	callTool(mcpClient, "list_all_documents", map[string]interface{}{
		"page":          1,
		"page_size":     20,
		"orderby":       "create_time",
		"desc":          true,
		"dataset_id":    "",
		"keywords":      "",
		"document_id":   "",
		"document_name": "",
	})

	// 🔹 工具3: are_all_done
	callTool(mcpClient, "are_all_done", map[string]interface{}{})
}

func callTool(c *client.Client, toolName string, args map[string]interface{}) {
	fmt.Printf("\n🛠️ 正在调用工具: %s\n", toolName)

	body, _ := json.Marshal(args)
	req := &protocol.CallToolRequest{
		Name:         toolName,
		RawArguments: body,
	}

	result, err := c.CallTool(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 调用工具失败: %v\n", err)
		return
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*protocol.TextContent); ok {
			fmt.Println(" =>", textContent.Text)
		} else {
			fmt.Printf(" => 不支持的内容类型: %T\n", content)
		}
	}
}

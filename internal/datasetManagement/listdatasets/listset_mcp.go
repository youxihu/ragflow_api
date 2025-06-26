package listdatasets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"ragflow_api/internal/str"
)

// 请求参数结构体
type ListDatasetsInput struct {
	Page     int    `json:"page" description:"页码"`
	PageSize int    `json:"page_size" description:"每页数量"`
	OrderBy  string `json:"orderby" description:"排序字段"`
	Desc     bool   `json:"desc" description:"是否降序"`
	Name     string `json:"name" description:"按名称过滤"`
	ID       string `json:"id" description:"按ID过滤"`
}

// 注册工具到 MCP Server 中
func RegisterDatasetListTool(mcpServer *server.Server, lister *DatasetLister) error {
	tool, err := protocol.NewTool("list_all_datasets", "列出RagFlow中的所有数据集", ListDatasetsInput{})
	if err != nil {
		return err
	}

	mcpServer.RegisterTool(tool, func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		var input ListDatasetsInput
		if err := protocol.VerifyAndUnmarshal(req.RawArguments, &input); err != nil {
			return nil, err
		}

		params := lister.buildQueryParams(
			input.Page,
			input.PageSize,
			input.OrderBy,
			input.Desc,
			input.Name,
			input.ID,
		)

		reqURL := fmt.Sprintf("http://%s:%d/api/v1/datasets?%s",
			lister.cfg.RagFlow.Address,
			lister.cfg.RagFlow.Port,
			params.Encode())

		httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+lister.cfg.RagFlow.APIKey)

		resp, err := lister.sendRequest(httpReq)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		var result str.DatasetListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("JSON 解析失败: %v", err)
		}

		if result.Code != 0 {
			return nil, fmt.Errorf("❌ 获取数据集失败: %s (code=%d)", result.Message, result.Code)
		}

		var contents []protocol.Content
		for _, ds := range result.Data {
			contents = append(contents, &protocol.TextContent{
				Type: "text",
				Text: fmt.Sprintf("ID: %s | Name: %s", ds.ID, ds.Name),
			})
		}

		return &protocol.CallToolResult{Content: contents}, nil
	})

	return nil
}

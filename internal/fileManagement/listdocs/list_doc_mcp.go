package listdocs

import (
	"context"
	"fmt"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"ragflow_api/internal/str"
)

type ListDocsInput struct {
	Page         int    `json:"page" description:"页码"`
	PageSize     int    `json:"page_size" description:"每页数量"`
	OrderBy      string `json:"orderby" description:"排序字段"`
	Desc         bool   `json:"desc" description:"是否降序"`
	DatasetID    string `json:"dataset_id" description:"指定数据集ID（可选）"`
	Keywords     string `json:"keywords" description:"按关键字过滤"`
	DocumentID   string `json:"document_id" description:"按文档ID过滤"`
	DocumentName string `json:"document_name" description:"按文档名称过滤"`
}

func RegisterDocListTool(mcpServer *server.Server, lister *DocLister) error {
	tool, err := protocol.NewTool("list_all_documents", "列出指定数据集的文档列表", ListDocsInput{})
	if err != nil {
		return err
	}

	mcpServer.RegisterTool(tool, func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		var input ListDocsInput
		if err := protocol.VerifyAndUnmarshal(req.RawArguments, &input); err != nil {
			return nil, err
		}

		var result *str.DocumentListResponse
		var err error

		if input.DatasetID != "" {
			// 指定数据集，直接调用对应方法
			result, err = lister.GetListResultByDatasetID(
				input.DatasetID, input.Page, input.PageSize, input.OrderBy,
				input.Desc, input.Keywords, input.DocumentID, input.DocumentName,
			)
			if err != nil {
				return nil, err
			}
		} else {
			// 不传 datasetID，遍历所有数据集，依次请求，合并结果
			var mergedDocs []str.Document
			total := 0
			for _, dsID := range lister.cfg.RagFlow.DatasetID {
				resp, err := lister.GetListResultByDatasetID(dsID, input.Page, input.PageSize, input.OrderBy, input.Desc, input.Keywords, input.DocumentID, input.DocumentName)
				if err != nil {
					return nil, err
				}
				total += resp.Data.Total
				mergedDocs = append(mergedDocs, resp.Data.Docs...)
			}

			result = &str.DocumentListResponse{
				Code:    0,
				Message: "success",
				Data: str.DocData{
					Total: total,
					Docs:  mergedDocs,
				},
			}
		}

		if err != nil {
			return nil, err
		}
		if result.Code != 0 {
			return nil, fmt.Errorf("获取失败: %s", result.Message)
		}

		var contents []protocol.Content
		for _, doc := range result.Data.Docs {
			text := fmt.Sprintf("📄|知识库: %s| 文档: %s | 状态: %s | 分片数: %d | 创建: %s",
				doc.Name, doc.Run, doc.ChunkCount, doc.CreateDate)
			contents = append(contents, &protocol.TextContent{Type: "text", Text: text})
		}

		return &protocol.CallToolResult{Content: contents}, nil
	})
	return nil
}

func RegisterAreAllDoneTool(mcpServer *server.Server, lister *DocLister) error {
	tool, err := protocol.NewTool("are_all_done", "检查所有数据集文档是否已完成解析", struct{}{})
	if err != nil {
		return err
	}

	mcpServer.RegisterTool(tool, func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		done, err := lister.AreAllDatasetsDone()
		if err != nil {
			return nil, err
		}

		text := "✅ 所有文档已完成解析（或失败）"
		if !done {
			text = "⏳ 存在未完成解析的文档"
		}

		return &protocol.CallToolResult{
			Content: []protocol.Content{
				&protocol.TextContent{Type: "text", Text: text},
			},
		}, nil
	})
	return nil
}

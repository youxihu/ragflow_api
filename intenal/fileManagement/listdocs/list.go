package listdocs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ragflow_api/intenal/str"
)

// DocLister 用于获取文档列表
type DocLister struct {
	cfg str.RagConf
}

func NewDocLister(cfg str.RagConf) *DocLister {
	return &DocLister{cfg: cfg}
}

// Start 启动文档列表查询流程
func (d *DocLister) Start(page, pageSize int, orderby string, desc bool, keywords, documentID, documentName string) error {
	// 遍历所有 DatasetID
	for _, datasetID := range d.cfg.RagFlow.DatasetID {
		log.Printf("🔄 正在查询 dataset [%s] 的文档列表", datasetID)

		// 构造请求
		req, err := d.buildRequest(page, pageSize, orderby, desc, keywords, documentID, documentName)
		if err != nil {
			return err
		}

		// 设置当前 datasetID 到 URL
		req.URL.Path = fmt.Sprintf("/api/v1/datasets/%s/documents", datasetID)

		// 发送请求
		resp, err := d.sendRequest(req)
		if err != nil {
			return err
		}

		// 处理响应
		if err := d.handleResponse(resp); err != nil {
			return err
		}

	}

	return nil
}

// buildRequest 构建 GET 请求
func (l *DocLister) buildRequest(
	page, pageSize int,
	orderby string,
	desc bool,
	keywords, documentID, documentName string,
) (*http.Request, error) {

	url := fmt.Sprintf(
		"http://%s:%d/api/v1/datasets/%s/documents?page=%d&page_size=%d&orderby=%s&desc=%t&keywords=%s&id=%s&name=%s",
		l.cfg.RagFlow.Address,
		l.cfg.RagFlow.Port,
		l.cfg.RagFlow.DatasetID,
		page,
		pageSize,
		orderby,
		desc,
		keywords,
		documentID,
		documentName,
	)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.cfg.RagFlow.APIKey)

	return req, nil
}

// sendRequest 发送 HTTP 请求
func (l *DocLister) sendRequest(req *http.Request) (*http.Response, error) {
	timeout, err := time.ParseDuration(l.cfg.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return client.Do(req)
}

// handleResponse 处理响应数据
func (l *DocLister) handleResponse(resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result str.DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("JSON 解析失败: %v", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("接口返回错误: %s", result.Message)
	}

	log.Printf("✅ 成功获取文档列表，共 %d 条记录", result.Data.Total)
	for _, doc := range result.Data.Docs {
		log.Printf("- 文档名称: %s | 分片数量: %d | 状态: %s | 创建时间: %s", doc.Name, doc.ChunkCount, doc.Run, doc.CreateDate)
	}

	return nil
}

// GetListResult ==============用于ParseDoc=============================
// 返回 DocumentListResponse
func (l *DocLister) GetListResult(page, pageSize int, orderby string, desc bool, keywords, documentID, documentName string) (*str.DocumentListResponse, error) {
	req, err := l.buildRequest(page, pageSize, orderby, desc, keywords, documentID, documentName)
	if err != nil {
		return nil, err
	}

	resp, err := l.sendRequest(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result str.DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("接口返回错误: %s", result.Message)
	}

	return &result, nil
}

// GetListResultByDatasetID 根据指定 datasetID 获取文档列表
func (l *DocLister) GetListResultByDatasetID(
	datasetID string,
	page, pageSize int,
	orderby string,
	desc bool,
	keywords, documentID, documentName string,
) (*str.DocumentListResponse, error) {

	url := fmt.Sprintf(
		"http://192.168.0.43:9381/api/v1/datasets/%s/documents?page=%d&page_size=%d&orderby=%s&desc=%t&keywords=%s&id=%s&name=%s",
		datasetID,
		page,
		pageSize,
		orderby,
		desc,
		keywords,
		documentID,
		documentName,
	)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.cfg.RagFlow.APIKey)

	resp, err := l.sendRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result str.DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("接口返回错误: %v", result.Message)
	}

	return &result, nil
}

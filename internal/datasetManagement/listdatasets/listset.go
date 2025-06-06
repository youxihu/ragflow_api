package listdatasets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"ragflow_api/internal/str"
	"time"
)

type DatasetLister struct {
	cfg str.RagConf
}

func NewDatasetLister(cfg str.RagConf) *DatasetLister {
	return &DatasetLister{cfg: cfg}
}

func (l *DatasetLister) Start() error {
	log.Println("🔄 开始列出数据集...")

	params := l.buildQueryParams(1, 10, "create_time", true, "", "")
	reqURL := fmt.Sprintf("http://%s:%d/api/v1/datasets?%s",
		l.cfg.RagFlow.Address,
		l.cfg.RagFlow.Port,
		params.Encode())

	req, err := http.NewRequestWithContext(context.Background(), "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+l.cfg.RagFlow.APIKey)

	resp, err := l.sendRequest(req)
	if err != nil {
		return err
	}

	return l.handleResponse(resp)
}

func (l *DatasetLister) buildQueryParams(
	page int,
	pageSize int,
	orderBy string,
	desc bool,
	name string,
	id string,
) url.Values {
	params := url.Values{}

	if page > 0 {
		params.Add("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Add("page_size", fmt.Sprintf("%d", pageSize))
	}
	if orderBy != "" {
		params.Add("orderby", orderBy)
	}
	params.Add("desc", fmt.Sprintf("%t", desc))
	if name != "" {
		params.Add("name", name)
	}
	if id != "" {
		params.Add("id", id)
	}

	return params
}

func (l *DatasetLister) sendRequest(req *http.Request) (*http.Response, error) {
	timeout := 30 * time.Second
	if l.cfg.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(l.cfg.Timeout)
		if err != nil {
			log.Printf("⚠️ 超时设置错误，使用默认值: %v", timeout)
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return client.Do(req)
}

func (l *DatasetLister) handleResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}
	defer resp.Body.Close()

	var result str.DatasetListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("JSON 解析失败: %v", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("❌ 获取数据集失败: %s (code=%d)", result.Message, result.Code)
	}

	log.Printf("✅ 成功获取 %d 个数据集", len(result.Data))
	for _, ds := range result.Data {
		log.Printf("ID: %s | Name: %s ", ds.ID, ds.Name)
	}

	return nil
}

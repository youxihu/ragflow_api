package parsedocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ragflow_api/intenal/fileManagement/listdocs"
	"ragflow_api/intenal/str"
)

type DocParser struct {
	cfg str.RagConf
}

func NewDocParser(cfg str.RagConf) *DocParser {
	return &DocParser{cfg: cfg}
}

// ParseZeroChunkDocs 获取每个 dataset 中 chunk_count == 0 的文档，并触发解析（并发版本）
func (p *DocParser) ParseZeroChunkDocs() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(p.cfg.RagFlow.DatasetID))

	for _, datasetID := range p.cfg.RagFlow.DatasetID {
		wg.Add(1)
		datasetID := datasetID // 避免闭包问题

		go func() {
			defer wg.Done()

			log.Printf("🔄 正在处理 dataset [%s]", datasetID)

			lister := listdocs.NewDocLister(p.cfg)

			resp, err := lister.GetListResultByDatasetID(
				datasetID,
				1, 30,
				"create_time", true,
				"", "", "",
			)
			if err != nil {
				log.Printf("❌ dataset [%s] 获取文档列表失败: %v", datasetID, err)
				errChan <- fmt.Errorf("dataset [%s] 获取文档列表失败: %v", datasetID, err)
				return
			}

			var documentIDs []string
			for _, doc := range resp.Data.Docs {
				if doc.ChunkCount == 0 {
					documentIDs = append(documentIDs, doc.ID)
					log.Printf("📌 dataset [%s] 发现未解析文档: %s", datasetID, doc.Name)
				}
			}

			if len(documentIDs) == 0 {
				log.Printf("✅ dataset [%s] 所有文档均已解析，无需操作", datasetID)
				return
			}

			log.Printf("📤 dataset [%s] 即将解析 %d 个未处理文档...", datasetID, len(documentIDs))
			err = p.parseDocuments(datasetID, documentIDs)
			if err != nil {
				log.Printf("❌ dataset [%s] 文档解析失败: %v", datasetID, err)
				errChan <- fmt.Errorf("dataset [%s] 解析失败: %v", datasetID, err)
				return
			}

			log.Printf("🎉 dataset [%s] 成功触发 %d 个文档的解析任务", datasetID, len(documentIDs))
		}()
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for e := range errChan {
		if e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分 dataset 解析失败: %v", errs)
	}

	return nil
}

// parseDocuments 向 RagFlow 接口发送解析请求
func (p *DocParser) parseDocuments(datasetID string, documentIDs []string) error {
	url := fmt.Sprintf(
		"http://%s:%d/api/v1/datasets/%s/chunks",
		p.cfg.RagFlow.Address,
		p.cfg.RagFlow.Port,
		datasetID,
	)

	reqBody := str.ParseRequest{
		DocumentIDs: documentIDs,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(reqBody); err != nil {
		return fmt.Errorf("构建请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.RagFlow.APIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("JSON 解析失败: %v", err)
	}

	code, ok := result["code"].(float64)
	if !ok || code != 0 {
		message := result["message"]
		return fmt.Errorf("接口返回错误: %v", message)
	}

	return nil
}

package parsedocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ragflow_api/internal/fileManagement/listdocs"
	"ragflow_api/internal/str"
	"sync"
	"time"
)

type DocParsStopper struct {
	cfg str.RagConf
}

func NewDocParsStopper(cfg str.RagConf) *DocParsStopper {
	return &DocParsStopper{cfg: cfg}
}

// StopRunningDocuments 停止所有 dataset 中处于运行状态的文档解析任务
func (s *DocParsStopper) StopRunningDocuments() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(s.cfg.RagFlow.DatasetID))

	for _, datasetID := range s.cfg.RagFlow.DatasetID {
		wg.Add(1)
		datasetID := datasetID // 避免闭包问题

		go func() {
			defer wg.Done()

			log.Printf("🔄 正在检查 dataset [%s] 中正在解析的文档...", datasetID)

			lister := listdocs.NewDocLister(s.cfg)

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
				if isProcessingStatus(doc.Run) {
					documentIDs = append(documentIDs, doc.ID)
					log.Printf("🛑 dataset [%s] 发现正在解析的文档: %s (%s)", datasetID, doc.Name, doc.Run)
				}
			}

			if len(documentIDs) == 0 {
				log.Printf("✅ dataset [%s] 没有正在解析的文档", datasetID)
				return
			}

			log.Printf("⛔ dataset [%s] 即将停止 %d 个正在解析的文档...", datasetID, len(documentIDs))
			err = s.stopParsingDocuments(datasetID, documentIDs)
			if err != nil {
				log.Printf("❌ dataset [%s] 停止解析失败: %v", datasetID, err)
				errChan <- fmt.Errorf("dataset [%s] 停止失败: %v", datasetID, err)
				return
			}

			log.Printf("🎉 dataset [%s] 成功停止 %d 个文档的解析任务", datasetID, len(documentIDs))
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
		return fmt.Errorf("部分 dataset 停止解析失败: %v", errs)
	}

	return nil
}

// stopParsingDocuments 向 RagFlow 接口发送 DELETE 请求，停止文档解析
func (s *DocParsStopper) stopParsingDocuments(datasetID string, documentIDs []string) error {
	url := fmt.Sprintf(
		"http://%s:%d/api/v1/datasets/%s/chunks",
		s.cfg.RagFlow.Address,
		s.cfg.RagFlow.Port,
		datasetID,
	)

	reqBody := struct {
		DocumentIDs []string `json:"document_ids"`
	}{
		DocumentIDs: documentIDs,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(reqBody); err != nil {
		return fmt.Errorf("构建请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "DELETE", url, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.RagFlow.APIKey)

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

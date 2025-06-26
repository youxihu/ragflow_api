package parsedocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	listdocp "ragflow_api/internal/fileManagement/listdocs"
	"time"

	"ragflow_api/internal/str"
)

type DocParser struct {
	cfg str.RagConf
}

// ParseConfig 解析任务配置
type ParseConfig struct {
	BatchSize   int           // 每批最大文档数
	MaxRetries  int           // 最大重试次数
	Timeout     time.Duration // 单个文档等待超时时间
	RetryWait   time.Duration // 重试间隔
	StopOnError bool          // 遇到失败是否停止整个流程
}

// 默认配置
var DefaultParseConfig = ParseConfig{
	BatchSize:   20,
	MaxRetries:  3,
	Timeout:     4 * time.Hour,
	RetryWait:   10 * time.Second,
	StopOnError: false,
}

func NewDocParser(cfg str.RagConf) *DocParser {
	return &DocParser{cfg: cfg}
}

// ParseZeroChunkDocs 按 dataset_id 串行解析文档，一个 dataset 全部解析完成后才开始下一个
func (p *DocParser) ParseZeroChunkDocs() error {
	cfg := DefaultParseConfig
	for _, datasetID := range p.cfg.RagFlow.DatasetID {
		log.Printf("🔄 正在处理 dataset [%s]", datasetID)

		// 1. 获取文档列表
		lister := listdocp.NewDocLister(p.cfg)
		resp, err := lister.GetListResultByDatasetID(datasetID, 1, 100, "create_time", true, "", "", "")
		if err != nil {
			log.Printf("❌ dataset [%s] 获取文档列表失败: %v", datasetID, err)
			if cfg.StopOnError {
				return fmt.Errorf("dataset [%s] 获取文档列表失败: %v", datasetID, err)
			}
			continue
		}

		// 2. 过滤出需要解析的文档
		var documentIDs []string
		for _, doc := range resp.Data.Docs {
			if doc.ChunkCount == 0 && !isProcessingStatus(doc.Run) {
				documentIDs = append(documentIDs, doc.ID)
				log.Printf("📌 dataset [%s] 发现未解析文档（ChunkCount == 0）: %s", datasetID, doc.Name)
			} else if doc.ChunkCount > 0 && !isStableStatus(doc.Run) && !isProcessingStatus(doc.Run) {
				documentIDs = append(documentIDs, doc.ID)
				log.Printf("📌 dataset [%s] 发现需重新解析文档（状态: %s）: %s", datasetID, doc.Run, doc.Name)
			}
		}

		if len(documentIDs) == 0 {
			log.Printf("✅ dataset [%s] 所有文档均已解析或状态正常，无需操作", datasetID)
			continue
		}

		log.Printf("📄 dataset [%s] 共发现 %d 个待解析文档，将按每批 %d 个进行处理", datasetID, len(documentIDs), cfg.BatchSize)

		// 3. 分批次处理
		for i := 0; i < len(documentIDs); i += cfg.BatchSize {
			end := i + cfg.BatchSize
			if end > len(documentIDs) {
				end = len(documentIDs)
			}
			batch := documentIDs[i:end]

			log.Printf("📤 dataset [%s] 即将解析第 %d - %d 个文档...", datasetID, i+1, end)

			// 3.1 再次检查文档状态
			finalBatch := p.filterNonProcessingDocs(datasetID, batch)
			if len(finalBatch) == 0 {
				log.Printf("⚠️ dataset [%s] 第 %d - %d 个文档都在解析中，跳过本次请求", datasetID, i+1, end)
				continue
			}

			// 3.2 尝试发送解析请求
			err := p.parseWithRetry(datasetID, finalBatch, cfg.MaxRetries, cfg.RetryWait)
			if err != nil {
				log.Printf("❌ dataset [%s] 第 %d - %d 个文档多次尝试失败，仍无法解析: %v", datasetID, i+1, end, err)
				if cfg.StopOnError {
					return fmt.Errorf("dataset [%s] 解析失败: %v", datasetID, err)
				}
				continue
			}

			// 3.3 等待完成
			log.Printf("⏳ dataset [%s] 等待第 %d - %d 个文档解析完成...", datasetID, i+1, end)
			err = p.waitForDocumentsDone(datasetID, finalBatch, cfg.Timeout, 10*time.Second)
			if err != nil {
				log.Printf("❌ dataset [%s] 第 %d - %d 个文档解析失败或超时: %v", datasetID, i+1, end, err)
				if cfg.StopOnError {
					return fmt.Errorf("dataset [%s] 解析失败: %v", datasetID, err)
				}
			}

			log.Printf("✅ dataset [%s] 第 %d - %d 个文档已成功解析", datasetID, i+1, end)
		}

		log.Printf("🎉 dataset [%s] 成功完成全部文档的解析任务", datasetID)
	}

	return nil
}

// parseWithRetry 加入重试机制
func (p *DocParser) parseWithRetry(datasetID string, docIDs []string, maxRetries int, retryWait time.Duration) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("🔁 第 %d/%d 次尝试解析文档: %v", attempt, maxRetries, docIDs)
		err := p.parseDocuments(datasetID, docIDs)
		if err == nil {
			return nil
		}

		// 打印失败文档
		failedDocs, _ := p.getFailedDocumentIDs(datasetID, docIDs)
		log.Printf("📌 第 %d 次尝试失败，失败文档: %v", attempt, failedDocs)

		if attempt < maxRetries {
			log.Printf("🕒 正在等待 %s 后重试...", retryWait)
			time.Sleep(retryWait)
		}
	}
	return fmt.Errorf("多次尝试失败")
}

// getFailedDocumentIDs 返回当前处于 RUNNING/PROCESSING 状态的文档
func (p *DocParser) getFailedDocumentIDs(datasetID string, docIDs []string) ([]string, error) {
	lister := listdocp.NewDocLister(p.cfg)
	resp, err := lister.GetListResultByDatasetID(datasetID, 1, 100, "create_time", true, "", "", "")
	if err != nil {
		return nil, err
	}

	docMap := make(map[string]string)
	for _, doc := range resp.Data.Docs {
		docMap[doc.ID] = doc.Run
	}

	var failedDocs []string
	for _, id := range docIDs {
		runStatus := docMap[id]
		if isProcessingStatus(runStatus) {
			failedDocs = append(failedDocs, id)
		}
	}

	return failedDocs, nil
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

// isStableStatus 判断文档是否处于稳定状态（无需重新解析）
func isStableStatus(status string) bool {
	switch status {
	case "DONE", "FAIL":
		return true
	default:
		return false
	}
}

// isProcessingStatus 判断文档是否处于解析过程中
func isProcessingStatus(status string) bool {
	switch status {
	case "RUNNING", "PROCESSING", "WAITING", "PENDING":
		return true
	default:
		return false
	}
}

// waitForDocumentsDone 轮询指定文档是否进入稳定状态（DONE/FAIL）
func (p *DocParser) waitForDocumentsDone(datasetID string, documentIDs []string, timeout, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取最新文档信息
			lister := listdocp.NewDocLister(p.cfg)
			resp, err := lister.GetListResultByDatasetID(datasetID, 1, 100, "create_time", true, "", "", "")
			if err != nil {
				continue
			}

			running := make(map[string]bool)
			for _, doc := range resp.Data.Docs {
				if isProcessingStatus(doc.Run) {
					running[doc.ID] = true
				}
			}

			completed := true
			for _, id := range documentIDs {
				if running[id] {
					completed = false
					break
				}
			}

			if completed {
				return nil
			}

		case <-timer.C:
			statuses := make(map[string]string)
			lister := listdocp.NewDocLister(p.cfg)
			resp, _ := lister.GetListResultByDatasetID(datasetID, 1, 100, "create_time", true, "", "", "")
			for _, doc := range resp.Data.Docs {
				statuses[doc.ID] = doc.Run
			}
			return fmt.Errorf("等待超时，部分文档仍在运行: %v", statuses)
		}
	}
}

// filterNonProcessingDocs 过滤出不在解析状态的文档 ID
func (p *DocParser) filterNonProcessingDocs(datasetID string, docIDs []string) []string {
	lister := listdocp.NewDocLister(p.cfg)
	resp, err := lister.GetListResultByDatasetID(datasetID, 1, 100, "create_time", true, "", "", "")
	if err != nil {
		log.Printf("⚠️ dataset [%s] 获取文档状态失败: %v，跳过过滤", datasetID, err)
		return docIDs
	}

	docMap := make(map[string]string)
	for _, doc := range resp.Data.Docs {
		docMap[doc.ID] = doc.Run
	}

	var result []string
	for _, id := range docIDs {
		runStatus := docMap[id]
		if !isProcessingStatus(runStatus) {
			result = append(result, id)
		} else {
			log.Printf("📌 文档 [%s] 当前状态为 [%s]，跳过解析", id, runStatus)
		}
	}

	return result
}

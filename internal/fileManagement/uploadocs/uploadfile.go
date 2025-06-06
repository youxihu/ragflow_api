package uploadocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"ragflow_api/internal/str"
	"time"
)

// DocUploader 上传
type DocUploader struct {
	cfg str.RagConf
}

func NewDocUploader(cfg str.RagConf, dirPath string, timeout string) *DocUploader {
	cfg.DirPath = dirPath
	cfg.Timeout = timeout
	return &DocUploader{cfg: cfg}
}

func (u *DocUploader) Start() error {
	// 遍历所有 DatasetID
	for _, datasetID := range u.cfg.RagFlow.DatasetID {
		log.Printf("🔄 正在上传到 dataset [%s]", datasetID)

		// 构造请求体
		body, ct, err := u.prepareRequestBody()
		if err != nil {
			return err
		}

		// 构造请求
		req, err := u.buildRequest(body, ct)
		if err != nil {
			return err
		}

		// 设置当前 datasetID 到 URL
		req.URL.Path = fmt.Sprintf("/api/v1/datasets/%s/documents", datasetID)

		// 发送请求
		resp, err := u.sendRequest(req)
		if err != nil {
			return err
		}

		// 处理响应
		if err := u.handleResponse(resp); err != nil {
			return fmt.Errorf("dataset [%s] 上传失败: %v", datasetID, err)
		}
	}
	return nil
}

func (u *DocUploader) prepareRequestBody() (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	files, err := os.ReadDir(u.cfg.DirPath)
	if err != nil {
		return nil, "", fmt.Errorf("无法读取目录: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || file.Name()[0] == '.' {
			continue
		}

		fullPath := filepath.Join(u.cfg.DirPath, file.Name())
		f, err := os.Open(fullPath)
		if err != nil {
			return nil, "", fmt.Errorf("无法打开文件 %s: %v", fullPath, err)
		}
		defer f.Close()

		part, err := writer.CreateFormFile("file", file.Name())
		if err != nil {
			return nil, "", fmt.Errorf("创建表单文件失败: %v", err)
		}

		if _, err := io.Copy(part, f); err != nil {
			return nil, "", fmt.Errorf("复制文件内容失败: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("关闭 multipart writer 失败: %v", err)
	}

	return body, writer.FormDataContentType(), nil
}

func (u *DocUploader) buildRequest(body *bytes.Buffer, contentType string) (*http.Request, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/datasets/%s/documents", u.cfg.RagFlow.Address, u.cfg.RagFlow.Port, u.cfg.RagFlow.DatasetID)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+u.cfg.RagFlow.APIKey)

	return req, nil
}

func (u *DocUploader) sendRequest(req *http.Request) (*http.Response, error) {
	timeout, err := time.ParseDuration(u.cfg.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return client.Do(req)
}

func (u *DocUploader) handleResponse(resp *http.Response) error {
	respBody := new(bytes.Buffer)
	if _, err := respBody.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result str.UploadResponse
	if err := json.Unmarshal(respBody.Bytes(), &result); err != nil {
		return fmt.Errorf("JSON 解析失败: %v", err)
	}

	if result.Code == 0 {
		log.Println("✅ 文件上传成功")
		return nil
	}

	return fmt.Errorf("❌ 文件上传失败: %s", result.Message)
}

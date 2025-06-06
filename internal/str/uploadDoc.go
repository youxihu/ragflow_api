package str

// UploadResponse 用于解析服务器返回的 JSON 响应
type UploadResponse struct {
	Code    float64     `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

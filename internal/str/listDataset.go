package str

// Dataset 是单个数据集的结构
type Dataset struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Status        string                 `json:"status"`
	ChunkMethod   string                 `json:"chunk_method"`
	CreateDate    string                 `json:"create_date"`
	CreateTime    int64                  `json:"create_time"`
	UpdateDate    string                 `json:"update_date"`
	UpdateTime    int64                  `json:"update_time"`
	DocumentCount int                    `json:"document_count"`
	TokenNum      int                    `json:"token_num"`
	ParserConfig  map[string]interface{} `json:"parser_config,omitempty"`
}

// DatasetListResponse 是接口返回的整体结构
type DatasetListResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message,omitempty"`
	Data    []Dataset `json:"data,omitempty"`
}

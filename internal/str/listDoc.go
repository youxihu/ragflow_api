package str

type DocumentListResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message,omitempty"` // 可能为空或省略
	Data    DocData `json:"data"`
}

type DocData struct {
	Docs  []Document `json:"docs"`
	Total int        `json:"total"`
}

type Document struct {
	ChunkCount      int          `json:"chunk_count"`
	CreateDate      string       `json:"create_date"`
	CreateTime      int64        `json:"create_time"`
	CreatedBy       string       `json:"created_by"`
	ID              string       `json:"id"`
	KnowledgeBaseID string       `json:"knowledgebase_id"`
	Location        string       `json:"location"`
	Name            string       `json:"name"`
	ParserConfig    ParserConfig `json:"parser_config"`
	ChunkMethod     string       `json:"chunk_method"`
	ProcessBeginAt  interface{}  `json:"process_begin_at"` // 可为 null
	ProcessDuration float64      `json:"process_duation"`  // 注意字段名是 duation（拼写错误）
	Progress        float64      `json:"progress"`
	ProgressMsg     string       `json:"progress_msg"`
	Run             string       `json:"run"`
	Size            int          `json:"size"`
	SourceType      string       `json:"source_type"`
	Status          string       `json:"status"`
	Thumbnail       interface{}  `json:"thumbnail"` // 可为 null
	TokenCount      int          `json:"token_count"`
	Type            string       `json:"type"`
	UpdateDate      string       `json:"update_date"`
	UpdateTime      int64        `json:"update_time"`
}

type ParserConfig struct {
	ChunkTokenCount int    `json:"chunk_token_count"`
	Delimiter       string `json:"delimiter"`
	LayoutRecognize string `json:"layout_recognize"`
	TaskPageSize    int    `json:"task_page_size"`
}

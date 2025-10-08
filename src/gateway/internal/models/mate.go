package models

type ChatReq struct {
	Prompt string `json:"prompt" binding:"required"`
}

type PageResp struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	SegmentID   uint   `json:"segment_id"`
	UserInput   string `json:"user_input"`
	AgentOutput string `json:"agent_output"`
	Status      string `json:"status"`
	CreateTime  int64  `json:"create_time"`
}

type GetUserPagesRequest struct {
	UserID   int    `json:"user_id"`
	Cursor   string `json:"cursor"`
	PageSize int    `json:"page_size"`
}

type GetUserPagesResponse struct {
	Pages      []PageResp `json:"pages"`
	NextCursor string     `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}

type ChatStreamChunk struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	MessageID  string `json:"message_id"`
	Timestamp  int64  `json:"timestamp"`
	Finished   bool   `json:"finished,omitempty"`
	Error      string `json:"error,omitempty"`
}

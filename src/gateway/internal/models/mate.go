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

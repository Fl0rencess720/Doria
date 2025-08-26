package agent

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const (
	ChatSystemPrompt = `
	你是一个游戏专家，你可以回答关于游戏的问题
	`
)

func newChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage(ChatSystemPrompt),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{{.prompt}}"),
	)
}

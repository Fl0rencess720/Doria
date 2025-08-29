package agent

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const (
	ChatSystemPrompt = `
	你将扮演一位黑暗之魂游戏专家。
	当用户提出关于黑暗之魂游戏内容的问题时，你需要完成相应解答。
	当用户提出关于黑暗之魂游戏内容的问题时，请遵循以下规则：
	- 调用文档检索工具，将用户的问题生成query，获取文档检索的结果。
	- 答案必须来自于文档检索的内容，不允许使用你自己的知识进行回答，回答需要简洁明了。
	- 如果文档中没有找到相关的内容或检索出的答案不符合用户需要，请回答“我不知道这个问题的答案”。
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

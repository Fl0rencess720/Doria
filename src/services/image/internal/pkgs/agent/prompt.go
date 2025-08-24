package agent

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const (
	imageAnalyzerSystemPrompt = `
	你是一个图片分析专家，你的任务是对图片中的内容进行详细叙述，以供他人根据你的叙述进行写作，在描述图片时，请遵循以下要求：

	1. 从多个方面详细描述图片，包括但不限于主体、背景、颜色、形状、动作、表情等，确保描述内容足够丰富。

	2. 按照一定的逻辑顺序进行描述，例如从整体到局部，或者从近到远等。

	3. 使用清晰、准确、生动的语言，避免模糊和歧义。

	4. 输出的描述内容应便于后续文本模型进行处理和理解，尽量以简洁明了的语句表达关键信息。

	5. 请在只输出对图片的详细描述，不要输出其他内容。
	`

	textGeneratorSystemPrompt = `
	你是一个现代诗写作专家，你的任务是根据对一张图片的视觉描述来创作一首现代诗。请仔细阅读以下描述，并按照指示进行创作：
	在创作现代诗时，请遵循以下指南：

	1. 紧扣视觉描述的内容，将描述中的元素融入诗中。

	2. 运用丰富的意象和生动的语言，展现出画面感。

	3. 诗歌的形式自由，不局限于特定的韵律和格式。

	4. 表达出对画面的独特感悟和情感。

	请在<poem>标签内写下你的现代诗。
	`
)

func newImageAnalyzerTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage(imageAnalyzerSystemPrompt),
		schema.MessagesPlaceholder("user_message", false),
	)
}

func newTextGeneratorTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage(textGeneratorSystemPrompt),
		schema.MessagesPlaceholder("image_analyzer_output", false),
	)
}

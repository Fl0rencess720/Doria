package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type state struct {
}

func buildTextGeneratorGraph(ctx context.Context, imageCm model.ToolCallingChatModel, textCm model.ToolCallingChatModel) *compose.Graph[map[string]any, *schema.Message] {
	compose.RegisterSerializableType[state]("state")

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{}
		}))

	g.AddLambdaNode("PrepareMultiModelMessageLambda", compose.InvokableLambda(prepareMultiModelMessage))
	g.AddLambdaNode("PrepareTextGeneratorInputLambda", compose.InvokableLambda(prepareTextGeneratorInput))
	g.AddChatTemplateNode("ImageAnalyzerTpl", newImageAnalyzerTemplate())
	g.AddChatTemplateNode("TextGeneratorTpl", newTextGeneratorTemplate())

	g.AddChatModelNode("ImageAnalyzerCm", imageCm)
	g.AddChatModelNode("TextGeneratorCm", textCm)

	g.AddEdge(compose.START, "PrepareMultiModelMessageLambda")
	g.AddEdge("PrepareMultiModelMessageLambda", "ImageAnalyzerTpl")
	g.AddEdge("ImageAnalyzerTpl", "ImageAnalyzerCm")
	g.AddEdge("ImageAnalyzerCm", "PrepareTextGeneratorInputLambda")
	g.AddEdge("PrepareTextGeneratorInputLambda", "TextGeneratorTpl")
	g.AddEdge("TextGeneratorTpl", "TextGeneratorCm")
	g.AddEdge("TextGeneratorCm", compose.END)

	return g
}

func prepareMultiModelMessage(ctx context.Context, input map[string]any) (map[string]any, error) {
	imageDataURI := input["image_data_uri"].(string)

	message := &schema.Message{
		Role: schema.User,
		MultiContent: []schema.ChatMessagePart{
			{
				Type: schema.ChatMessagePartTypeImageURL,
				ImageURL: &schema.ChatMessageImageURL{
					URL:      imageDataURI,
					Detail:   schema.ImageURLDetailAuto,
					MIMEType: "image/jpeg",
				},
			},
		},
	}

	return map[string]any{
		"user_message": []*schema.Message{message},
	}, nil
}

func prepareTextGeneratorInput(ctx context.Context, input *schema.Message) (map[string]any, error) {
	return map[string]any{
		"image_analyzer_output": []*schema.Message{input},
	}, nil
}

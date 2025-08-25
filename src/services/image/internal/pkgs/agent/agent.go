package agent

import (
	"context"
	"encoding/json"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/pkgs/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type TextGeneratorResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TextGenerator struct {
	agent compose.Runnable[map[string]any, *schema.Message]
}

func NewTextGenerator(ctx context.Context) (*TextGenerator, error) {
	imageCm, err := newImageChatModel(ctx)
	if err != nil {
		return nil, err
	}
	textCm, err := newTextChatModel(ctx)
	if err != nil {
		return nil, err
	}

	g := buildTextGeneratorGraph(ctx, imageCm, textCm)
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return &TextGenerator{agent: runnable}, nil
}

func (t *TextGenerator) Generator(ctx context.Context, imageData []byte) (*TextGeneratorResponse, error) {
	imageDataURI := utils.GenImageDataURI(imageData)
	output, err := t.agent.Invoke(ctx, map[string]any{
		"image_data_uri": imageDataURI,
	})
	if err != nil {
		return nil, err
	}

	response := &TextGeneratorResponse{}
	if err = json.Unmarshal([]byte(output.Content), response); err != nil {
		return nil, err
	}

	return response, nil
}

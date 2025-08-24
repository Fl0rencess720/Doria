package biz

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/pkgs/agent"
)

type ImageRepo interface {
}

type ImageUseCase struct {
	imageRepo ImageRepo
}

func NewImageUseCase(imageRepo ImageRepo) *ImageUseCase {
	return &ImageUseCase{
		imageRepo: imageRepo,
	}
}

func (*ImageUseCase) GenerateTextOfImage(ctx context.Context, imageData []byte, style string) (string, error) {
	textGenerator, err := agent.NewTextGenerator(ctx)
	if err != nil {
		return "", err
	}
	text, err := textGenerator.Generator(ctx, imageData)
	if err != nil {
		return "", err
	}
	return text, nil
}

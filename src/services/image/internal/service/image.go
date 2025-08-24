package service

import (
	"context"

	imageapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/image"
)

func (s *ImageService) GenerateTextOnImage(ctx context.Context, req *imageapi.GenerateTextRequest) (resp *imageapi.GenerateTextResponse, err error) {
	text, err := s.imageUseCase.GenerateTextOfImage(ctx, req.ImageData, req.TextStyle)
	if err != nil {
		return nil, err
	}

	return &imageapi.GenerateTextResponse{
		Text: text,
	}, nil
}

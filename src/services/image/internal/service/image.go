package service

import (
	"context"

	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
)

func (s *ImageService) GenerateTextOfImage(ctx context.Context, req *imageapi.GenerateTextRequest) (resp *imageapi.GenerateTextResponse, err error) {
	result, err := s.imageUseCase.GenerateTextOfImage(ctx, req.ImageData, req.TextStyle)
	if err != nil {
		return nil, err
	}

	return &imageapi.GenerateTextResponse{
		Name:        result.Name,
		Description: result.Description,
	}, nil
}

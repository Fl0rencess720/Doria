package service

import (
	"errors"
	"io"

	chatapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/chat"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/chat/internal/biz"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/chat/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (s *ChatService) ChatStream(
	req *chatapi.ChatStreamRequest,
	resp grpc.ServerStreamingServer[chatapi.ChatStreamResponse],
) error {
	ctx := resp.Context()
	stream, id, err := s.chatUseCase.ChatStream(ctx, &biz.ChatStreamReq{
		UserID:         uint(req.UserId),
		ConversationID: uint(req.ConversationId),
		Prompt:         req.Prompt,
	})
	if err != nil {
		zap.L().Error("chat stream error", zap.Error(err))
		return err
	}

	role := ""
	fullResponse := ""

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			err2 := s.chatUseCase.CreateMessages(ctx, []*models.Message{
				{
					ConversationID: uint(id),
					Role:           "user",
					Content:        models.JSONContent{Text: req.Prompt},
				},
				{
					ConversationID: uint(id),
					Role:           role,
					Content:        models.JSONContent{Text: fullResponse},
				},
			})
			if err != nil {
				zap.L().Error("chat stream error", zap.Error(err2))
				break
			}
			break
		}
		if err != nil {
			zap.L().Error("chat stream error", zap.Error(err))
			return err
		}

		if err := resp.Send(&chatapi.ChatStreamResponse{
			Chunk:          chunk.Content,
			ConversationId: int32(id),
		}); err != nil {
			zap.L().Error("chat stream error", zap.Error(err))
			return err
		}

		if role != string(chunk.Role) {
			role = string(chunk.Role)
		}
		fullResponse += chunk.Content
	}

	return nil
}

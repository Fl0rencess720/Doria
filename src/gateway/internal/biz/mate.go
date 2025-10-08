package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type MateRepo interface {
}

type mateUseCase struct {
	repo           MateRepo
	mateClient     mateapi.MateServiceClient
	circuitBreaker *circuitbreaker.CircuitBreakerManager
}

func NewMateUsecase(repo MateRepo, mateClient mateapi.MateServiceClient, cbManager *circuitbreaker.CircuitBreakerManager) MateUseCase {
	return &mateUseCase{
		repo:           repo,
		mateClient:     mateClient,
		circuitBreaker: cbManager,
	}
}

func (u *mateUseCase) Chat(ctx context.Context, req *models.ChatReq, userID int) (string, response.ErrorCode, error) {
	result, err := u.circuitBreaker.Do(ctx, "mate-service.Chat",
		func(ctx context.Context) (any, error) {
			return u.mateClient.Chat(ctx, &mateapi.ChatRequest{
				UserId: int32(userID),
				Prompt: req.Prompt,
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("mate fallback triggered", zap.Error(err))
			return "抱歉，智能助手服务暂时不可用，请稍后再试", nil
		},
	)

	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		return "", response.ServerError, err
	}

	switch v := result.(type) {
	case *mateapi.ChatResponse:
		return v.Message, response.NoError, nil
	case string:
		return v, response.DegradedError, nil
	default:
		return "", response.ServerError, fmt.Errorf("unexpected response type")
	}
}

func (u *mateUseCase) CreateChatStream(ctx context.Context, req *models.ChatReq, userID int) (mateapi.MateService_ChatStreamClient, error) {
	result, err := u.circuitBreaker.Do(ctx, "mate-service.ChatStream",
		func(ctx context.Context) (any, error) {
			stream, err := u.mateClient.ChatStream(ctx, &mateapi.ChatRequest{
				UserId: int32(userID),
				Prompt: req.Prompt,
			})
			if err != nil {
				return nil, err
			}
			return stream, nil
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("mate stream fallback triggered", zap.Error(err))
			return &MockChatStreamClient{
				content:   "抱歉，智能助手服务暂时不可用，请稍后再试",
				messageID: "fallback-" + fmt.Sprintf("%d", time.Now().Unix()),
				finished:  false,
			}, nil
		},
	)

	if err != nil {
		zap.L().Error("create chat stream error", zap.Error(err))
		return nil, err
	}

	switch v := result.(type) {
	case *MockChatStreamClient:
		return v, nil
	case mateapi.MateService_ChatStreamClient:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected response type")
	}
}

func (u *mateUseCase) GetUserPages(ctx context.Context, req *models.GetUserPagesRequest) (*models.GetUserPagesResponse, response.ErrorCode, error) {
	result, err := u.circuitBreaker.Do(ctx, "mate-service.GetUserPages",
		func(ctx context.Context) (any, error) {
			return u.mateClient.GetUserPages(ctx, &mateapi.GetUserPagesRequest{
				UserId:   int32(req.UserID),
				Cursor:   req.Cursor,
				PageSize: int32(req.PageSize),
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("mate GetUserPages fallback triggered", zap.Error(err))
			return &models.GetUserPagesResponse{
				Pages:      []models.PageResp{},
				NextCursor: "",
				HasMore:    false,
			}, nil
		},
	)

	if err != nil {
		zap.L().Error("GetUserPages error", zap.Error(err))
		return nil, response.ServerError, err
	}

	switch v := result.(type) {
	case *mateapi.GetUserPagesResponse:
		pages := make([]models.PageResp, len(v.Pages))
		for i, page := range v.Pages {
			pages[i] = models.PageResp{
				ID:          uint(page.Id),
				UserID:      uint(page.UserId),
				SegmentID:   uint(page.SegmentId),
				UserInput:   page.UserInput,
				AgentOutput: page.AgentOutput,
				Status:      page.Status,
				CreateTime:  page.CreateTime,
			}
		}
		return &models.GetUserPagesResponse{
			Pages:      pages,
			NextCursor: v.NextCursor,
			HasMore:    v.HasMore,
		}, response.NoError, nil
	case *models.GetUserPagesResponse:
		return v, response.DegradedError, nil
	default:
		return nil, response.ServerError, fmt.Errorf("unexpected response type")
	}
}

type MockChatStreamClient struct {
	content   string
	messageID string
	finished  bool
	sent      bool
}

func (m *MockChatStreamClient) Recv() (*mateapi.ChatStreamResponse, error) {
	if m.sent {
		return &mateapi.ChatStreamResponse{
			Content:   "",
			MessageId: m.messageID,
			Timestamp: time.Now().Unix(),
			Finished:  true,
		}, nil
	}

	m.sent = true
	return &mateapi.ChatStreamResponse{
		Content:   m.content,
		MessageId: m.messageID,
		Timestamp: time.Now().Unix(),
		Finished:  false,
	}, nil
}

func (m *MockChatStreamClient) CloseSend() error {
	return nil
}

func (m *MockChatStreamClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *MockChatStreamClient) Trailer() metadata.MD {
	return nil
}

func (m *MockChatStreamClient) Context() context.Context {
	return context.Background()
}

func (m *MockChatStreamClient) SendMsg(v interface{}) error {
	return nil
}

func (m *MockChatStreamClient) RecvMsg(v interface{}) error {
	return nil
}

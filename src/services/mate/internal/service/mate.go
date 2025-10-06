package service

import (
	"context"

	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
)

func (s *MateService) Chat(ctx context.Context, req *mateapi.ChatRequest) (*mateapi.ChatResponse, error) {
	resp, err := s.mateUseCase.Chat(ctx, &biz.ChatReq{
		UserID: uint(req.UserId),
		Prompt: req.Prompt,
	})
	if err != nil {
		return nil, err
	}

	return &mateapi.ChatResponse{
		Message: resp,
	}, nil
}

func (s *MateService) GetUserPages(ctx context.Context, req *mateapi.GetUserPagesRequest) (*mateapi.GetUserPagesResponse, error) {
	pagesResp, err := s.mateUseCase.GetUserPages(ctx, &models.GetUserPagesRequest{
		UserID:   uint(req.UserId),
		Cursor:   req.Cursor,
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	resp := &mateapi.GetUserPagesResponse{
		Pages:      make([]*mateapi.Page, len(pagesResp.Pages)),
		NextCursor: pagesResp.NextCursor,
		HasMore:    pagesResp.HasMore,
	}

	for i, page := range pagesResp.Pages {
		resp.Pages[i] = &mateapi.Page{
			Id:          uint32(page.ID),
			UserId:      uint32(page.UserID),
			SegmentId:   uint32(page.SegmentID),
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			Status:      page.Status,
			CreateTime:  page.CreatedAt.Unix(),
		}
	}

	return resp, nil
}

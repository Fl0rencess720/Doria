package service

import (
	"context"

	userapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/user"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/user/internal/biz"
)

func (s *UserService) Register(ctx context.Context, req *userapi.RegisterRequest) (*userapi.RegisterResponse, error) {
	userID, code, err := s.userUseCase.Register(ctx, &biz.UserRegisterReq{
		Phone:    req.Phone,
		Code:     req.Code,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &userapi.RegisterResponse{
		UserId: int32(userID),
		Code:   int32(code),
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userapi.LoginRequest) (*userapi.LoginResponse, error) {
	userID, code, err := s.userUseCase.Login(ctx, &biz.UserLoginReq{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userapi.LoginResponse{
		UserId: int32(userID),
		Code:   int32(code),
	}, nil
}

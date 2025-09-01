package service

import (
	"context"

	userapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/user"
)

func (s *UserService) Register(ctx context.Context, req *userapi.RegisterRequest) (*userapi.RegisterResponse, error) {
	return s.userUseCase.Register(ctx, req)
}

func (s *UserService) Login(ctx context.Context, req *userapi.LoginRequest) (*userapi.LoginResponse, error) {
	return s.userUseCase.Login(ctx, req)
}
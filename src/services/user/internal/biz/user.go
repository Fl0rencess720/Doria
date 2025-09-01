package biz

import (
	"context"

	userapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/user"
)

type UserUseCase struct {
	repo UserRepository
}

type UserRepository interface {
}

func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Register(ctx context.Context, req *userapi.RegisterRequest) (*userapi.RegisterResponse, error) {
	// TODO: Implement register logic
	return &userapi.RegisterResponse{}, nil
}

func (uc *UserUseCase) Login(ctx context.Context, req *userapi.LoginRequest) (*userapi.LoginResponse, error) {
	// TODO: Implement login logic
	return &userapi.LoginResponse{}, nil
}

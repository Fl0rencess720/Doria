package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/services/user/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/user/internal/pkgs/response"
	"github.com/Fl0rencess720/Doria/src/services/user/internal/pkgs/utils"
)

type UserUseCase struct {
	repo UserRepo
}

type UserRepo interface {
	CreateUser(ctx context.Context, user *models.User) (uint, error)
	GetUser(ctx context.Context, userID uint) (*models.User, error)
	FindUser(ctx context.Context, phone string) (bool, error)
	VerifyRegisterCode(ctx context.Context, phone string, code string) (bool, error)
	VerifyUserPassword(ctx context.Context, phone string, password string) (bool, uint, error)
}

type UserRegisterReq struct {
	Phone    string
	Code     string
	Password string
}

type UserLoginReq struct {
	Phone    string
	Password string
}

func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Register(ctx context.Context, req *UserRegisterReq) (uint, uint, error) {
	findUser, err := uc.repo.FindUser(ctx, req.Phone)
	if err != nil {
		return 0, response.OtherError, err
	}

	if findUser {
		return 0, response.UserExistError, nil
	}

	verify, err := uc.repo.VerifyRegisterCode(ctx, req.Phone, req.Code)
	if err != nil {
		return 0, response.OtherError, err
	}

	if !verify {
		return 0, response.CodeError, nil
	}

	user := &models.User{
		Phone:    &req.Phone,
		Password: utils.MD5(req.Password),
		Status:   "user",
	}

	userID, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, response.OtherError, err
	}

	return userID, response.Success, nil
}

func (uc *UserUseCase) Login(ctx context.Context, req *UserLoginReq) (uint, uint, error) {
	verify, userID, err := uc.repo.VerifyUserPassword(ctx, req.Phone, utils.MD5(req.Password))
	if err != nil {
		return 0, response.UserNotExistError, err
	}
	if !verify {
		return 0, response.PasswordError, nil
	}

	return userID, response.Success, nil
}

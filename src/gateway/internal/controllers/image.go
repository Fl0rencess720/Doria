package controllers

import (
	"mime/multipart"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	imageapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/image"
	"github.com/gin-gonic/gin"
)

type ImageRepo interface {
}

type ImageUsecase struct {
	repo        ImageRepo
	imageClient imageapi.ImageServiceClient
}

type GenerateReq struct {
	Image     *multipart.FileHeader `form:"image" binding:"required"`
	TextStyle string                `form:"text_style" binding:"required"`
}

func NewImageUsecase(repo ImageRepo, imageClient imageapi.ImageServiceClient) *ImageUsecase {
	return &ImageUsecase{
		repo:        repo,
		imageClient: imageClient,
	}
}

func (u *ImageUsecase) GenerateText(c *gin.Context) {
	var req GenerateReq
	if err := c.ShouldBind(&req); err != nil {
		response.ErrorResponse(c, response.ServerError, err)
		return
	}
}

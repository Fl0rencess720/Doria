package controllers

import (
	"io"
	"mime/multipart"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	imageapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/image"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

type GeneratorResp struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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
		zap.L().Warn("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	file, err := req.Image.Open()
	if err != nil {
		zap.L().Warn("image open error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		zap.L().Warn("image read error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	resp, err := u.imageClient.GenerateTextOfImage(c, &imageapi.GenerateTextRequest{
		ImageData: imgBytes,
		TextStyle: req.TextStyle,
	})
	if err != nil {
		zap.L().Error("generate text on image failed", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, GeneratorResp{
		Name:        resp.Name,
		Description: resp.Description,
	})
}

package controllers

import "github.com/gin-gonic/gin"

type ImageRepo interface {
}

type ImageUsecase struct {
	repo ImageRepo
}

func NewImageUsecase(repo ImageRepo) *ImageUsecase {
	return &ImageUsecase{
		repo: repo,
	}
}

func (u *ImageUsecase) Generate(c *gin.Context) {

}

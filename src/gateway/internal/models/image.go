package models

import "mime/multipart"

type GenerateReq struct {
	Image     *multipart.FileHeader `form:"image" binding:"required"`
	TextStyle string                `form:"text_style" binding:"required"`
}

type GenerateResp struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

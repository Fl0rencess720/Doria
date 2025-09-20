package data

import "github.com/Fl0rencess720/Doria/src/services/image/internal/biz"

type imageRepo struct {
}

func NewImageRepo() biz.ImageRepo {
	return &imageRepo{}
}

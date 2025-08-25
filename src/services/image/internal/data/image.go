package data

import "github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/biz"

type imageRepo struct {
}

func NewImageRepo() biz.ImageRepo {
	return &imageRepo{}
}

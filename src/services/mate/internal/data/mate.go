package data

import (
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/biz"
)

type mateRepo struct {
}

func NewMateRepo() biz.MateRepo {
	return &mateRepo{}
}

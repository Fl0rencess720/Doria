package data

import "github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"

type ImageRepo struct {
}

func NewImageRepo() controllers.ImageRepo {
	return &ImageRepo{}
}

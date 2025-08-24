//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/configs"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/biz"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/data"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/internal/service"
)

type App struct {
	Server *service.ImageService
}

func NewApp(server *service.ImageService) *App {
	return &App{
		Server: server,
	}
}

func wireApp() *App {
	panic(wire.Build(
		NewApp,
		configs.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}

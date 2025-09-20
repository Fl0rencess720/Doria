//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/services/mate/configs"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/data"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/service"
)

type App struct {
	Server *service.MateService
}

func NewApp(server *service.MateService) *App {
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

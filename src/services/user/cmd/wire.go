//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/services/user/configs"
	"github.com/Fl0rencess720/Doria/src/services/user/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/user/internal/data"
	"github.com/Fl0rencess720/Doria/src/services/user/internal/service"
)

type App struct {
	Server *service.UserService
}

func NewApp(server *service.UserService) *App {
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

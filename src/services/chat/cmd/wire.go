//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/services/chat/configs"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/data"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/service"
)

type App struct {
	Server *service.ChatService
}

func NewApp(server *service.ChatService) *App {
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

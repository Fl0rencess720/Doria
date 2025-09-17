//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/tts/configs"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/tts/internal/biz"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/tts/internal/data"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/tts/internal/service"
)

type App struct {
	Server *service.TTSService
}

func NewApp(server *service.TTSService) *App {
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

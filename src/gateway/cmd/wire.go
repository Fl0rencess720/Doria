//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"net/http"

	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/data"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service"
)

type App struct {
	HttpServer *http.Server
}

func NewApp(server *http.Server) *App {
	return &App{
		HttpServer: server,
	}
}

func wireApp() *App {
	panic(wire.Build(
		NewApp,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}

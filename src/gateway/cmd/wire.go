//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/data"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
)

type App struct {
	HttpServer      *service.HTTPServer
	SignalingServer *service.SignalingServer
}

func NewApp(httpServer *service.HTTPServer, signalingServer *service.SignalingServer) *App {
	return &App{
		HttpServer:      httpServer,
		SignalingServer: signalingServer,
	}
}

func wireApp() *App {
	panic(wire.Build(
		NewApp,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		circuitbreaker.ProviderSet,
	))
}

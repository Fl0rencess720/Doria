//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/Fl0rencess720/Doria/src/services/memory/configs"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/data"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/service"
)

type App struct {
	Server    *service.MemoryService
	MCPServer *service.RAGMCPService
}

func NewApp(server *service.MemoryService, mcpServer *service.RAGMCPService) *App {
	return &App{
		Server:    server,
		MCPServer: mcpServer,
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

//go:build wireinject

package main

import (
	"classroom-analysis/internal/ioc"

	"github.com/google/wire"
)

func InitWebServer() *App {
	wire.Build(
		ioc.InitViper,
		ioc.MinimalSet,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

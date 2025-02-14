//go:build wireinject
// +build wireinject

package cmd

import (
	"context"
	"web-service/base/app"
	"web-service/base/conf"
	"web-service/base/middleware"
	"web-service/base/server"
	"web-service/controller"
	"web-service/repo"
	"web-service/router"
	"web-service/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

const (
	Version = "1.0.0"
)

func newApplication(e *gin.Engine) *app.Application {
	return app.NewApp(
		app.WithName(conf.GetProjectName()),
		app.WithVersion(Version),
		app.WithServer(server.NewServer(e)),
	)
}
func InitApplication(ctx context.Context, cabinModelFile string) (*app.Application, func(), error) {
	panic(wire.Build(
		server.NewHttpServer,
		repo.ProviderRepo,
		service.ProviderService,
		controller.ProviderContr,
		middleware.ProviderMiddleware,
		router.ProviderRouter,
		newApplication,
	))
}

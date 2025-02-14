package middleware

import (
	"web-service/base"
	"web-service/pkg/permissions"
	"web-service/repo"

	"github.com/google/wire"
)

var ProviderMiddleware = wire.NewSet(
	wire.Bind(new(base.GetUserRepoer), new(*repo.GetUserRepo)),
	wire.Bind(new(permissions.Authorizer), new(*permissions.Authoriz)),
	repo.NewGetUserRepo,
	permissions.NewAuthoriz,
	NewAuthorization,
)

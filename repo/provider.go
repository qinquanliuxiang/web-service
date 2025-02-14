package repo

import (
	"web-service/base"
	"web-service/base/data"
	"web-service/pkg/permissions"

	"github.com/google/wire"
)

var ProviderRepo = wire.NewSet(
	wire.Bind(new(base.Cache), new(*data.Redis)),
	wire.Bind(new(base.GeneralUserRepoer), new(*UserRepo)),
	wire.Bind(new(base.GeneralRoleRepoer), new(*GeneralRoleRepo)),
	wire.Bind(new(base.GeneralPolicyRepoer), new(*PolicyRepo)),
	wire.Bind(new(base.AssociationPolicyer), new(*RoleAssociationRepo)),
	wire.Bind(new(permissions.GeneralAuthorizer), new(*permissions.GeneralAuthorizRepo)),
	wire.Bind(new(base.GetRoleRepoer), new(*GeneralRoleRepo)),
	wire.Bind(new(base.GetPolicyRepoer), new(*PolicyRepo)),
	data.CreateRDB,
	data.NewDB,
	data.NewRedis,
	NewUserRepo,
	NewRoleRepo,
	NewPolicyRepo,
	NewRoleAssociationRepo,
	permissions.NewGeneralAuthorizRepo,
	permissions.InitCasbin,
)

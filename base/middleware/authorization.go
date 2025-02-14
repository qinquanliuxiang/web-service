package middleware

import (
	"web-service/base"
	"web-service/base/apierrs"
	"web-service/base/constant"
	"web-service/base/data"
	"web-service/base/handler"
	"web-service/base/reason"
	"web-service/pkg/jwt"
	"web-service/pkg/permissions"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorizationMiddleware struct {
	cache      base.Cache
	authorizer permissions.Authorizer
	userRepo   base.GetUserRepoer
}

func NewAuthorization(cache base.Cache, authorizer permissions.Authorizer, userRepo base.GetUserRepoer) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		cache:      cache,
		authorizer: authorizer,
		userRepo:   userRepo,
	}
}

// Authorization 基于 Casbin 的鉴权中间件
func (a *AuthorizationMiddleware) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwt.GetMyCustomClaims(c)
		if err != nil {
			handler.HandleResponseAuthFailed(c, err)
			return
		}

		var roleName string
		roleName, err = a.cache.GetString(c, constant.RoleCacheKeyPrefix+claims.UserName)
		if err != nil {
			user, err := a.userRepo.GetUserByID(c, claims.UserID, base.WithUserRole())
			if err != nil {
				handler.HandleResponseAuthFailed(c, err)
				return
			}
			roleName = user.Role.Name
			_ = a.cache.SetString(c, constant.RoleCacheKeyPrefix+claims.UserName, roleName, &data.NeverExpires)
		}

		// 判断是否有权限D
		ok, err := a.authorizer.EnforceWithCtx(c, roleName, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			handler.HandleResponseAuthFailed(c, err)
			return
		}
		if !ok {
			zap.S().Errorf("用户: '%s', 角色: '%s', 没有权限执行操作: '%s', 访问资源: '%s'", claims.UserName, roleName, c.Request.Method, c.Request.URL.Path)
			handler.HandleResponseAuthFailed(c, apierrs.NewAuthError(reason.ErrPermission))
			return
		}

		c.Next()
	}
}

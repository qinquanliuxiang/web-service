package middleware

import (
	"errors"
	"web-service/base/apierrs"
	"web-service/base/constant"
	"web-service/base/handler"
	"web-service/pkg/jwt"

	"strings"

	"github.com/gin-gonic/gin"
)

var ErrHeaderEmpty = errors.New("auth in the request header is empty")
var ErrHeaderMalformed = errors.New("the auth format in the request header is incorrect")

// Authentication 基于JWT的认证中间件
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			handler.HandleResponseAuthFailed(c, apierrs.NewAuthError(ErrHeaderEmpty))
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			handler.HandleResponseAuthFailed(c, apierrs.NewAuthError(ErrHeaderMalformed))
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			handler.HandleResponseAuthFailed(c, apierrs.NewAuthError(err))
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set(constant.AuthMidwareKey, mc)
		c.Next()
	}
}

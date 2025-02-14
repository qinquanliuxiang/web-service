package middleware

import (
	"fmt"
	"time"
	"web-service/base/apierrs"
	"web-service/base/constant"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapMiddleware 日志中间件
func ZapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latency := endTime.Sub(startTime).Milliseconds()
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		path := c.Request.URL.Path

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("clientIP", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("latency", fmt.Sprintf("%dms", latency)),
		}

		_err, ok := c.Get(constant.LogErrMidwareKey)
		if ok {
			err := _err.(*apierrs.ApiError)
			caller := err.Stack
			errCode := err.Code

			fields = append(fields,
				zap.Int("errorCode", errCode),
				zap.String("msg", err.Msg),
				zap.String("error", err.Error()),
				zap.String("caller", caller),
			)
			zap.L().Error("request failed", fields...)
		} else {
			zap.L().Info("request success", fields...)
		}
	}
}

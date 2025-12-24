package ioc

import (
	"classroom-analysis/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

// 创建全局JWT中间件实例
var jwtMiddleware gin.HandlerFunc

func init() {
	// 初始化JWT中间件，忽略登录路径
	jwtMiddleware = middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/teacher/login").
		IgnorePaths("/teacher/register").
		Build()
}

func GetJWTMiddleware() gin.HandlerFunc {
	return jwtMiddleware
}
func InitMiddlewares(loggers []gin.HandlerFunc) []gin.HandlerFunc {
	// 这里可以添加各种中间件
	return loggers
}

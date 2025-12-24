package ioc

import (
	"classroom-analysis/internal/web"

	"github.com/gin-gonic/gin"
)

// InitGin 初始化Gin引擎
func InitGin(loggers []gin.HandlerFunc, emailHandler *web.EmailHandler, teacherHandler *web.TeacherHandler, fileHandler *web.FileHandler) *gin.Engine {
	engine := gin.Default()

	middlewares := InitMiddlewares(loggers)
	for _, middleware := range middlewares {
		engine.Use(middleware)
	}
	// 注册教师相关路由
	teacherHandler.RegisterRoutes(engine)
	//注册邮件相关路由
	emailHandler.RegisterRoutes(engine)
	// 注册文件上传相关路由
	fileHandler.RegisterRoutes(engine)

	return engine
}

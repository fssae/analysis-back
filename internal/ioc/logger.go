package ioc

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func InitLogger(client *mongo.Client) []gin.HandlerFunc {
	// 这里可以添加日志中间件
	return []gin.HandlerFunc{}
}

// NewLogger 创建日志器
func NewLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

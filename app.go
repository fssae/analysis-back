package main

import (
	"classroom-analysis/internal/ioc"
	"classroom-analysis/internal/web"
	"github.com/robfig/cron/v3"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	server  *gin.Engine
	mongodb *mongo.Client
	redis   *redis.Client
	minio   *minio.Client
	config  *ioc.Config

	FileHandler *web.FileHandler
}

func setupEnvironment() {
	//测试
	ioc.TimezoneInit()
	//初始化viper
	ioc.InitViper()
	//系统存活性监控
	ioc.InitPrometheus()
}

func (a *App) RegisterDebugRoute() {
	// a.server 是私有的，只有在同一个 main 包下才能访问
	a.server.GET("/debug/config", func(c *gin.Context) {
		c.JSON(200, a.config)
	})
}
func (app *App) Start() error {
	//挂载debug路由
	app.RegisterDebugRoute()
	ioc.StartKafkaResponseConsumer()
	ioc.InitApiColl(app.server, app.mongodb, app.redis)
	return app.server.Run(":8080")
}
func startCronJobs() {
	c := cron.New()
	//TODO
	//定时器处理信息，推送消息
	c.AddFunc("@hourly", func() {
		log.Println("开始执行聚合预处理...")
	})
	c.Start()
}

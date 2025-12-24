package main

import (
	"log"
)

func main() {
	//初始化
	setupEnvironment()
	//启动服务
	app := InitWebServer()
	if err := app.Start(); err != nil {
		log.Fatalf("应用启动失败: %v", err)
	}
}

package ioc

import (
	"classroom-analysis/internal/repository/dao"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitApiColl(engine *gin.Engine, db1 *mongo.Client, client redis.Cmdable) {
	db := db1.Database("classroom")
	//更新api集合的api
	routes := engine.Routes()
	apiColl := db.Collection("api")
	_, err := apiColl.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}
	var api dao.Api
	for _, route := range routes {
		api.Url = route.Path
		api.Method = route.Method
		_, err = apiColl.InsertOne(context.TODO(), api)
		if err != nil {
			panic("api数据库初始化失败")
		}
	}
}

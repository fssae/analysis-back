package ioc

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"log"
	"time"
)

func InitRedis() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       0,
		PoolSize: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis连接失败: %v", err)
	}

	return client
}

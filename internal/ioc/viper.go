package ioc

import (
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	viperOnce    sync.Once
	configOnce   sync.Once
	cachedConfig *Config
)

type Config struct {
	Mongodb MongodbConfig `mapstructure:"mongodb"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Minio   MinioConfig   `mapstructure:"minio"`
	Kafka   KafkaConfig   `mapstructure:"kafka"`
	General GeneralConfig `mapstructure:"general"`
}

type MongodbConfig struct {
	Account  string `mapstructure:"account"`
	Address1 string `mapstructure:"address1"`
	Port1    int    `mapstructure:"port1"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
}

type KafkaConfig struct {
	Address       string   `mapstructure:"address"`
	Brokers       []string `mapstructure:"brokers"`
	RequestTopic  string   `mapstructure:"RequestTopic"`
	ResponseTopic string   `mapstructure:"ResponseTopic"`
}

type GeneralConfig struct {
	Password string `mapstructure:"password"`
	JWT      string `mapstructure:"jwt"`
}

func InitViper() *Config {
	// 使用 sync.Once 确保 flag 定义和解析只执行一次
	viperOnce.Do(func() {
		file := pflag.String("config", "config/conf.yaml", "指定配置文件路径")
		pflag.Parse()
		viper.SetConfigFile(*file)
		viper.WatchConfig()

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	})

	// 使用 sync.Once 确保配置只解析一次
	configOnce.Do(func() {
		cfg := &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			panic(err)
		}
		cachedConfig = cfg
	})

	return cachedConfig
}

package ioc

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

type MinioConfig struct {
	Address   string `yaml:"address"`
	Port      string `yaml:"port"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
}

func NewMinioClient() *minio.Client {
	// 直接使用 viper 读取配置，避免结构体解析问题
	address := viper.GetString("minio.address")
	port := viper.GetString("minio.port")
	accessKey := viper.GetString("minio.access_key")
	secretKey := viper.GetString("minio.secret_key")
	bucketName := viper.GetString("minio.bucket")

	//// 调试信息
	//log.Printf("MinIO 配置读取结果:")
	//log.Printf("  Address: '%s'", address)
	//log.Printf("  Port: '%s'", port)
	//log.Printf("  AccessKey: '%s'", accessKey)
	//log.Printf("  SecretKey: '%s'", secretKey)
	//log.Printf("  Bucket: '%s'", bucketName)

	// 检查配置是否完整
	if address == "" || port == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		log.Println("MinIO 配置不完整，请检查 address、port、access_key、secret_key、bucket")
		return nil
	}

	// 初始化 Minio 客户端
	endpoint := address + ":" + port
	log.Printf("正在连接 MinIO: %s", endpoint)

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // 如果使用 HTTPS，Secure 设置为 true
	})

	if err != nil {
		log.Printf("创建 MinIO 客户端失败: %v", err)
		return nil
	}

	// 测试连接
	ctx := context.Background()
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		log.Printf("MinIO 连接测试失败: %v", err)
		return nil
	}

	// 确保bucket存在
	err = ensureBucketExists(ctx, minioClient, bucketName)
	if err != nil {
		log.Printf("确保bucket存在失败: %v", err)
		return nil
	}

	// 设置bucket策略为公开读取
	err = setBucketPolicy(ctx, minioClient, bucketName)
	err = setBucketPolicy(ctx, minioClient, "analysis")
	if err != nil {
		log.Printf("设置bucket策略失败: %v", err)
		// 不返回nil，因为这不是致命错误
	}

	log.Println("MinIO 客户端初始化成功")
	return minioClient
}

// ensureBucketExists 确保bucket存在，如果不存在则创建
func ensureBucketExists(ctx context.Context, client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		log.Printf("Bucket '%s' 不存在，正在创建...", bucketName)
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("Bucket '%s' 创建成功", bucketName)
	} else {
		log.Printf("Bucket '%s' 已存在", bucketName)
	}

	return nil
}

// setBucketPolicy 设置bucket策略为公开读取
func setBucketPolicy(ctx context.Context, client *minio.Client, bucketName string) error {
	// 设置允许公开读取的策略
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": ["*"]
				},
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::` + bucketName + `/*"
				]
			}
		]
	}`

	err := client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		return err
	}

	log.Printf("Bucket '%s' 策略设置为公开读取", bucketName)
	return nil
}

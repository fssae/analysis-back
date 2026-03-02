package service

import (
	"classroom-analysis/internal/repository"
	"fmt"
	"mime/multipart"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

type FileServiceInterface interface {
	Upload(ctx *gin.Context, file *multipart.FileHeader) (string, string, error)
	UploadFileMessage(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error)
}

type FileService struct {
	client         *minio.Client
	imageWhitelist []string
	videoWhitelist []string
	fileRepo       repository.FileInterface
}

func (f *FileService) UploadFileMessage(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error) {
	return f.fileRepo.UploadFileMessageRepository(ctx, url, fileType)
}

func (f *FileService) Upload(ctx *gin.Context, file *multipart.FileHeader) (string, string, error) {
	// 检查 MinIO 客户端是否初始化
	if f.client == nil {
		return "", "", errors.New("MinIO 客户端未初始化")
	}

	var contentType string
	var fileType string
	var err error

	// 判断文件类型
	if fileType, contentType, err = f.getFileType(file); err != nil {
		return "", "", err
	}

	open, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer open.Close() // 确保文件流关闭

	// 生成uuid
	u := uuid.NewV1()
	u = uuid.NewV5(u, file.Filename)
	bucketName := viper.GetString("minio.bucket")
	address := viper.GetString("minio.address")
	port := viper.GetString("minio.port")

	// 检查配置是否完整
	if bucketName == "" || address == "" || port == "" {
		return "", "", errors.New("MinIO 配置不完整")
	}

	// 上传对象：使用 UUID 作为文件名，直接放入桶中
	_, err = f.client.PutObject(ctx, bucketName, u.String(), open, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", "", fmt.Errorf("上传到 MinIO 失败: %w", err)
	}

	// 生成正确的MinIO访问URL
	url := fmt.Sprintf("http://%s:%s/%s/%s", address, port, bucketName, u.String())
	return url, fileType, nil
}

func (f *FileService) getFileType(file *multipart.FileHeader) (string, string, error) {
	name := file.Filename
	split := strings.Split(name, ".")
	typeString := split[len(split)-1]

	// 检查是否为图片
	if slices.Contains(f.imageWhitelist, typeString) {
		return "image", "image/" + typeString, nil
	}

	// 检查是否为视频
	if slices.Contains(f.videoWhitelist, typeString) {
		return "video", "video/" + typeString, nil
	}

	return "", "", errors.New("不支持的文件格式")
}

func NewFileService(
	client *minio.Client,
	fileRepo repository.FileInterface,
) FileServiceInterface {
	return &FileService{
		client:         client,
		imageWhitelist: []string{"jpg", "jpeg", "png", "gif", "bmp", "tif", "tiff", "webp", "svg"},
		videoWhitelist: []string{"mp4", "avi", "mov", "mkv", "wmv", "flv", "webm"},
		fileRepo:       fileRepo,
	}
}

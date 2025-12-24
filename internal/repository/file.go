package repository

import (
	"classroom-analysis/internal/repository/dao"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileInterface interface {
	UploadFileMessageRepository(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error)
}
type FileRepository struct {
	dao dao.FileDaoInterface
}

func (r *FileRepository) UploadFileMessageRepository(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error) {
	// 根据文件类型调用相应的DAO方法
	if fileType == "video" {
		return r.dao.UploadVideoMessage(ctx, url)
	}
	return r.dao.UploadImageMessage(ctx, url)
}

func NewFileRepository(dao dao.FileDaoInterface) *FileRepository {
	return &FileRepository{
		dao: dao,
	}
}

// ProvideFileInterface 提供 FileInterface 的 wire provider
func ProvideFileInterface(repo *FileRepository) FileInterface {
	return repo
}

package repository

import (
	"classroom-analysis/internal/repository/dao"
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileInterface interface {
	UploadFileMessageRepository(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error)
	GetFileMessageById(ctx context.Context, id primitive.ObjectID, fileType string) (interface{}, error)
}
type FileRepository struct {
	dao dao.FileDaoInterface
}

func (r *FileRepository) UploadFileMessageRepository(ctx *gin.Context, url string, fileType string) (primitive.ObjectID, error) {
	if fileType == "video" {
		return r.dao.UploadVideoMessage(ctx, url)
	}
	return r.dao.UploadImageMessage(ctx, url)
}

func (r *FileRepository) GetFileMessageById(ctx context.Context, id primitive.ObjectID, fileType string) (interface{}, error) {
	if fileType == "video" {
		return r.dao.GetVideoMessageById(ctx, id)
	}
	return r.dao.GetImageMessageById(ctx, id)
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

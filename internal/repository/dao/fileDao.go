package dao

import (
	"classroom-analysis/internal/domain"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileDaoInterface interface {
	UploadImageMessage(ctx *gin.Context, url string) (primitive.ObjectID, error)
	UploadVideoMessage(ctx *gin.Context, url string) (primitive.ObjectID, error)
}
type FileDao struct {
	image *mongo.Collection
	video *mongo.Collection
}

func (f *FileDao) UploadImageMessage(ctx *gin.Context, url string) (primitive.ObjectID, error) {
	// 上传图片消息
	// 查询是否已存在该视频
	var existing domain.ImageMessage
	err := f.video.FindOne(ctx.Request.Context(), bson.M{"url": url}).Decode(&existing)
	if err == nil {
		return existing.Id, nil
	}
	// 图片不存在，插入新文档
	newImage := domain.ImageMessage{
		Id:  primitive.NewObjectID(),
		Url: url,
	}
	insertResult, err := f.image.InsertOne(ctx.Request.Context(), newImage)
	if err != nil {
		log.Println("insert image fail: ", err)
		return primitive.NilObjectID, nil
	}

	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	return insertedID, nil

}
func (f *FileDao) UploadVideoMessage(ctx *gin.Context, url string) (primitive.ObjectID, error) {

	// 查询是否已存在该视频
	var existing domain.VideoMessage
	err := f.video.FindOne(ctx.Request.Context(), bson.M{"url": url}).Decode(&existing)
	if err == nil {
		return existing.Id, nil
	}

	// 视频不存在，插入新文档
	newVideo := domain.VideoMessage{
		Id:  primitive.NewObjectID(),
		Url: url,
	}
	insertResult, err := f.video.InsertOne(ctx.Request.Context(), newVideo)
	if err != nil {
		log.Println("insert video fail: ", err)
		return primitive.NilObjectID, nil
	}

	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

func NewFileDao(db *mongo.Database) *FileDao {
	return &FileDao{
		image: db.Collection("uploadImage"),
		video: db.Collection("uploadVideo"),
	}
}

// ProvideFileDaoInterface 提供 FileDaoInterface 的 wire provider
func ProvideFileDaoInterface(dao *FileDao) FileDaoInterface {
	return dao
}

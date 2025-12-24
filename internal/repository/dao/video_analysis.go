package dao

import (
	"classroom-analysis/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VideoAnalysisDAO struct {
	collection *mongo.Collection
}

func NewVideoAnalysisDAO(db *mongo.Database) *VideoAnalysisDAO {
	return &VideoAnalysisDAO{
		collection: db.Collection("video_analyses"),
	}
}

// Create 创建视频分析
func (dao *VideoAnalysisDAO) Create(ctx context.Context, videoAnalysis *domain.VideoAnalysis) error {
	videoAnalysis.CreatedAt = time.Now()
	result, err := dao.collection.InsertOne(ctx, videoAnalysis)
	if err != nil {
		return err
	}
	videoAnalysis.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByAnalysisId 根据分析ID查找视频分析
func (dao *VideoAnalysisDAO) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) (*domain.VideoAnalysis, error) {
	var videoAnalysis domain.VideoAnalysis
	err := dao.collection.FindOne(ctx, bson.M{"analysisId": analysisId}).Decode(&videoAnalysis)
	if err != nil {
		return nil, err
	}
	return &videoAnalysis, nil
}

// Update 更新视频分析
func (dao *VideoAnalysisDAO) Update(ctx context.Context, videoAnalysis *domain.VideoAnalysis) error {
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": videoAnalysis.Id},
		bson.M{"$set": videoAnalysis},
	)
	return err
}

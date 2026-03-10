package dao

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImageAnalysisDAO struct {
	collection *mongo.Collection
}

func NewImageAnalysisDAO(db *mongo.Database) *ImageAnalysisDAO {
	return &ImageAnalysisDAO{
		collection: db.Collection("image_analyses"),
	}
}

// Create 创建图片分析
func (dao *ImageAnalysisDAO) Create(ctx context.Context, imageAnalysis *domain.ImageAnalysis) error {
	imageAnalysis.CreatedAt = util.GetBeijingTime()
	result, err := dao.collection.InsertOne(ctx, imageAnalysis)
	if err != nil {
		return err
	}
	imageAnalysis.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByAnalysisId 根据分析ID查找图片分析
func (dao *ImageAnalysisDAO) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) (*domain.ImageAnalysis, error) {
	var imageAnalysis domain.ImageAnalysis
	err := dao.collection.FindOne(ctx, bson.M{"analysisId": analysisId}).Decode(&imageAnalysis)
	if err != nil {
		return nil, err
	}
	return &imageAnalysis, nil
}

// Update 更新图片分析
func (dao *ImageAnalysisDAO) Update(ctx context.Context, imageAnalysis *domain.ImageAnalysis) error {
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": imageAnalysis.Id},
		bson.M{"$set": imageAnalysis},
	)
	return err
}

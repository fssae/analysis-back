package dao

import (
	"classroom-analysis/internal/domain"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FaceAnalysisDAO struct {
	collection *mongo.Collection
}

func NewFaceAnalysisDAO(db *mongo.Database) *FaceAnalysisDAO {
	return &FaceAnalysisDAO{
		collection: db.Collection("face_analysis"),
	}
}

// Create 创建人脸分析记录

// FindByTaskId 根据任务ID查找所有人脸分析记录
func (dao *FaceAnalysisDAO) FindByTaskId(ctx context.Context, taskId string) ([]*domain.FaceAnalysis, error) {
	cursor, err := dao.collection.Find(ctx, bson.M{"taskId": taskId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var faceAnalyses []*domain.FaceAnalysis
	if err = cursor.All(ctx, &faceAnalyses); err != nil {
		return nil, err
	}
	return faceAnalyses, nil
}

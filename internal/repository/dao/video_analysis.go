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
	// 首先尝试从 video_analyses 集合查找
	var videoAnalysis domain.VideoAnalysis
	err := dao.collection.FindOne(ctx, bson.M{"analysisId": analysisId}).Decode(&videoAnalysis)
	if err == nil {
		// 如果在 video_analyses 集合中找到数据，直接返回
		return &videoAnalysis, nil
	}
	
	// 如果在 video_analyses 集合中未找到，尝试从 status 集合查找
	// 这是为了兼容旧的数据结构
	statusCollection := dao.collection.Database().Collection("status")
	var statusRecord bson.M
	err2 := statusCollection.FindOne(ctx, bson.M{"imageId": analysisId.Hex()}).Decode(&statusRecord)
	if err2 != nil {
		// 如果两个集合都没有找到数据，返回原始错误
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, err2
	}
	
	// 如果在 status 集合中找到数据，创建一个基本的 VideoAnalysis 对象
	videoAnalysis = domain.VideoAnalysis{
		AnalysisId: analysisId,
		Duration:   0, // 默认时长，可能需要从其他地方获取
		CreatedAt:  time.Now(), // 使用当前时间，或者从 statusRecord 中提取时间
		FocusTrend: []domain.FocusPoint{}, // 空的焦点趋势，如果有数据可以从 status 中解析
	}
	
	// 如果 statusRecord 中有时间信息，使用它
	if updatedAt, ok := statusRecord["updatedAt"].(primitive.DateTime); ok {
		videoAnalysis.CreatedAt = updatedAt.Time()
	}
	
	// 如果有 duration 信息，也可以设置
	// 这里可以根据实际需求进一步完善
	
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

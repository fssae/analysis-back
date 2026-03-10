package dao

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AnalysisTaskDAO struct {
	collection       *mongo.Collection
	statusCollection *mongo.Collection
}

func NewAnalysisTaskDAO(db *mongo.Database) *AnalysisTaskDAO {
	return &AnalysisTaskDAO{
		collection:       db.Collection("analysis_tasks"),
		statusCollection: db.Collection("status"),
	}
}

// Create 创建分析任务
func (dao *AnalysisTaskDAO) Create(ctx context.Context, task *domain.AnalysisTask) error {
	task.CreatedAt = util.GetBeijingTime()
	task.UpdatedAt = util.GetBeijingTime()
	result, err := dao.collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}
	task.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByTaskId 根据任务ID查找任务
func (dao *AnalysisTaskDAO) FindByTaskId(ctx context.Context, taskId string) ([]*domain.UpdateStatus, error) {
	filter := bson.M{}
	if taskId != "" {
		filter = bson.M{"taskId": taskId}
	}

	cur, err := dao.statusCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var status []*domain.UpdateStatus
	if err = cur.All(ctx, &status); err != nil {
		return nil, err
	}
	return status, nil
}

// Update 更新任务
func (dao *AnalysisTaskDAO) Update(ctx context.Context, task *domain.AnalysisTask) error {
	task.UpdatedAt = util.GetBeijingTime()
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": task.Id},
		bson.M{"$set": task},
	)
	return err
}

// UpdateStatus 更新任务状态
func (dao *AnalysisTaskDAO) UpdateStatus(ctx context.Context, taskId string, status string, progress int) error {
	_, err := dao.statusCollection.UpdateOne(
		ctx,
		bson.M{"taskId": taskId},
		bson.M{"$set": bson.M{
			"status":    status,
			"progress":  progress,
			"updatedAt": util.GetBeijingTime(),
		}},
	)
	return err
}

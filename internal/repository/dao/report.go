package dao

import (
	"classroom-analysis/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportDAO struct {
	collection *mongo.Collection
}

func NewReportDAO(db *mongo.Database) *ReportDAO {
	return &ReportDAO{
		collection: db.Collection("reports"),
	}
}

// Create 创建报告
func (dao *ReportDAO) Create(ctx context.Context, report *domain.Report) error {
	report.CreatedAt = time.Now()
	result, err := dao.collection.InsertOne(ctx, report)
	if err != nil {
		return err
	}
	report.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByAnalysisId 根据分析ID查找报告
func (dao *ReportDAO) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) (*domain.Report, error) {
	var report domain.Report
	err := dao.collection.FindOne(ctx, bson.M{"analysisId": analysisId}).Decode(&report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// FindByTeacherId 根据教师ID查找报告列表
func (dao *ReportDAO) FindByTeacherId(ctx context.Context, teacherId primitive.ObjectID, limit int64) ([]*domain.Report, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit)
	cursor, err := dao.collection.Find(ctx, bson.M{"teacherId": teacherId}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []*domain.Report
	if err = cursor.All(ctx, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// Update 更新报告
func (dao *ReportDAO) Update(ctx context.Context, report *domain.Report) error {
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": report.Id},
		bson.M{"$set": report},
	)
	return err
}

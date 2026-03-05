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

type TeacherDAO struct {
	collection *mongo.Collection
}

func NewTeacherDAO(db *mongo.Database) *TeacherDAO {
	return &TeacherDAO{
		collection: db.Collection("teachers"),
	}
}

// Create 创建教师
func (dao *TeacherDAO) Create(ctx context.Context, teacher *domain.Teacher) error {
	teacher.CreatedAt = time.Now()
	teacher.UpdatedAt = time.Now()
	result, err := dao.collection.InsertOne(ctx, teacher)
	if err != nil {
		return err
	}
	teacher.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByTeacherId 根据工号查找教师
func (dao *TeacherDAO) FindByTeacherAccount(ctx context.Context, teacherId string) (*domain.Teacher, error) {
	var teacher domain.Teacher
	err := dao.collection.FindOne(ctx, bson.M{"teacherId": teacherId}).Decode(&teacher)
	if err != nil {
		return nil, err
	}
	return &teacher, nil
}

// FindById 根据ID查找教师
func (dao *TeacherDAO) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Teacher, error) {
	var teacher domain.Teacher
	err := dao.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&teacher)
	if err != nil {
		return nil, err
	}
	return &teacher, nil
}

// Update 更新教师信息
func (dao *TeacherDAO) Update(ctx context.Context, teacher *domain.Teacher) error {
	teacher.UpdatedAt = time.Now()
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": teacher.Id},
		bson.M{"$set": teacher},
	)
	return err
}

// GetSettings 获取教师设置
func (dao *TeacherDAO) GetSettings(ctx context.Context, teacherId string) (*domain.SettingsData, error) {
	settingsColl := dao.collection.Database().Collection("teacher_settings")
	var settings domain.SettingsData
	err := settingsColl.FindOne(ctx, bson.M{"teacherId": teacherId}).Decode(&settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveSettings 保存教师设置
func (dao *TeacherDAO) SaveSettings(ctx context.Context, teacherId string, settings interface{}) error {
	settingsColl := dao.collection.Database().Collection("teacher_settings")

	_, err := settingsColl.UpdateOne(
		ctx,
		bson.M{"teacherId": teacherId},
		bson.M{
			"$set": bson.M{
				"teacherId": teacherId,
				"settings":  settings,
				"updatedAt": time.Now(),
			},
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

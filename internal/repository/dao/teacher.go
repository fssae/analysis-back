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

	// 定义一个与数据库结构匹配的结构体（嵌套在 settings 字段下）
	var doc struct {
		TeacherId   string    `bson:"teacherId"`
		Settings    struct {
			UISettings           domain.UISettings           `bson:"uisettings"`
			AnalysisSettings     domain.AnalysisSettings     `bson:"analysissettings"`
			NotificationSettings domain.NotificationSettings `bson:"notificationsettings"`
		} `bson:"settings"`
		CreatedAt time.Time `bson:"createdAt"`
		UpdatedAt time.Time `bson:"updatedAt"`
	}

	err := settingsColl.FindOne(ctx, bson.M{"teacherId": teacherId}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	// 转换为 domain.SettingsData
	settings := &domain.SettingsData{
		TeacherId:            doc.TeacherId,
		UISettings:           doc.Settings.UISettings,
		AnalysisSettings:     doc.Settings.AnalysisSettings,
		NotificationSettings: doc.Settings.NotificationSettings,
		CreatedAt:            doc.CreatedAt,
		UpdatedAt:            doc.UpdatedAt,
	}

	return settings, nil
}

// SaveSettings 保存教师设置
func (dao *TeacherDAO) SaveSettings(ctx context.Context, teacherId string, settings interface{}) error {
	settingsColl := dao.collection.Database().Collection("teacher_settings")

	// 将 settings 转换为 bson.M 以便使用小写字段名
	var settingsDoc bson.M
	settingsBytes, err := bson.Marshal(settings)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(settingsBytes, &settingsDoc)
	if err != nil {
		return err
	}

	// 构建嵌套的 settings 文档（与数据库现有结构一致）
	updateDoc := bson.M{
		"teacherId": teacherId,
		"settings": bson.M{
			"uisettings":           settingsDoc["uiSettings"],
			"analysissettings":     settingsDoc["analysisSettings"],
			"notificationsettings": settingsDoc["notificationSettings"],
		},
		"updatedAt": time.Now(),
	}

	_, err = settingsColl.UpdateOne(
		ctx,
		bson.M{"teacherId": teacherId},
		bson.M{
			"$set": updateDoc,
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

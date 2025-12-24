package dao

import (
	"classroom-analysis/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StudentDAO struct {
	collection *mongo.Collection
}

func NewStudentDAO(db *mongo.Database) *StudentDAO {
	return &StudentDAO{
		collection: db.Collection("students"),
	}
}

// Create 创建学生
func (dao *StudentDAO) Create(ctx context.Context, student *domain.Student) error {
	student.CreatedAt = time.Now()
	result, err := dao.collection.InsertOne(ctx, student)
	if err != nil {
		return err
	}
	student.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindById 根据ID查找学生
func (dao *StudentDAO) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Student, error) {
	var student domain.Student
	err := dao.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&student)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// FindAll 查找所有学生
// FindAll函数用于查找所有学生信息
func (dao *StudentDAO) FindAll(ctx context.Context) ([]*domain.Student, error) {
	cursor, err := dao.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var students []*domain.Student
	if err = cursor.All(ctx, &students); err != nil {
		return nil, err
	}
	return students, nil
}

// FindByName 根据姓名查找学生
// 根据名字查找学生
func (dao *StudentDAO) FindByName(ctx context.Context, name string) ([]*domain.Student, error) {
	// 在集合中查找名字为name的学生
	cursor, err := dao.collection.Find(ctx, bson.M{"name": name})
	if err != nil {
		return nil, err
	}
	// 关闭游标
	defer cursor.Close(ctx)

	var students []*domain.Student
	// 将游标中的数据转换为Student结构体
	if err = cursor.All(ctx, &students); err != nil {
		return nil, err
	}
	return students, nil

}

// Count 统计学生数量
func (dao *StudentDAO) Count(ctx context.Context) (int64, error) {
	return dao.collection.CountDocuments(ctx, bson.M{})
}

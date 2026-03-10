package dao

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmailDaoInterface interface {
	DeleteUnread(context.Context, primitive.ObjectID) error
	EmailUpdate(primitive.ObjectID, string, string, float64) error
	DeleteByTaskId(primitive.ObjectID, string) error
	GetUnread(context.Context, primitive.ObjectID) (int, error)
	GetAllByTeacherId(context.Context, primitive.ObjectID) (domain.Email, error)
}
type EmailDao struct {
	collection *mongo.Collection
}

func NewEmailDao(db *mongo.Client) EmailDaoInterface {
	return &EmailDao{
		collection: db.Database("classroom").Collection("email"),
	}
}
func (dao *EmailDao) GetAllByTeacherId(ctx context.Context, id primitive.ObjectID) (domain.Email, error) {
	// 构建过滤条件
	filter := bson.M{"teacherId": id}

	// 查找教师邮箱文档
	var email domain.Email
	err := dao.collection.FindOne(ctx, filter).Decode(&email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Email{}, err
		}
		return domain.Email{}, err
	}

	return email, nil
}

//
//func (dao *EmailDao) GetAllByTeacherId(ctx context.Context, id primitive.ObjectID, request domain.GetUrlsRequest) (domain.Email, error) {
//	// 构建过滤条件
//	filter := bson.M{"teacherId": id}
//
//	// 查找教师邮箱文档
//	var email domain.Email
//	err := dao.collection.FindOne(ctx, filter).Decode(&email)
//	if err != nil {
//		if errors.Is(err, mongo.ErrNoDocuments) {
//			return domain.Email{}, err
//		}
//		return domain.Email{}, err
//	}
//
//	// 对 urls 数组进行分页处理
//	page := request.Page
//	pageSize := request.PageSize
//
//	// 设置默认值
//	if page <= 0 {
//		page = 1
//	}
//	if pageSize <= 0 {
//		pageSize = 5
//	}
//
//	start := (page - 1) * pageSize
//	end := start + pageSize
//
//	// 边界检查
//	if start >= len(email.Urls) {
//		email.Urls = []domain.EmailUrl{}
//	} else {
//		if end > len(email.Urls) {
//			end = len(email.Urls)
//		}
//		email.Urls = email.Urls[start:end]
//	}
//
//	return email, nil
//}

func (dao *EmailDao) GetUnread(ctx context.Context, teacherId primitive.ObjectID) (int, error) {
	filter := bson.M{"teacherId": teacherId}
	var email domain.Email
	err := dao.collection.FindOne(ctx, filter).Decode(&email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, err
		}
	}
	return email.Unread, nil
}
func (dao *EmailDao) DeleteByTaskId(teacherId primitive.ObjectID, taskId string) error {
	filter := bson.M{"teacherId": teacherId}
	update := bson.M{
		"$pull": bson.M{
			"urls": bson.M{"taskId": taskId},
		},
	}
	result, err := dao.collection.UpdateOne(context.Background(), filter, update)
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}
func (dao *EmailDao) DeleteUnread(ctx context.Context, teacherId primitive.ObjectID) error {
	filter := bson.M{"teacherId": teacherId}
	update := bson.M{"$set": bson.M{"unread": 0}}
	result, err := dao.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
func (dao *EmailDao) EmailUpdate(teacherId primitive.ObjectID, imageId string, url string, confidence float64) error {
	newEmailUrl := domain.EmailUrl{
		Confidence: confidence,
		Url:        url,
		TaskId:     imageId,
		Time:       util.GetBeijingTime(),
	}
	filter := bson.M{
		"teacherId": teacherId,
	}
	update := bson.M{
		"$inc":  bson.M{"unread": 1},
		"$push": bson.M{"urls": newEmailUrl},
		"$setOnInsert": bson.M{
			"teacherId": teacherId,
			"_id":       primitive.NewObjectID(),
		},
		"$set": bson.M{
			"lastUpdated": util.GetBeijingTime(),
			"date":        util.GetBeijingTime(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := dao.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

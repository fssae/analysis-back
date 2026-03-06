package dao

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseDAO 泛型基础 DAO，封装通用 CRUD
type BaseDAO[T any] struct {
	Coll *mongo.Collection
}

// NewBaseDAO 创建一个新的基础 DAO
func NewBaseDAO[T any](db *mongo.Database, collectionName string) BaseDAO[T] {
	return BaseDAO[T]{
		Coll: db.Collection(collectionName),
	}
}

// InsertOne 插入单条数据
func (d *BaseDAO[T]) InsertOne(ctx context.Context, entity *T) (*mongo.InsertOneResult, error) {
	return d.Coll.InsertOne(ctx, entity)
}

// FindOne 通用单条查询
func (d *BaseDAO[T]) FindOne(ctx context.Context, filter bson.M, opts ...*options.FindOneOptions) (*T, error) {
	var entity T
	err := d.Coll.FindOne(ctx, filter, opts...).Decode(&entity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没找到不报错，返回 nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindList 通用列表查询 (支持排序、分页)
func (d *BaseDAO[T]) FindList(ctx context.Context, filter bson.M, skip, limit int64, sort bson.D) ([]*T, int64, error) {
	// 1. 构建选项
	findOpts := options.Find()
	if sort != nil {
		findOpts.SetSort(sort)
	}
	if limit > 0 {
		findOpts.SetLimit(limit)
	}
	if skip > 0 {
		findOpts.SetSkip(skip)
	}

	// 2. 执行查询
	cursor, err := d.Coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 3. 解码结果
	var results []*T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// 4. 统计总数 (如果需要分页展示总数)
	// 注意：如果只是查列表不需要总数，外部可以忽略这个返回值，或者这里优化为不查 count
	total, err := d.Coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Count 统计数量
func (d *BaseDAO[T]) Count(ctx context.Context, filter bson.M) (int64, error) {
	return d.Coll.CountDocuments(ctx, filter)
}

// UpdateOne 通用更新
func (d *BaseDAO[T]) UpdateOne(ctx context.Context, filter bson.M, update bson.M, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return d.Coll.UpdateOne(ctx, filter, update, opts...)
}

// DeleteOne 通用删除
func (d *BaseDAO[T]) DeleteOne(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {
	return d.Coll.DeleteOne(ctx, filter)
}

package ioc

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongodb() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:zjh770910@82.156.64.69:27017"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化索引
	InitIndexes(client)

	return client
}

func InitMongoDatabase(client *mongo.Client) *mongo.Database {
	return client.Database("classroom")
}

// InitIndexes 初始化数据库索引，提升查询性能
func InitIndexes(client *mongo.Client) {
	db := client.Database("classroom")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 分析集合索引
	analysisColl := db.Collection("analysis")
	analysisIndexes := []struct {
		keys     bson.D
		name     string
		unique   bool
		expireAt bool
	}{
		{bson.D{{Key: "timestamp", Value: -1}}, "idx_timestamp_desc", false, false},
		{bson.D{{Key: "className", Value: 1}, {Key: "courseName", Value: 1}}, "idx_class_course", false, false},
		{bson.D{{Key: "filetype", Value: 1}}, "idx_filetype", false, false},
		{bson.D{{Key: "imageid", Value: 1}}, "idx_imageid", false, false},
		{bson.D{{Key: "createdAt", Value: -1}}, "idx_createdat_desc", false, false},
	}

	for _, idx := range analysisIndexes {
		createIndex(ctx, analysisColl, idx.keys, idx.name, idx.unique)
	}

	// 学生专注度集合索引
	studentFocusColl := db.Collection("student_focus")
	studentFocusIndexes := []struct {
		keys     bson.D
		name     string
		unique   bool
		expireAt bool
	}{
		{bson.D{{Key: "studentId", Value: 1}, {Key: "analysisId", Value: 1}}, "idx_student_analysis", false, false},
		{bson.D{{Key: "date", Value: -1}}, "idx_date_desc", false, false},
	}

	for _, idx := range studentFocusIndexes {
		createIndex(ctx, studentFocusColl, idx.keys, idx.name, idx.unique)
	}

	// 图片分析集合索引
	imageAnalysisColl := db.Collection("image_analyses")
	imageAnalysisIndexes := []struct {
		keys     bson.D
		name     string
		unique   bool
		expireAt bool
	}{
		{bson.D{{Key: "analysisId", Value: 1}}, "idx_analysisid", false, false},
		{bson.D{{Key: "createdAt", Value: -1}}, "idx_createdat_desc", false, false},
	}

	for _, idx := range imageAnalysisIndexes {
		createIndex(ctx, imageAnalysisColl, idx.keys, idx.name, idx.unique)
	}

	// 视频分析集合索引
	videoAnalysisColl := db.Collection("video_analyses")
	videoAnalysisIndexes := []struct {
		keys     bson.D
		name     string
		unique   bool
		expireAt bool
	}{
		{bson.D{{Key: "analysisId", Value: 1}}, "idx_analysisid", false, false},
		{bson.D{{Key: "createdAt", Value: -1}}, "idx_createdat_desc", false, false},
	}

	for _, idx := range videoAnalysisIndexes {
		createIndex(ctx, videoAnalysisColl, idx.keys, idx.name, idx.unique)
	}

	log.Println("✅ 数据库索引初始化完成")
}

func createIndex(ctx context.Context, coll *mongo.Collection, keys bson.D, name string, unique bool) {
	// 检查索引是否已存在
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		log.Printf("⚠️  列出索引失败: %v", err)
		return
	}

	defer cursor.Close(ctx)

	exists := false
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err == nil {
			if doc["name"] == name {
				exists = true
				break
			}
		}
	}

	if exists {
		log.Printf("ℹ️  索引 %s 已存在", name)
		return
	}

	// 创建索引
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(name).SetUnique(unique),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel, opts)
	if err != nil {
		log.Printf("⚠️  创建索引 %s 失败: %v", name, err)
	} else {
		log.Printf("✅ 索引 %s 创建成功", name)
	}
}

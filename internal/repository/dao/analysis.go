package dao

import (
	"classroom-analysis/internal/domain"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalysisDAO struct {
	collection       *mongo.Collection
	statusCollection *mongo.Collection
}

func NewAnalysisDAO(db *mongo.Database) *AnalysisDAO {
	return &AnalysisDAO{
		collection:       db.Collection("analysis"),
		statusCollection: db.Collection("status"),
	}
}

// FindRecentAnalysis 查找最近的一条分析记录
func (dao *AnalysisDAO) FindRecentAnalysis(ctx context.Context) ([]*domain.Analysis, error) {
	var analysis []*domain.Analysis
	cur, err := dao.collection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(5),
	)
	if err != nil {
		return nil, err
	}
	err = cur.All(ctx, &analysis)
	if err != nil {
		return nil, err
	}
	return analysis, nil
}
func (dao *AnalysisDAO) FindByTeacherId(ctx context.Context, teacherId primitive.ObjectID, limit int64) ([]*domain.Analysis, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit)
	cursor, err := dao.collection.Find(ctx, bson.M{"teacherId": teacherId}, opts)
	if err != nil {
		return nil, err
	}
	err = cursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	var analyses []*domain.Analysis
	if err = cursor.All(ctx, &analyses); err != nil {
		return nil, err
	}
	return analyses, nil
}

// Create 创建分析记录
func (dao *AnalysisDAO) Create(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = time.Now()
	result, err := dao.collection.InsertOne(ctx, analysis)
	if err != nil {
		return err
	}
	analysis.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindById 根据ID查找分析记录
func (dao *AnalysisDAO) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Analysis, error) {
	var analysis domain.Analysis
	err := dao.collection.FindOne(ctx, bson.M{"imageid": id}).Decode(&analysis)
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// Update 更新分析记录
func (dao *AnalysisDAO) Update(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = time.Now()
	_, err := dao.collection.UpdateOne(
		ctx,
		bson.M{"_id": analysis.Id},
		bson.M{"$set": analysis},
	)
	return err
}

// CountByTeacherId 统计教师的分析数量
func (dao *AnalysisDAO) CountByTeacherId(ctx context.Context) (int64, int64, int64, int64, error) {
	//总图片数
	totalImages, err := dao.collection.CountDocuments(ctx, bson.M{"filetype": "image"})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	//总视频数
	totalVideos, err := dao.collection.CountDocuments(ctx, bson.M{"filetype": "video"})
	cursor, err := dao.collection.Find(ctx, bson.M{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	//总学生数,总分析报告数
	totalFaces := int64(0)
	totalDoc := int64(0)
	for cursor.Next(ctx) {
		var analysis domain.Analysis
		err := cursor.Decode(&analysis)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		totalFaces += int64(len(analysis.Faces))
		totalDoc++
	}
	return totalImages, totalVideos, totalFaces, totalDoc, err
}

// GetLastAnalysisTime 获取最后分析时间
func (dao *AnalysisDAO) GetLastAnalysisTime(ctx context.Context, teacherId primitive.ObjectID) (*time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	var analysis domain.Analysis
	err := dao.collection.FindOne(ctx, bson.M{"teacherId": teacherId}, opts).Decode(&analysis)
	if err != nil {
		return nil, err
	}
	return &analysis.Timestamp, nil
}

func (dao *AnalysisDAO) UpdateConfig(update *domain.UpdateConfigRequest) error {
	_, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"imageid": update.AnalysisId},
		bson.M{"$set": bson.M{
			"analysisMode": update.AnalysisMode,
			"className":    update.ClassName,
			"courseName":   update.CourseName,
			"description":  update.Description,
			"timestamp":    update.Timestamp,
		}},
	)
	return err
}
func (dao *AnalysisDAO) GetRankDao(ctx context.Context, req *domain.RankRequest) (*domain.RankItem, error) {
	filter := bson.M{}
	if req.CourseName != "" {
		filter["courseName"] = req.CourseName
	}
	if req.ClassName != "" {
		filter["className"] = req.ClassName
	}

	// 更改 _id 为 className + courseName 的组合以避免重复项
	groupID := bson.M{
		"className":  "$className",
		"courseName": "$courseName",
	}

	sortField := "avgFocusScore"
	sortOrder := -1

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":           groupID,
			"className":     bson.M{"$first": "$className"},
			"courseName":    bson.M{"$first": "$courseName"},
			"avgFocusScore": bson.M{"$avg": "$faces.focus_score"},
			"timestamp":     bson.M{"$first": "$timestamp"},
			"resultUrl":     bson.M{"$first": "$result_url"},
		}}},
		// 添加数值精度控制阶段
		{{"$addFields", bson.M{
			"avgFocusScore": bson.M{
				"$round": []interface{}{"$avgFocusScore", 2},
			},
		}}},
		{{"$facet", bson.M{
			"data": []bson.M{
				{"$sort": bson.D{{Key: sortField, Value: sortOrder}}},
				{"$skip": int64((req.Page - 1) * req.PageSize)},
				{"$limit": int64(req.PageSize)},
			},
			"total": []bson.M{
				{"$count": "count"},
			},
		}}},
	}

	cursor, err := dao.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var facetResult []struct {
		Data []struct {
			ID         interface{} `bson:"_id"`
			ClassName  string      `bson:"className"`
			CourseName string      `bson:"courseName"`
			FocusAvg   float64     `bson:"avgFocusScore"`
		} `bson:"data"`
		Total []struct {
			Count int64 `bson:"count"`
		} `bson:"total"`
	}
	if err := cursor.All(ctx, &facetResult); err != nil {
		return nil, err
	}

	var total int64
	var list []*domain.Rank
	if len(facetResult) > 0 {
		if len(facetResult[0].Total) > 0 {
			total = facetResult[0].Total[0].Count
		}
		for _, d := range facetResult[0].Data {
			list = append(list, &domain.Rank{
				ClassName:  d.ClassName,
				CourseName: d.CourseName,
				FocusAvg:   d.FocusAvg,
			})
		}
	}
	return &domain.RankItem{
		List:  list,
		Total: total,
	}, nil
}

// FindClassAnalysisList 条件分页查询班级分析
// FindClassAnalysisList 条件分页查询班级分析
func (dao *AnalysisDAO) FindClassAnalysisList(ctx context.Context, courseName, className, startDate, endDate string, page, pageSize int) ([]*domain.Analysis, int64, error) {
	filter := bson.M{}
	if courseName != "" {
		filter["courseName"] = courseName
	}
	if className != "" {
		filter["className"] = className
	}

	// 修改时间过滤逻辑以支持精确到分钟
	if startDate != "" || endDate != "" {
		timeFilter := bson.M{}
		if startDate != "" {
			// 支持两种格式：日期格式和日期时间格式
			var t time.Time
			var err error

			// 首先尝试解析完整的时间格式(包含小时分钟)
			t, err = time.Parse("2006-01-02 15:04", startDate)
			if err != nil {
				// 如果失败，则尝试解析仅日期格式
				t, err = time.Parse("2006-01-02", startDate)
			}
			if err == nil {
				timeFilter["$gte"] = t
			}
		}
		if endDate != "" {
			// 支持两种格式：日期格式和日期时间格式
			var t time.Time
			var err error

			// 首先尝试解析完整的时间格式(包含小时分钟)
			t, err = time.Parse("2006-01-02 15:04", endDate)
			if err != nil {
				// 如果失败，则尝试解析仅日期格式，并加一天
				t, err = time.Parse("2006-01-02", endDate)
				if err == nil {
					t = t.Add(24 * time.Hour)
				}
			}
			if err == nil {
				timeFilter["$lte"] = t
			}
		}
		if len(timeFilter) > 0 {
			filter["timestamp"] = timeFilter
		}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := dao.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var analyses []*domain.Analysis
	if err = cursor.All(ctx, &analyses); err != nil {
		return nil, 0, err
	}
	total, err := dao.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return analyses, total, nil
}

// InsertAvg 计算指定 taskId 的 faces 平均专注度并更新到数据库
func (dao *AnalysisDAO) InsertAvg(taskId string) error {
	// 定义聚合管道
	pipeline := mongo.Pipeline{
		// 匹配条件：taskId 和存在 faces 字段
		{{"$match", bson.M{
			"Id":    taskId,
			"faces": bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		// 展开 faces 数组
		{{"$unwind", "$faces"}},
		// 按 imageId 分组，计算每个 image 的平均专注度
		{{"$group", bson.M{
			"_id":           "$taskId",
			"className":     bson.M{"$first": "$className"},
			"courseName":    bson.M{"$first": "$courseName"},
			"avgFocusScore": bson.M{"$avg": "$faces.focus_score"},
			"timestamp":     bson.M{"$first": "$timestamp"},
			"resultUrl":     bson.M{"$first": "$result_url"},
		}}},
		// 更新原集合中的文档
		{{"$merge", bson.M{
			"into":           "analysis",
			"whenMatched":    "merge",   // 合并字段而不是替换
			"whenNotMatched": "discard", // 不匹配时不插入新文档
		}}},
	}

	// 执行聚合操作
	cursor, err := dao.collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	// 等待操作完成
	return cursor.Err()
}

// UpdateStatus 更新任务状态
func (dao *AnalysisDAO) UpdateStatus(update *domain.UpdateStatus) error {
	//task_status
	filter := bson.M{"imageId": update.TaskId, "teacherId": update.TeacherId}
	updateDoc := bson.M{
		"$set": bson.M{
			"status":              update.Status,
			"resultUrl":           update.ResultUrl,
			"errorMsg":            update.ErrorMsg,
			"confidenceThreshold": update.ConfidenceThreshold,
			"updatedAt":           time.Now(),
		},
	}
	// 使用upsert确保如果任务不存在则创建
	opts := options.Update().SetUpsert(true)
	_, err := dao.statusCollection.UpdateOne(context.Background(), filter, updateDoc, opts)
	return err
}

// GetTaskStatus 获取任务状态
func (dao *AnalysisDAO) GetTaskStatus(ctx context.Context, taskId string) (*domain.UpdateStatus, error) {
	filter := bson.M{"taskId": taskId}
	var status domain.UpdateStatus
	err := dao.collection.FindOne(ctx, filter).Decode(&status)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 任务不存在
		}
		return nil, err
	}
	return &status, nil
}

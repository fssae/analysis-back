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

// AnalysisDAO 继承泛型 BaseDAO
type AnalysisDAO struct {
	BaseDAO[domain.Analysis] // 嵌入 BaseDAO，直接获得基础能力
	statusCollection         *mongo.Collection
}

func NewAnalysisDAO(db *mongo.Database) *AnalysisDAO {
	return &AnalysisDAO{
		BaseDAO:          NewBaseDAO[domain.Analysis](db, "analysis"),
		statusCollection: db.Collection("status"),
	}
}

// FindRecentAnalysis 查找最近的一条分析记录
func (dao *AnalysisDAO) FindRecentAnalysis(ctx context.Context) ([]*domain.Analysis, error) {
	// 直接复用 FindList
	list, _, err := dao.FindList(ctx, bson.M{}, 0, 5, bson.D{{Key: "timestamp", Value: -1}})
	return list, err
}

// FindByTeacherId 根据 TeacherID 查找
func (dao *AnalysisDAO) FindByTeacherId(ctx context.Context, teacherId primitive.ObjectID, limit int64) ([]*domain.Analysis, error) {
	list, _, err := dao.FindList(ctx, bson.M{"teacherId": teacherId}, 0, limit, bson.D{{Key: "createdAt", Value: -1}})
	return list, err
}

// Create 创建分析记录
func (dao *AnalysisDAO) Create(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = time.Now()
	// 复用 InsertOne
	res, err := dao.InsertOne(ctx, analysis)
	if err != nil {
		return err
	}
	// 回填 ID
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		analysis.Id = oid
	}
	return nil
}

// FindById 根据ID查找 (根据 imageid 字段，保持原有业务逻辑)
func (dao *AnalysisDAO) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Analysis, error) {
	return dao.BaseDAO.FindOne(ctx, bson.M{"imageid": id})
}

// FindByImageId 根据 ImageID 查找
func (dao *AnalysisDAO) FindByImageId(ctx context.Context, imageId primitive.ObjectID) (*domain.Analysis, error) {
	// 根据 imageid 字段查找，保持原业务逻辑
	return dao.FindById(ctx, imageId) // 复用 FindById 方法
}

// Update 更新分析记录
func (dao *AnalysisDAO) Update(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = time.Now()
	_, err := dao.UpdateOne(ctx, bson.M{"_id": analysis.Id}, bson.M{"$set": analysis})
	return err
}

// CountByTeacherId 统计逻辑 (保持业务逻辑，但使用 BaseDAO 简化)
func (dao *AnalysisDAO) CountByTeacherId(ctx context.Context) (int64, int64, int64, int64, error) {
	// 1. 统计图片
	totalImages, err := dao.Count(ctx, bson.M{"filetype": "image"})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	// 2. 统计视频
	totalVideos, err := dao.Count(ctx, bson.M{"filetype": "video"})
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// 3. 统计人脸总数 (这里的逻辑比较重，建议未来优化为聚合查询)
	// 为了不破坏原有逻辑，这里我们还是查出来遍历，但使用 FindList 简化写法
	// 注意：这里 limit 传 0 表示查所有，慎用，数据量大建议用 Aggregate $sum
	allData, _, err := dao.FindList(ctx, bson.M{}, 0, 0, nil)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	totalFaces := int64(0)
	totalDoc := int64(len(allData))
	for _, analysis := range allData {
		totalFaces += int64(len(analysis.Faces))
	}

	return totalImages, totalVideos, totalFaces, totalDoc, nil
}

// GetLastAnalysisTime 获取最后分析时间
func (dao *AnalysisDAO) GetLastAnalysisTime(ctx context.Context, teacherId primitive.ObjectID) (*time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	analysis, err := dao.FindOne(ctx, bson.M{"teacherId": teacherId}, opts)
	if err != nil {
		return nil, err
	}
	if analysis == nil {
		return nil, nil
	}
	return &analysis.Timestamp, nil
}

// UpdateConfig 更新配置
func (dao *AnalysisDAO) UpdateConfig(update *domain.UpdateConfigRequest) error {
	_, err := dao.UpdateOne(
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

// UpdateAnalysisName 更新分析文件名
func (dao *AnalysisDAO) UpdateAnalysisName(update *domain.UpdateAnalysisNameRequest) error {
	imageId, err := primitive.ObjectIDFromHex(update.ImageId)
	if err != nil {
		return err
	}

	result, err := dao.UpdateOne(
		context.Background(),
		bson.M{"imageid": imageId},
		bson.M{"$set": bson.M{
			"fileName": update.FileName,
		}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("未找到对应的分析记录")
	}

	return nil
}

// FindClassAnalysisList 复杂查询
func (dao *AnalysisDAO) FindClassAnalysisList(ctx context.Context, courseName, className, startDate, endDate string, page, pageSize int) ([]*domain.Analysis, int64, error) {
	filter := bson.M{}
	if courseName != "" {
		filter["courseName"] = courseName
	}
	if className != "" {
		filter["className"] = className
	}

	// 抽取时间处理逻辑，保持代码整洁
	if timeFilter := buildTimeFilter(startDate, endDate); len(timeFilter) > 0 {
		filter["timestamp"] = timeFilter
	}

	// 一行代码搞定查询+分页+统计
	return dao.FindList(ctx, filter, int64((page-1)*pageSize), int64(pageSize), bson.D{{Key: "timestamp", Value: -1}})
}

// buildTimeFilter 辅助函数：构建时间查询条件
func buildTimeFilter(startDate, endDate string) bson.M {
	timeFilter := bson.M{}
	parseTime := func(dateStr string, endOfDay bool) (time.Time, error) {
		layouts := []string{"2006-01-02 15:04", "2006-01-02"}
		for _, layout := range layouts {
			t, err := time.Parse(layout, dateStr)
			if err == nil {
				if endOfDay && layout == "2006-01-02" {
					return t.Add(24 * time.Hour), nil
				}
				return t, nil
			}
		}
		return time.Time{}, errors.New("invalid date")
	}

	if startDate != "" {
		if t, err := parseTime(startDate, false); err == nil {
			timeFilter["$gte"] = t
		}
	}
	if endDate != "" {
		if t, err := parseTime(endDate, true); err == nil {
			timeFilter["$lte"] = t
		}
	}
	return timeFilter
}

// =================================================================================
// 复杂的聚合操作 (Aggregate) 无法泛型化，保留原生写法，但清理了 context
// =================================================================================

func (dao *AnalysisDAO) GetRankDao(ctx context.Context, req *domain.RankRequest) (*domain.RankItem, error) {
	filter := bson.M{}
	if req.CourseName != "" {
		filter["courseName"] = req.CourseName
	}
	if req.ClassName != "" {
		filter["className"] = req.ClassName
	}

	groupID := bson.M{
		"className":  "$className",
		"courseName": "$courseName",
	}

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
		{{"$addFields", bson.M{
			"avgFocusScore": bson.M{"$round": []interface{}{"$avgFocusScore", 2}},
		}}},
		{{"$facet", bson.M{
			"data": []bson.M{
				{"$sort": bson.D{{Key: "avgFocusScore", Value: -1}}},
				{"$skip": int64((req.Page - 1) * req.PageSize)},
				{"$limit": int64(req.PageSize)},
			},
			"total": []bson.M{
				{"$count": "count"},
			},
		}}},
	}

	// 使用 dao.Coll 直接访问原生 collection
	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 定义临时结构体接收结果
	var facetResult []struct {
		Data []struct {
			ClassName  string  `bson:"className"`
			CourseName string  `bson:"courseName"`
			FocusAvg   float64 `bson:"avgFocusScore"`
		} `bson:"data"`
		Total []struct {
			Count int64 `bson:"count"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &facetResult); err != nil {
		return nil, err
	}

	result := &domain.RankItem{List: []*domain.Rank{}, Total: 0}
	if len(facetResult) > 0 {
		if len(facetResult[0].Total) > 0 {
			result.Total = facetResult[0].Total[0].Count
		}
		for _, d := range facetResult[0].Data {
			result.List = append(result.List, &domain.Rank{
				ClassName:  d.ClassName,
				CourseName: d.CourseName,
				FocusAvg:   d.FocusAvg,
			})
		}
	}
	return result, nil
}

func (dao *AnalysisDAO) InsertAvg(taskId string) error {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"Id":    taskId, // 注意：这里匹配的是 Id 字段，请确保数据库字段名一致
			"faces": bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":           "$taskId",
			"className":     bson.M{"$first": "$className"},
			"courseName":    bson.M{"$first": "$courseName"},
			"avgFocusScore": bson.M{"$avg": "$faces.focus_score"},
			"timestamp":     bson.M{"$first": "$timestamp"},
			"resultUrl":     bson.M{"$first": "$result_url"},
		}}},
		{{"$merge", bson.M{
			"into":           "analysis",
			"whenMatched":    "merge",
			"whenNotMatched": "discard",
		}}},
	}

	cursor, err := dao.Coll.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	// 聚合操作不需要 decode，close 即可
	defer cursor.Close(context.Background())
	return nil
}

// =================================================================================
// Status 相关操作 (涉及另一个 Collection，无法使用 BaseDAO，手动实现)
// =================================================================================

func (dao *AnalysisDAO) UpdateStatus(update *domain.UpdateStatus) error {
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
	opts := options.Update().SetUpsert(true)
	_, err := dao.statusCollection.UpdateOne(context.Background(), filter, updateDoc, opts)
	return err
}

func (dao *AnalysisDAO) GetTaskStatus(ctx context.Context, taskId string) (*domain.UpdateStatus, error) {
	var status domain.UpdateStatus
	err := dao.statusCollection.FindOne(ctx, bson.M{"taskId": taskId}).Decode(&status)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (dao *AnalysisDAO) FindByTaskId(ctx context.Context, taskId string) (*domain.Analysis, error) {
	var analysis domain.Analysis
	err := dao.Coll.FindOne(ctx, bson.M{"taskId": taskId}).Decode(&analysis)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &analysis, nil
}

func (dao *AnalysisDAO) FindByImageIdString(ctx context.Context, imageIdStr string) (*domain.Analysis, error) {
	imageId, err := primitive.ObjectIDFromHex(imageIdStr)
	if err != nil {
		return nil, err
	}
	return dao.FindByImageId(ctx, imageId)
}

func (dao *AnalysisDAO) UpdateFileNameByTaskId(ctx context.Context, taskId, fileName string) error {
	result, err := dao.UpdateOne(
		ctx,
		bson.M{"taskId": taskId},
		bson.M{"$set": bson.M{
			"fileName": fileName,
		}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("analysis not found for taskId: " + taskId)
	}

	return nil
}

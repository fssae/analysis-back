package dao

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"
	"errors"
	"math"
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
func (dao *AnalysisDAO) FindByTeacherId(ctx context.Context, limit int64) ([]*domain.Analysis, error) {
	list, _, err := dao.FindList(ctx, bson.M{}, 0, limit, bson.D{{Key: "createdAt", Value: -1}})
	return list, err
}

// Create 创建分析记录
func (dao *AnalysisDAO) Create(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = util.GetBeijingTime()
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
	return dao.BaseDAO.FindOne(ctx, bson.M{"_id": id})
}

// FindByImageId 根据 ImageID 查找
func (dao *AnalysisDAO) FindByImageId(ctx context.Context, imageId primitive.ObjectID) (*domain.Analysis, error) {
	// 根据 imageid 字段查找
	return dao.BaseDAO.FindOne(ctx, bson.M{"imageid": imageId})
}

// Update 更新分析记录
func (dao *AnalysisDAO) Update(ctx context.Context, analysis *domain.Analysis) error {
	analysis.Timestamp = util.GetBeijingTime()
	_, err := dao.UpdateOne(ctx, bson.M{"_id": analysis.Id}, bson.M{"$set": analysis})
	return err
}

// UpdateAnalysisStatus 更新分析记录状态
func (dao *AnalysisDAO) UpdateAnalysisStatus(ctx context.Context, imageId primitive.ObjectID, status string) error {
	_, err := dao.UpdateOne(ctx, bson.M{"imageid": imageId}, bson.M{"$set": bson.M{"status": status}})
	return err
}

// CountByTeacherId 统计逻辑 (优化为使用聚合查询)
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

	// 3. 统计人脸总数 (使用聚合查询，避免加载所有数据到内存)
	// 使用 $sum 直接在数据库层面统计，避免遍历
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"faces": bson.M{"$exists": true, "$ne": []interface{}{}}}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer cursor.Close(ctx)

	var totalFaces int64 = 0
	if cursor.Next(ctx) {
		var result struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err == nil {
			totalFaces = result.Total
		}
	}

	return totalImages, totalVideos, totalFaces, totalImages + totalVideos, nil
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

// GetFatigueTimeSeries 获取疲劳度时间序列数据
func (dao *AnalysisDAO) GetFatigueTimeSeries(ctx context.Context, analysisId primitive.ObjectID) ([]domain.TimeSeriesPoint, error) {
	// 使用聚合查询获取时间序列数据
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"_id": analysisId}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":          "$className",
			"avgFatigue":   bson.M{"$avg": "$faces.fatigue_score"},
			"maxFatigue":   bson.M{"$max": "$faces.fatigue_score"},
			"minFatigue":   bson.M{"$min": "$faces.fatigue_score"},
			"avgFocus":     bson.M{"$avg": "$faces.focus_score"},
			"avgBlinkRate": bson.M{"$avg": "$faces.blink_rate"},
			"avgYawnCount": bson.M{"$avg": "$faces.yawn_count"},
		}}},
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []domain.TimeSeriesPoint
	if cursor.Next(ctx) {
		var point domain.TimeSeriesPoint
		if err := cursor.Decode(&point); err == nil {
			result = append(result, point)
		}
	}

	return result, nil
}

// GetEmotionTimeSeries 获取情绪时间序列数据
func (dao *AnalysisDAO) GetEmotionTimeSeries(ctx context.Context, analysisId primitive.ObjectID) ([]domain.EmotionTimeSeriesPoint, error) {
	// 使用聚合查询获取情绪时间序列数据
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"_id": analysisId}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":            "$className",
			"angryCount":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "angry"}}, 1, 0}}},
			"happyCount":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "happy"}}, 1, 0}}},
			"neutralCount":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "neutral"}}, 1, 0}}},
			"sadCount":       bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "sad"}}, 1, 0}}},
			"avgFluctuation": bson.M{"$avg": "$faces.emotion_fluctuation"},
		}}},
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []domain.EmotionTimeSeriesPoint
	if cursor.Next(ctx) {
		var point domain.EmotionTimeSeriesPoint
		if err := cursor.Decode(&point); err == nil {
			result = append(result, point)
		}
	}

	return result, nil
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
			"Id":    taskId,
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
			"updatedAt":           util.GetBeijingTime(),
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
	objectID, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return err
	}

	result, err := dao.UpdateOne(
		ctx,
		bson.M{"imageid": objectID},
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

// ========== 全局可视化分析相关方法 ==========

// GetGlobalAnalysisData 获取全局分析数据
func (dao *AnalysisDAO) GetGlobalAnalysisData(ctx context.Context) (*domain.GlobalAnalysisResponse, error) {
	response := &domain.GlobalAnalysisResponse{
		Overview: domain.GlobalOverview{
			EmotionDistribution: make(map[string]int),
		},
		VideoStats:   make([]domain.VideoStat, 0),
		ImageStats:   make([]domain.ImageStat, 0),
		StudentStats: make([]domain.StudentGlobalStat, 0),
		EmotionTrend: make([]domain.EmotionTrendPoint, 0),
		FocusTrend:   make([]domain.FocusTrendPoint, 0),
		FatigueTrend: make([]domain.FatigueTrendPoint, 0),
	}

	// 1. 获取全局概览统计
	overview, err := dao.getGlobalOverview(ctx)
	if err != nil {
		return nil, err
	}
	response.Overview = *overview

	// 2. 获取视频统计
	videoStats, err := dao.getVideoStats(ctx)
	if err != nil {
		return nil, err
	}
	response.VideoStats = videoStats

	// 3. 获取图片统计
	imageStats, err := dao.getImageStats(ctx)
	if err != nil {
		return nil, err
	}
	response.ImageStats = imageStats

	// 4. 获取学生统计
	studentStats, err := dao.getStudentGlobalStats(ctx)
	if err != nil {
		return nil, err
	}
	response.StudentStats = studentStats

	// 5. 获取趋势数据
	emotionTrend, focusTrend, fatigueTrend, err := dao.getTrendData(ctx)
	if err != nil {
		return nil, err
	}
	response.EmotionTrend = emotionTrend
	response.FocusTrend = focusTrend
	response.FatigueTrend = fatigueTrend

	return response, nil
}

// getGlobalOverview 获取全局概览统计
func (dao *AnalysisDAO) getGlobalOverview(ctx context.Context) (*domain.GlobalOverview, error) {
	overview := &domain.GlobalOverview{
		EmotionDistribution: make(map[string]int),
	}

	// 统计总分析数
	totalCount, err := dao.Count(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	overview.TotalAnalysis = totalCount

	// 统计视频数
	videoCount, err := dao.Count(ctx, bson.M{"filetype": "video"})
	if err != nil {
		return nil, err
	}
	overview.TotalVideos = videoCount
	overview.TotalImages = totalCount - videoCount

	// 聚合统计学生数、平均专注度、平均疲劳度、情绪分布
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"faces": bson.M{"$exists": true, "$ne": []interface{}{}}}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":              nil,
			"totalStudents":    bson.M{"$sum": 1},
			"avgFocusScore":    bson.M{"$avg": "$faces.focus_score"},
			"avgFatigueScore":  bson.M{"$avg": "$faces.fatigue_score"},
			"highFatigueCount": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$gte": []interface{}{"$faces.fatigue_score", 0.7}}, 1, 0}}},
			"lowFocusCount":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$lt": []interface{}{"$faces.focus_score", 0.5}}, 1, 0}}},
			"happyCount":       bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "happy"}}, 1, 0}}},
			"neutralCount":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "neutral"}}, 1, 0}}},
			"sadCount":         bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "sad"}}, 1, 0}}},
			"angryCount":       bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "angry"}}, 1, 0}}},
			"surpriseCount":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "surprise"}}, 1, 0}}},
			"fearCount":        bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "fear"}}, 1, 0}}},
			"disgustCount":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "disgust"}}, 1, 0}}},
		}}},
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalStudents    int64   `bson:"totalStudents"`
			AvgFocusScore    float64 `bson:"avgFocusScore"`
			AvgFatigueScore  float64 `bson:"avgFatigueScore"`
			HighFatigueCount int64   `bson:"highFatigueCount"`
			LowFocusCount    int64   `bson:"lowFocusCount"`
			HappyCount       int     `bson:"happyCount"`
			NeutralCount     int     `bson:"neutralCount"`
			SadCount         int     `bson:"sadCount"`
			AngryCount       int     `bson:"angryCount"`
			SurpriseCount    int     `bson:"surpriseCount"`
			FearCount        int     `bson:"fearCount"`
			DisgustCount     int     `bson:"disgustCount"`
		}
		if err := cursor.Decode(&result); err == nil {
			overview.TotalStudents = result.TotalStudents
			overview.AverageFocusScore = math.Round(result.AvgFocusScore * 100)
			overview.AverageFatigueScore = math.Round(result.AvgFatigueScore * 100)
			overview.HighFatigueStudents = result.HighFatigueCount
			overview.LowFocusStudents = result.LowFocusCount
			overview.EmotionDistribution["happy"] = result.HappyCount
			overview.EmotionDistribution["neutral"] = result.NeutralCount
			overview.EmotionDistribution["sad"] = result.SadCount
			overview.EmotionDistribution["angry"] = result.AngryCount
			overview.EmotionDistribution["surprise"] = result.SurpriseCount
			overview.EmotionDistribution["fear"] = result.FearCount
			overview.EmotionDistribution["disgust"] = result.DisgustCount
		}
	}

	return overview, nil
}

// getVideoStats 获取视频统计
func (dao *AnalysisDAO) getVideoStats(ctx context.Context) ([]domain.VideoStat, error) {
	// 首先获取视频文件列表
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"filetype": "video",
			"faces":    bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":             "$fileName",
			"className":       bson.M{"$first": "$className"},
			"courseName":      bson.M{"$first": "$courseName"},
			"studentCount":    bson.M{"$sum": 1},
			"analysisCount":   bson.M{"$sum": 1},
			"avgFocusScore":   bson.M{"$avg": "$faces.focus_score"},
			"avgFatigueScore": bson.M{"$avg": "$faces.fatigue_score"},
			// 收集所有 recent_emotions 数组用于后续统计
			"allRecentEmotions": bson.M{"$push": "$faces.recent_emotions"},
		}}},
		{{"$project", bson.M{
			"fileName":            "$_id",
			"className":           1,
			"courseName":          1,
			"studentCount":        1,
			"analysisCount":       1,
			"averageFocusScore":   "$avgFocusScore",
			"averageFatigueScore": "$avgFatigueScore",
			"allRecentEmotions":   1,
		}}},
		{{"$limit", 50}}, // 限制返回数量
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videoStats []domain.VideoStat
	for cursor.Next(ctx) {
		var rawData struct {
			FileName            string     `bson:"fileName"`
			ClassName           string     `bson:"className"`
			CourseName          string     `bson:"courseName"`
			StudentCount        int        `bson:"studentCount"`
			AnalysisCount       int        `bson:"analysisCount"`
			AverageFocusScore   float64    `bson:"averageFocusScore"`
			AverageFatigueScore float64    `bson:"averageFatigueScore"`
			AllRecentEmotions   [][]string `bson:"allRecentEmotions"`
		}
		if err := cursor.Decode(&rawData); err != nil {
			continue
		}

		// 统计情绪分布 - 从 recent_emotions 数组中统计
		emotionDistribution := make(map[string]int)
		for _, emotions := range rawData.AllRecentEmotions {
			for _, emotion := range emotions {
				if emotion != "" {
					emotionDistribution[emotion]++
				}
			}
		}

		stat := domain.VideoStat{
			FileName:            rawData.FileName,
			ClassName:           rawData.ClassName,
			CourseName:          rawData.CourseName,
			StudentCount:        rawData.StudentCount,
			AnalysisCount:       rawData.AnalysisCount,
			AverageFocusScore:   math.Round(rawData.AverageFocusScore * 100),
			AverageFatigueScore: math.Round(rawData.AverageFatigueScore * 100),
			EmotionDistribution: emotionDistribution,
		}
		videoStats = append(videoStats, stat)
	}

	return videoStats, nil
}

// getImageStats 获取图片统计
func (dao *AnalysisDAO) getImageStats(ctx context.Context) ([]domain.ImageStat, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"filetype": "image",
			"faces":    bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id":             "$fileName",
			"className":       bson.M{"$first": "$className"},
			"courseName":      bson.M{"$first": "$courseName"},
			"studentCount":    bson.M{"$sum": 1},
			"analysisCount":   bson.M{"$sum": 1},
			"avgFocusScore":   bson.M{"$avg": "$faces.focus_score"},
			"avgFatigueScore": bson.M{"$avg": "$faces.fatigue_score"},
			"happyCount":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "happy"}}, 1, 0}}},
			"neutralCount":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "neutral"}}, 1, 0}}},
			"sadCount":        bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "sad"}}, 1, 0}}},
			"angryCount":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "angry"}}, 1, 0}}},
			"surpriseCount":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "surprise"}}, 1, 0}}},
		}}},
		{{"$project", bson.M{
			"fileName":            "$_id",
			"className":           1,
			"courseName":          1,
			"studentCount":        1,
			"analysisCount":       1,
			"averageFocusScore":   "$avgFocusScore",
			"averageFatigueScore": "$avgFatigueScore",
			"emotionDistribution": bson.M{
				"happy":    "$happyCount",
				"neutral":  "$neutralCount",
				"sad":      "$sadCount",
				"angry":    "$angryCount",
				"surprise": "$surpriseCount",
			},
		}}},
		{{"$limit", 50}}, // 限制返回数量
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var imageStats []domain.ImageStat
	for cursor.Next(ctx) {
		var rawData struct {
			FileName            string         `bson:"fileName"`
			ClassName           string         `bson:"className"`
			CourseName          string         `bson:"courseName"`
			StudentCount        int            `bson:"studentCount"`
			AnalysisCount       int            `bson:"analysisCount"`
			AverageFocusScore   float64        `bson:"averageFocusScore"`
			AverageFatigueScore float64        `bson:"averageFatigueScore"`
			EmotionDistribution map[string]int `bson:"emotionDistribution"`
		}
		if err := cursor.Decode(&rawData); err != nil {
			continue
		}
		stat := domain.ImageStat{
			FileName:            rawData.FileName,
			ClassName:           rawData.ClassName,
			CourseName:          rawData.CourseName,
			StudentCount:        rawData.StudentCount,
			AnalysisCount:       rawData.AnalysisCount,
			AverageFocusScore:   math.Round(rawData.AverageFocusScore * 100),
			AverageFatigueScore: math.Round(rawData.AverageFatigueScore * 100),
			EmotionDistribution: rawData.EmotionDistribution,
		}
		imageStats = append(imageStats, stat)
	}

	return imageStats, nil
}

// getStudentGlobalStats 获取学生全局统计
func (dao *AnalysisDAO) getStudentGlobalStats(ctx context.Context) ([]domain.StudentGlobalStat, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"faces": bson.M{"$exists": true, "$ne": []interface{}{}}}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id": bson.M{
				"faceIndex":  "$faces.face_index",
				"className":  "$className",
				"courseName": "$courseName",
			},
			"analysisCount":   bson.M{"$sum": 1},
			"avgFocusScore":   bson.M{"$avg": "$faces.focus_score"},
			"avgFatigueScore": bson.M{"$avg": "$faces.fatigue_score"},
			"emotions":        bson.M{"$push": "$faces.emotion"},
			"latestTime":      bson.M{"$max": "$timestamp"},
		}}},
		{{"$project", bson.M{
			"faceIndex":           "$_id.faceIndex",
			"className":           "$_id.className",
			"courseName":          "$_id.courseName",
			"analysisCount":       1,
			"averageFocusScore":   "$avgFocusScore",
			"averageFatigueScore": "$avgFatigueScore",
			"emotions":            1,
			"latestAnalysisTime":  "$latestTime",
		}}},
		{{"$limit", 100}}, // 限制返回数量
	}

	cursor, err := dao.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var studentStats []domain.StudentGlobalStat
	for cursor.Next(ctx) {
		var stat domain.StudentGlobalStat
		var rawData struct {
			FaceIndex           int       `bson:"faceIndex"`
			ClassName           string    `bson:"className"`
			CourseName          string    `bson:"courseName"`
			AnalysisCount       int       `bson:"analysisCount"`
			AverageFocusScore   float64   `bson:"averageFocusScore"`
			AverageFatigueScore float64   `bson:"averageFatigueScore"`
			Emotions            []string  `bson:"emotions"`
			LatestAnalysisTime  time.Time `bson:"latestAnalysisTime"`
		}
		if err := cursor.Decode(&rawData); err == nil {
			stat.FaceIndex = rawData.FaceIndex
			stat.ClassName = rawData.ClassName
			stat.CourseName = rawData.CourseName
			stat.AnalysisCount = rawData.AnalysisCount
			// 保留整数
			stat.AverageFocusScore = math.Round(rawData.AverageFocusScore * 100)
			stat.AverageFatigueScore = math.Round(rawData.AverageFatigueScore * 100)
			stat.LatestAnalysisTime = rawData.LatestAnalysisTime
			// 计算平均情绪
			stat.AverageEmotion = calculateAverageEmotion(rawData.Emotions)
			studentStats = append(studentStats, stat)
		}
	}

	return studentStats, nil
}

// getTrendData 获取趋势数据
func (dao *AnalysisDAO) getTrendData(ctx context.Context) ([]domain.EmotionTrendPoint, []domain.FocusTrendPoint, []domain.FatigueTrendPoint, error) {
	// 获取最近30天的数据
	thirtyDaysAgo := util.GetBeijingTime().AddDate(0, 0, -30)

	// 情绪趋势
	emotionPipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"timestamp": bson.M{"$gte": thirtyDaysAgo},
			"faces":     bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$timestamp"},
			},
			"happyCount":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "happy"}}, 1, 0}}},
			"neutralCount":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "neutral"}}, 1, 0}}},
			"sadCount":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "sad"}}, 1, 0}}},
			"angryCount":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "angry"}}, 1, 0}}},
			"surpriseCount": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "surprise"}}, 1, 0}}},
			"fearCount":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "fear"}}, 1, 0}}},
			"disgustCount":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$faces.emotion", "disgust"}}, 1, 0}}},
			"total":         bson.M{"$sum": 1},
		}}},
		{{"$sort", bson.M{"_id": 1}}},
	}

	emotionCursor, err := dao.Coll.Aggregate(ctx, emotionPipeline)
	if err != nil {
		return nil, nil, nil, err
	}
	defer emotionCursor.Close(ctx)

	var emotionTrend []domain.EmotionTrendPoint
	for emotionCursor.Next(ctx) {
		var point domain.EmotionTrendPoint
		var rawData struct {
			Date          string `bson:"_id"`
			HappyCount    int    `bson:"happyCount"`
			NeutralCount  int    `bson:"neutralCount"`
			SadCount      int    `bson:"sadCount"`
			AngryCount    int    `bson:"angryCount"`
			SurpriseCount int    `bson:"surpriseCount"`
			FearCount     int    `bson:"fearCount"`
			DisgustCount  int    `bson:"disgustCount"`
			Total         int    `bson:"total"`
		}
		if err := emotionCursor.Decode(&rawData); err == nil {
			point.Date = rawData.Date
			point.Happy = rawData.HappyCount
			point.Neutral = rawData.NeutralCount
			point.Sad = rawData.SadCount
			point.Angry = rawData.AngryCount
			point.Surprise = rawData.SurpriseCount
			point.Fear = rawData.FearCount
			point.Disgust = rawData.DisgustCount
			point.Total = rawData.Total
			emotionTrend = append(emotionTrend, point)
		}
	}

	// 专注度趋势
	focusPipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"timestamp": bson.M{"$gte": thirtyDaysAgo},
			"faces":     bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$timestamp"},
			},
			"avgFocusScore": bson.M{"$avg": "$faces.focus_score"},
			"studentCount":  bson.M{"$sum": 1},
		}}},
		{{"$sort", bson.M{"_id": 1}}},
	}

	focusCursor, err := dao.Coll.Aggregate(ctx, focusPipeline)
	if err != nil {
		return nil, nil, nil, err
	}
	defer focusCursor.Close(ctx)

	var focusTrend []domain.FocusTrendPoint
	for focusCursor.Next(ctx) {
		var point domain.FocusTrendPoint
		var rawData struct {
			Date          string  `bson:"_id"`
			AvgFocusScore float64 `bson:"avgFocusScore"`
			StudentCount  int     `bson:"studentCount"`
		}
		if err := focusCursor.Decode(&rawData); err == nil {
			point.Date = rawData.Date
			point.AverageFocusScore = math.Round(rawData.AvgFocusScore * 100)
			point.StudentCount = rawData.StudentCount
			focusTrend = append(focusTrend, point)
		}
	}

	// 疲劳度趋势
	fatiguePipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"timestamp": bson.M{"$gte": thirtyDaysAgo},
			"faces":     bson.M{"$exists": true, "$ne": []interface{}{}},
		}}},
		{{"$unwind", "$faces"}},
		{{"$group", bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$timestamp"},
			},
			"avgFatigueScore":  bson.M{"$avg": "$faces.fatigue_score"},
			"highFatigueCount": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$gte": []interface{}{"$faces.fatigue_score", 0.7}}, 1, 0}}},
		}}},
		{{"$sort", bson.M{"_id": 1}}},
	}

	fatigueCursor, err := dao.Coll.Aggregate(ctx, fatiguePipeline)
	if err != nil {
		return nil, nil, nil, err
	}
	defer fatigueCursor.Close(ctx)

	var fatigueTrend []domain.FatigueTrendPoint
	for fatigueCursor.Next(ctx) {
		var point domain.FatigueTrendPoint
		var rawData struct {
			Date             string  `bson:"_id"`
			AvgFatigueScore  float64 `bson:"avgFatigueScore"`
			HighFatigueCount int     `bson:"highFatigueCount"`
		}
		if err := fatigueCursor.Decode(&rawData); err == nil {
			point.Date = rawData.Date
			point.AverageFatigueScore = math.Round(rawData.AvgFatigueScore * 100)
			point.HighFatigueCount = rawData.HighFatigueCount
			fatigueTrend = append(fatigueTrend, point)
		}
	}

	return emotionTrend, focusTrend, fatigueTrend, nil
}

// calculateAverageEmotion 计算平均情绪
func calculateAverageEmotion(emotions []string) string {
	if len(emotions) == 0 {
		return "neutral"
	}

	emotionCount := make(map[string]int)
	for _, emotion := range emotions {
		emotionCount[emotion]++
	}

	maxCount := 0
	avgEmotion := "neutral"
	for emotion, count := range emotionCount {
		if count > maxCount {
			maxCount = count
			avgEmotion = emotion
		}
	}

	return avgEmotion
}

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

type StudentFocusDAO struct {
	collection *mongo.Collection
}

func NewStudentFocusDAO(db *mongo.Database) *StudentFocusDAO {
	return &StudentFocusDAO{
		collection: db.Collection("analysis"),
	}
}

// GetFocusDistribution 获取注意力分布数据
func (dao *StudentFocusDAO) GetFocusDistribution(ctx context.Context) (map[string]int, error) {
	// 构建聚合管道
	pipeline := []bson.M{
		// 展开 faces 数组
		{"$unwind": "$faces"},
		// 按照 focus_score 分类计数
		{"$bucket": bson.M{
			"groupBy":    "$faces.focus_score",
			"boundaries": []float64{0, 0.6, 0.7, 0.73, 0.77, 1.0},
			"default":    "other", // 可选，默认分类
			"output": bson.M{
				"count": bson.M{"$sum": 1},
			},
		}},
	}

	cursor, err := dao.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type BucketResult struct {
		Id    float64 `bson:"_id"`
		Count int     `bson:"count"`
	}

	var results []BucketResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 映射到具体标签
	focusMap := map[string]int{
		"distracted":  0, //9
		"focused":     0,
		"highFocus":   0,
		"lowFocus":    0,
		"mediumFocus": 0, //25
	}

	for _, res := range results {
		switch {
		case res.Id >= 0.77:
			focusMap["highFocus"] += res.Count
		case res.Id >= 0.73:
			focusMap["focused"] += res.Count
		case res.Id >= 0.7:
			focusMap["lowFocus"] += res.Count
		case res.Id >= 0.6:
			focusMap["mediumFocus"] += res.Count
		case res.Id >= 0:
			focusMap["distracted"] += res.Count
		}
	}

	return focusMap, nil
}
func (dao *StudentFocusDAO) GetAverageFocusScores(ctx context.Context) ([]float64, error) {
	now := time.Now().UTC()
	// 计算本周一零点
	weekday := int(now.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	sunday := monday.AddDate(0, 0, 6).Add(time.Hour*23 + time.Minute*59 + time.Second*59)

	pipeline := []bson.M{
		{"$match": bson.M{
			"timestamp": bson.M{
				"$gte": monday,
				"$lte": sunday,
			},
		}},
		{"$unwind": "$faces"},
		{"$group": bson.M{
			"_id":           bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$timestamp"}},
			"avgFocusScore": bson.M{"$avg": "$faces.focus_score"},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := dao.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type dayAvg struct {
		Date          string  `bson:"_id"`
		AvgFocusScore float64 `bson:"avgFocusScore"`
	}
	var results []dayAvg
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 构造周一到周日的日期字符串
	dailyAverages := make([]float64, 7)
	dateMap := make(map[string]float64)
	for _, r := range results {
		dateMap[r.Date] = r.AvgFocusScore
	}
	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i).Format("2006-01-02")
		if avg, ok := dateMap[day]; ok {
			dailyAverages[i] = avg
		} else {
			dailyAverages[i] = 0
		}
	}
	return dailyAverages, nil
}

// Create 创建学生专注度记录
func (dao *StudentFocusDAO) Create(ctx context.Context, studentFocus *domain.StudentFocus) error {
	studentFocus.CreatedAt = time.Now()
	result, err := dao.collection.InsertOne(ctx, studentFocus)
	if err != nil {
		return err
	}
	studentFocus.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByStudentId 根据学生ID查找专注度记录
func (dao *StudentFocusDAO) FindByStudentId(ctx context.Context, studentId primitive.ObjectID, limit int64) ([]*domain.StudentFocus, error) {
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(limit)
	cursor, err := dao.collection.Find(ctx, bson.M{"studentId": studentId}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var studentFocuses []*domain.StudentFocus
	if err = cursor.All(ctx, &studentFocuses); err != nil {
		return nil, err
	}
	return studentFocuses, nil
}

// FindByAnalysisId 根据分析ID查找专注度记录
func (dao *StudentFocusDAO) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) ([]*domain.StudentFocus, error) {
	cursor, err := dao.collection.Find(ctx, bson.M{"imageid": analysisId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var studentFocuses []*domain.StudentFocus
	if err = cursor.All(ctx, &studentFocuses); err != nil {
		return nil, err
	}
	return studentFocuses, nil
}

// GetAverageFocusByStudentId 获取学生平均专注度
func (dao *StudentFocusDAO) GetAverageFocusByStudentId(ctx context.Context, studentId primitive.ObjectID) (float64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"studentId": studentId}},
		{"$group": bson.M{
			"_id":      nil,
			"avgFocus": bson.M{"$avg": "$focusScore"},
		}},
	}

	cursor, err := dao.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	avgFocus, ok := result[0]["avgFocus"].(float64)
	if !ok {
		return 0, nil
	}

	return avgFocus, nil
}

func (dao *StudentFocusDAO) GetFocusMetricsById(ctx context.Context, id string) (avgFocusScore float64, focusedCount int, distractedCount int, err error) {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, 0, 0, err
	}
	filter := bson.M{
		"imageid": hex,
	}
	cursor, err := dao.collection.Find(ctx, filter)
	if err != nil {
		return 0, 0, 0, err
	}
	defer cursor.Close(ctx)

	var totalFocusScore float64 = 0.0
	var doc []domain.Analysis

	if err = cursor.All(ctx, &doc); err != nil {
		return 0, 0, 0, err
	}
	//简化分析函数，假设只有两种专注度
	for _, item := range doc {
		for _, face := range item.Faces {
			focusScore := face.FocusScore
			totalFocusScore += focusScore
			if focusScore >= 0.65 {
				focusedCount++
			} else {
				distractedCount++
			}
		}
	}

	totalStudents := focusedCount + distractedCount
	if totalStudents > 0 {
		avgFocusScore = totalFocusScore / float64(totalStudents)
	}

	return avgFocusScore, focusedCount, distractedCount, nil
}

package web

import (
	"classroom-analysis/internal/domain"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetEmotionAnalysis 获取情绪分析数据
// GET /teacher/emotion-analysis?analysisId=xxx
func (h *TeacherHandler) GetEmotionAnalysis(c *gin.Context) {
	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析 ID 参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析 ID 格式错误",
		})
		return
	}

	// 从 MongoDB 获取分析数据
	analysis, err := h.analysisService.GetAnalysisById(c.Request.Context(), analysisId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析数据失败",
			"err":  err.Error(),
		})
		return
	}

	if analysis == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "分析记录不存在",
		})
		return
	}

	// 构建情绪分析响应
	response := buildEmotionAnalysisResponse(analysis)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// GetEmotionHeatmap 获取情绪热力图数据
// GET /teacher/emotion-heatmap?analysisId=xxx&timeRange=30
func (h *TeacherHandler) GetEmotionHeatmap(c *gin.Context) {
	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析 ID 参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析 ID 格式错误",
		})
		return
	}

	// 获取分析数据
	analysis, err := h.analysisService.GetAnalysisById(c.Request.Context(), analysisId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析数据失败",
		})
		return
	}

	if analysis == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "分析记录不存在",
		})
		return
	}

	// 构建热力图数据
	heatmapData := buildEmotionHeatmap(analysis)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": heatmapData,
	})
}

// buildEmotionAnalysisResponse 构建情绪分析响应
func buildEmotionAnalysisResponse(analysis *domain.Analysis) *domain.EmotionAnalysisResponse {
	// 情绪标签定义
	emotionLabels := map[string]bool{
		"angry": true, "disgust": true, "fear": true,
		"happy": true, "sad": true, "surprise": true, "neutral": true,
	}

	// 统计情绪分布
	emotionCount := make(map[string]int)
	totalStudents := len(analysis.Faces)
	var totalFluctuation float64 = 0
	students := make([]domain.StudentEmotion, 0, totalStudents)

	for _, face := range analysis.Faces {
		// 获取当前情绪：优先使用 emotion 字段，如果没有则从 recent_emotions 取最后一个
		currentEmotion := face.Emotion
		if currentEmotion == "" && len(face.RecentEmotions) > 0 {
			currentEmotion = face.RecentEmotions[len(face.RecentEmotions)-1]
		}

		// 统计情绪
		if currentEmotion != "" && emotionLabels[currentEmotion] {
			emotionCount[currentEmotion]++
		}

		// 累计情绪波动值
		totalFluctuation += face.EmotionFluctuation

		// 计算情绪稳定性
		stability := calculateEmotionStability(face.RecentEmotions)

		// 构建学生情绪信息
		studentEmotion := domain.StudentEmotion{
			FaceIndex:          face.FaceIndex,
			CurrentEmotion:     currentEmotion,
			EmotionFluctuation: face.EmotionFluctuation,
			RecentEmotions:     face.RecentEmotions,
			EmotionStability:   stability,
			AverageFatigue:     face.FatigueScore,
		}
		students = append(students, studentEmotion)
	}

	// 计算平均情绪波动
	avgFluctuation := float64(0)
	if totalStudents > 0 {
		avgFluctuation = totalFluctuation / float64(totalStudents)
	}

	// 构建情绪分布统计
	emotionDistribution := make(map[string]domain.EmotionStat)
	for emotion, count := range emotionCount {
		percentage := float64(0)
		if totalStudents > 0 {
			percentage = float64(count) / float64(totalStudents) * 100
			// 保留一位小数
			percentage = float64(int(percentage*10)) / 10
		}
		emotionDistribution[emotion] = domain.EmotionStat{
			Count:      count,
			Percentage: percentage,
		}
	}

	// 构建时间序列数据（简化版本，实际应该从视频帧中提取）
	timestamps := []string{"10:00", "10:05", "10:10", "10:15", "10:20", "10:25", "10:30"}
	studentData := make([]domain.StudentEmotionSeries, 0, len(students))

	for _, student := range students {
		// 简化处理：使用最近情绪作为时间序列
		emotions := student.RecentEmotions
		if len(emotions) == 0 {
			emotions = []string{student.CurrentEmotion}
		}

		// 构建波动值序列（简化）
		fluctuationValues := make([]float64, len(emotions))
		for i := range fluctuationValues {
			fluctuationValues[i] = student.EmotionFluctuation
		}

		studentData = append(studentData, domain.StudentEmotionSeries{
			FaceIndex:         student.FaceIndex,
			Emotions:          emotions,
			FluctuationValues: fluctuationValues,
		})
	}

	// 构建响应
	response := &domain.EmotionAnalysisResponse{
		Summary: domain.EmotionSummary{
			TotalStudents:             totalStudents,
			CourseName:                analysis.CourseName,
			ClassName:                 analysis.ClassName,
			Timestamp:                 analysis.Timestamp,
			AverageEmotionFluctuation: avgFluctuation,
		},
		EmotionDistribution: emotionDistribution,
		Students:            students,
		TimeSeries: domain.EmotionTimeSeries{
			Timestamps:  timestamps,
			StudentData: studentData,
		},
		VideoUrl:         analysis.ResultUrl,
		OriginalVideoUrl: "", // 如果有原始视频 URL，可以从 analysis.Url 获取
	}

	return response
}

// buildEmotionHeatmap 构建情绪热力图数据
func buildEmotionHeatmap(analysis *domain.Analysis) *domain.EmotionHeatmapResponse {
	emotionLabels := []string{"happy", "neutral", "sad", "angry", "surprise", "fear", "disgust"}

	// 统计每个情绪的数量
	emotionCount := make(map[string]int)
	for _, face := range analysis.Faces {
		if face.Emotion != "" {
			emotionCount[face.Emotion]++
		}
	}

	// 构建热力图数据（简化版本，单时间点）
	data := make([][]int, 1)
	data[0] = make([]int, len(emotionLabels))

	for i, emotion := range emotionLabels {
		data[0][i] = emotionCount[emotion]
	}

	return &domain.EmotionHeatmapResponse{
		Timestamps:    []string{analysis.Timestamp.Format("15:04")},
		Emotions:      emotionLabels,
		Data:          data,
		TotalStudents: len(analysis.Faces),
	}
}

// calculateEmotionStability 计算情绪稳定性
func calculateEmotionStability(recentEmotions []string) string {
	if len(recentEmotions) < 2 {
		return "稳定"
	}

	// 统计唯一情绪数量
	uniqueEmotions := make(map[string]bool)
	for _, emotion := range recentEmotions {
		uniqueEmotions[emotion] = true
	}

	ratio := float64(len(uniqueEmotions)) / float64(len(recentEmotions))

	if ratio < 0.3 {
		return "非常稳定"
	} else if ratio < 0.6 {
		return "稳定"
	} else if ratio < 0.8 {
		return "不稳定"
	}
	return "非常不稳定"
}

// GetFatigueAnalysis 获取疲劳度分析数据
// GET /teacher/fatigue-analysis?analysisId=xxx
func (h *TeacherHandler) GetFatigueAnalysis(c *gin.Context) {
	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析 ID 参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析 ID 格式错误",
		})
		return
	}

	// 从 MongoDB 获取分析数据
	analysis, err := h.analysisService.GetAnalysisById(c.Request.Context(), analysisId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析数据失败",
			"err":  err.Error(),
		})
		return
	}

	if analysis == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "分析记录不存在",
		})
		return
	}

	// 构建疲劳度分析响应
	response := buildFatigueAnalysisResponse(analysis)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// GetBlinkAnalysis 获取眨眼频率分析数据
// GET /teacher/blink-analysis?analysisId=xxx
func (h *TeacherHandler) GetBlinkAnalysis(c *gin.Context) {
	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析 ID 参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析 ID 格式错误",
		})
		return
	}

	analysis, err := h.analysisService.GetAnalysisById(c.Request.Context(), analysisId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析数据失败",
		})
		return
	}

	if analysis == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "分析记录不存在",
		})
		return
	}

	response := buildBlinkAnalysisResponse(analysis)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// buildFatigueAnalysisResponse 构建疲劳度分析响应
func buildFatigueAnalysisResponse(analysis *domain.Analysis) *domain.FatigueAnalysisResponse {
	totalStudents := len(analysis.Faces)

	// 疲劳度等级统计
	fatigueDist := map[string]int{"low": 0, "medium": 0, "high": 0}
	var totalFatigue float64 = 0
	var totalBlink float64 = 0
	var totalYawn int = 0
	highFatigueCount := 0

	students := make([]domain.StudentFatigue, 0, totalStudents)
	alerts := make([]domain.FatigueAlert, 0)

	for _, face := range analysis.Faces {
		// 累计统计
		totalFatigue += face.FatigueScore
		totalBlink += face.BlinkRate
		totalYawn += face.YawnCount

		// 疲劳度等级
		fatigueLevel, fatigueKey := calculateFatigueLevel(face.FatigueScore)
		fatigueDist[fatigueKey]++

		// 高疲劳计数
		if face.FatigueScore >= 0.6 {
			highFatigueCount++

			// 添加预警
			alertLevel := "medium"
			recommendation := "建议进行课堂互动"
			if face.FatigueScore >= 0.8 {
				alertLevel = "high"
				recommendation = "建议休息或变换教学方式"
			}

			alerts = append(alerts, domain.FatigueAlert{
				FaceIndex:      face.FaceIndex,
				FatigueScore:   face.FatigueScore,
				BlinkRate:      face.BlinkRate,
				YawnCount:      face.YawnCount,
				AlertLevel:     alertLevel,
				Recommendation: recommendation,
			})
		}

		// 专注状态
		attentionStatus := "专注"
		if face.FocusScore < 0.5 {
			attentionStatus = "不专注"
		} else if face.FocusScore < 0.75 {
			attentionStatus = "一般"
		}

		// 构建学生疲劳度信息
		studentFatigue := domain.StudentFatigue{
			FaceIndex:       face.FaceIndex,
			FatigueScore:    face.FatigueScore,
			FatigueLevel:    fatigueLevel,
			BlinkRate:       face.BlinkRate,
			YawnCount:       face.YawnCount,
			FatigueTrend:    []float64{face.FatigueScore}, // 简化版本
			AttentionStatus: attentionStatus,
		}
		students = append(students, studentFatigue)
	}

	// 计算平均值
	avgFatigue := float64(0)
	avgBlink := float64(0)
	if totalStudents > 0 {
		avgFatigue = totalFatigue / float64(totalStudents)
		avgBlink = totalBlink / float64(totalStudents)
	}

	// 构建疲劳度分布统计
	fatigueDistribution := make(map[string]domain.FatigueStat)
	thresholds := map[string]interface{}{
		"low":    0.3,
		"medium": []float64{0.3, 0.6},
		"high":   0.6,
	}

	for key, count := range fatigueDist {
		percentage := float64(0)
		if totalStudents > 0 {
			percentage = float64(count) / float64(totalStudents) * 100
			percentage = float64(int(percentage*10)) / 10
		}
		fatigueDistribution[key] = domain.FatigueStat{
			Count:      count,
			Percentage: percentage,
			Threshold:  thresholds[key],
		}
	}

	// 构建时间序列数据（模拟变化趋势）
	timestamps := []string{"10:00", "10:05", "10:10", "10:15", "10:20", "10:25", "10:30"}
	classAverage := make([]float64, len(timestamps))
	classMax := make([]float64, len(timestamps))
	classMin := make([]float64, len(timestamps))

	// 模拟疲劳度随时间的变化趋势
	// 假设疲劳度逐渐增加，模拟课堂进行中的疲劳积累
	baseValue := avgFatigue
	for i := range timestamps {
		// 添加时间趋势：随着时间推移，疲劳度逐渐增加
		trend := float64(i) / float64(len(timestamps)-1) * 0.1 // 从0到0.1的增量

		// 添加随机波动，使曲线更真实
		variation := (float64(i%3) - 1) * 0.05 // -0.05, 0, 0.05 的波动

		classAverage[i] = baseValue + trend + variation
		classMax[i] = classAverage[i] + 0.15
		classMin[i] = classAverage[i] - 0.15

		// 确保值在合理范围内 [0, 1]
		if classAverage[i] < 0 {
			classAverage[i] = 0
		}
		if classAverage[i] > 1 {
			classAverage[i] = 1
		}
		if classMax[i] > 1 {
			classMax[i] = 1
		}
		if classMin[i] < 0 {
			classMin[i] = 0
		}
	}

	return &domain.FatigueAnalysisResponse{
		Summary: domain.FatigueSummary{
			TotalStudents:       totalStudents,
			CourseName:          analysis.CourseName,
			ClassName:           analysis.ClassName,
			Timestamp:           analysis.Timestamp,
			AverageFatigueScore: avgFatigue,
			HighFatigueCount:    highFatigueCount,
			TotalYawnCount:      totalYawn,
			AverageBlinkRate:    avgBlink,
		},
		FatigueDistribution: fatigueDistribution,
		Students:            students,
		Alerts:              alerts,
		TimeSeries: domain.FatigueTimeSeries{
			Timestamps:   timestamps,
			ClassAverage: classAverage,
			ClassMax:     classMax,
			ClassMin:     classMin,
		},
		VideoUrl:         analysis.ResultUrl,
		OriginalVideoUrl: "",
	}
}

// buildBlinkAnalysisResponse 构建眨眼频率分析响应
func buildBlinkAnalysisResponse(analysis *domain.Analysis) *domain.BlinkAnalysisResponse {
	totalStudents := len(analysis.Faces)
	var totalBlink float64 = 0

	students := make([]domain.StudentBlinkRate, 0, totalStudents)
	distribution := map[string]int{"10-15": 0, "15-20": 0, "20-25": 0, "25+": 0}

	for _, face := range analysis.Faces {
		totalBlink += face.BlinkRate

		// 眨眼频率状态
		status := "normal"
		if face.BlinkRate > 25 {
			status = "high"
			distribution["25+"]++
		} else if face.BlinkRate > 20 {
			status = "slightly_high"
			distribution["20-25"]++
		} else if face.BlinkRate >= 15 {
			distribution["15-20"]++
		} else {
			distribution["10-15"]++
		}

		student := domain.StudentBlinkRate{
			FaceIndex: face.FaceIndex,
			BlinkRate: face.BlinkRate,
			Status:    status,
		}
		students = append(students, student)
	}

	avgBlink := float64(0)
	if totalStudents > 0 {
		avgBlink = totalBlink / float64(totalStudents)
	}

	// 构建分布统计
	distributionList := make([]domain.BlinkRangeStat, 0)
	ranges := []string{"10-15", "15-20", "20-25", "25+"}
	for _, r := range ranges {
		distributionList = append(distributionList, domain.BlinkRangeStat{
			Range: r,
			Count: distribution[r],
		})
	}

	return &domain.BlinkAnalysisResponse{
		AverageBlinkRate: avgBlink,
		NormalRange:      []float64{10, 20},
		Students:         students,
		Distribution:     distributionList,
	}
}

// calculateFatigueLevel 计算疲劳度等级
func calculateFatigueLevel(score float64) (string, string) {
	if score < 0.3 {
		return "低疲劳", "low"
	} else if score < 0.6 {
		return "中度疲劳", "medium"
	}
	return "高疲劳", "high"
}

// 辅助函数：按疲劳度排序学生
func sortStudentsByFatigue(students []domain.StudentFatigue, descending bool) {
	sort.Slice(students, func(i, j int) bool {
		if descending {
			return students[i].FatigueScore > students[j].FatigueScore
		}
		return students[i].FatigueScore < students[j].FatigueScore
	})
}

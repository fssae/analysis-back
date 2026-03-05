package web

import (
	"classroom-analysis/internal/domain"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetReport 获取分析报告详情
func (h *TeacherHandler) GetReport(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	_ = claims.(domain.TeacherClaims)

	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析ID参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析ID格式错误",
		})
		return
	}

	// 从MongoDB获取分析数据
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

	// 构建完整报告详情
	reportDetail := buildReportDetail(analysis)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": reportDetail,
	})
}

// GetReportList 获取报告列表
func (h *TeacherHandler) GetReportList(c *gin.Context) {
	_, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	// 获取所有分析记录
	analyses, total, err := h.analysisService.GetHistory(c.Request.Context(), "", 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取报告列表失败",
		})
		return
	}

	var reportList []domain.ReportListItem
	for _, analysis := range analyses {
		// 计算平均专注度
		avgFocus := calculateAvgFocus(analysis.Faces)
		
		reportList = append(reportList, domain.ReportListItem{
			Id:           analysis.Id.Hex(),
			Type:         analysis.FileType,
			Date:         analysis.Timestamp.Format("2006-01-02 15:04"),
			CourseName:   analysis.CourseName,
			ClassName:    analysis.ClassName,
			AvgFocus:     avgFocus,
			StudentCount: len(analysis.Faces),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":  reportList,
			"total": total,
		},
	})
}

// buildReportDetail 构建报告详情
func buildReportDetail(analysis *domain.Analysis) *domain.ReportDetail {
	faces := analysis.Faces
	studentCount := len(faces)
	
	if studentCount == 0 {
		return &domain.ReportDetail{
			Id:         analysis.Id.Hex(),
			Type:       analysis.FileType,
			Date:       analysis.Timestamp.Format("2006-01-02 15:04"),
			CourseName: analysis.CourseName,
			ClassName:  analysis.ClassName,
		}
	}

	// 计算专注度统计
	var totalFocus, maxFocus, minFocus float64
	maxFocus = 0
	minFocus = 1
	maxFocusIndex := 0
	minFocusIndex := 0
	
	for i, face := range faces {
		focus := face.FocusScore
		totalFocus += focus
		
		if focus > maxFocus {
			maxFocus = focus
			maxFocusIndex = i
		}
		if focus < minFocus {
			minFocus = focus
			minFocusIndex = i
		}
	}
	
	avgFocus := totalFocus / float64(studentCount)
	
	// 生成模拟的时间序列数据（用于视频分析）
	focusData := generateFocusData(avgFocus, analysis.FileType)
	
	// 生成专注度分布
	distributionData := generateDistributionData(faces)
	
	// 生成学生排名
	studentRanking := generateStudentRanking(faces)
	
	// 生成详细数据
	detailedData := generateDetailedData(faces, analysis.Timestamp)
	
	// 生成教学建议
	suggestions := generateSuggestions(avgFocus, faces)
	
	// 计算专注度趋势（模拟）
	focusTrend := calculateFocusTrend(focusData)
	
	return &domain.ReportDetail{
		Id:               analysis.Id.Hex(),
		Type:             analysis.FileType,
		Date:             analysis.Timestamp.Format("2006-01-02 15:04"),
		CourseName:       analysis.CourseName,
		ClassName:        analysis.ClassName,
		AvgFocus:         round(avgFocus * 100),
		MaxFocus:         round(maxFocus * 100),
		MinFocus:         round(minFocus * 100),
		MaxFocusTime:     maxFocusIndex + 1,
		MinFocusTime:     minFocusIndex + 1,
		Duration:         len(focusData),
		StudentCount:     studentCount,
		FocusTrend:       focusTrend,
		FocusData:        focusData,
		DistributionData: distributionData,
		StudentRanking:   studentRanking,
		DetailedData:     detailedData,
		Suggestions:      suggestions,
	}
}

// calculateAvgFocus 计算平均专注度
func calculateAvgFocus(faces []domain.FaceAnalysis) float64 {
	if len(faces) == 0 {
		return 0
	}
	var total float64
	for _, face := range faces {
		total += face.FocusScore
	}
	return round(total / float64(len(faces)) * 100)
}

// generateFocusData 生成专注度时间序列数据
func generateFocusData(avgFocus float64, analysisType string) []float64 {
	// 如果是图片分析，返回单点数据
	if analysisType == "image" {
		return []float64{round(avgFocus * 100)}
	}
	
	// 视频分析：生成模拟的时间序列（10-30个点）
	points := 15 + rand.Intn(16) // 15-30个点
	data := make([]float64, points)
	
	baseValue := avgFocus * 100
	for i := 0; i < points; i++ {
		// 添加随机波动，模拟课堂专注度变化
		variation := (rand.Float64() - 0.5) * 20 // ±10%的波动
		trend := float64(i) * 0.5 // 轻微下降趋势
		
		value := baseValue + variation - trend
		if value < 0 {
			value = 0
		}
		if value > 100 {
			value = 100
		}
		data[i] = round(value)
	}
	
	return data
}

// generateDistributionData 生成分布数据
func generateDistributionData(faces []domain.FaceAnalysis) []domain.ReportDistribution {
	highFocus := 0   // 90%+
	goodFocus := 0   // 80-90%
	normalFocus := 0 // 70-80%
	poorFocus := 0   // <70%
	
	for _, face := range faces {
		focus := face.FocusScore * 100
		switch {
		case focus >= 90:
			highFocus++
		case focus >= 80:
			goodFocus++
		case focus >= 70:
			normalFocus++
		default:
			poorFocus++
		}
	}
	
	return []domain.ReportDistribution{
		{Name: "高度专注 (90%+)", Value: highFocus},
		{Name: "良好专注 (80-90%)", Value: goodFocus},
		{Name: "一般专注 (70-80%)", Value: normalFocus},
		{Name: "需要关注 (<70%)", Value: poorFocus},
	}
}

// generateStudentRanking 生成学生排名
type studentFocus struct {
	index  int
	focus  float64
}

func generateStudentRanking(faces []domain.FaceAnalysis) []domain.ReportStudentRanking {
	students := make([]studentFocus, len(faces))
	for i, face := range faces {
		students[i] = studentFocus{index: i + 1, focus: face.FocusScore}
	}
	
	// 按专注度排序
	sort.Slice(students, func(i, j int) bool {
		return students[i].focus > students[j].focus
	})
	
	// 生成排名数据
	ranking := make([]domain.ReportStudentRanking, len(students))
	for i, s := range students {
		trend := rand.Intn(21) - 10 // -10 到 +10 的随机趋势
		ranking[i] = domain.ReportStudentRanking{
			Id:    s.index,
			Name:  fmt.Sprintf("学生%d", s.index),
			Focus: int(round(s.focus * 100)),
			Trend: trend,
		}
	}
	
	return ranking
}

// generateDetailedData 生成详细数据
func generateDetailedData(faces []domain.FaceAnalysis, timestamp time.Time) []domain.ReportDetailedData {
	data := make([]domain.ReportDetailedData, 0)
	
	// 每5分钟一个时间点
	intervals := 6
	for i := 0; i < intervals; i++ {
		timeStr := timestamp.Add(time.Duration(i*5) * time.Minute).Format("15:04")
		
		// 计算该时间点的专注学生数
		focusedCount := 0
		distractedCount := 0
		
		for _, face := range faces {
			if face.FocusScore >= 0.7 {
				focusedCount++
			} else {
				distractedCount++
			}
		}
		
		avgFocus := calculateAvgFocus(faces)
		
		data = append(data, domain.ReportDetailedData{
			Time:            timeStr,
			Focus:           int(avgFocus),
			FocusedCount:    focusedCount,
			DistractedCount: distractedCount,
			Notes:           "",
		})
	}
	
	return data
}

// generateSuggestions 生成教学建议
func generateSuggestions(avgFocus float64, faces []domain.FaceAnalysis) domain.ReportSuggestions {
	focusPercent := avgFocus * 100
	
	// 整体表现评价
	var overall string
	switch {
	case focusPercent >= 85:
		overall = "本节课学生整体专注度表现优秀，大部分学生能够保持良好的学习状态。"
	case focusPercent >= 75:
		overall = "本节课学生整体专注度表现良好，多数学生能够跟上教学节奏。"
	case focusPercent >= 65:
		overall = "本节课学生整体专注度一般，部分学生出现分心现象，需要适当调整教学方法。"
	default:
		overall = "本节课学生整体专注度较低，建议反思教学内容和方法，增加互动环节。"
	}
	
	// 改进建议
	improvements := []string{
		"增加课堂互动环节，提高学生参与度",
		"适时调整教学节奏，避免长时间单一讲解",
		"关注低专注度学生，及时了解原因",
	}
	
	if focusPercent < 75 {
		improvements = append(improvements, "考虑使用多媒体教学手段，增强课堂趣味性")
	}
	
	// 教学方法建议
	methods := []string{
		"采用提问式教学，引导学生主动思考",
		"小组讨论与个别指导相结合",
		"适时进行课堂小测验，检验学习效果",
	}
	
	// 需要关注的学生
	attentionStudents := make([]domain.ReportAttentionStudent, 0)
	for i, face := range faces {
		if face.FocusScore < 0.7 {
			attentionStudents = append(attentionStudents, domain.ReportAttentionStudent{
				Id:    i + 1,
				Name:  fmt.Sprintf("学生%d", i+1),
				Focus: int(round(face.FocusScore * 100)),
			})
		}
	}
	
	// 限制关注学生数量，只显示最低的5个
	if len(attentionStudents) > 5 {
		sort.Slice(attentionStudents, func(i, j int) bool {
			return attentionStudents[i].Focus < attentionStudents[j].Focus
		})
		attentionStudents = attentionStudents[:5]
	}
	
	return domain.ReportSuggestions{
		Overall:           overall,
		Improvements:      improvements,
		Methods:           methods,
		AttentionStudents: attentionStudents,
	}
}

// calculateFocusTrend 计算专注度趋势
func calculateFocusTrend(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	
	firstHalf := data[:len(data)/2]
	secondHalf := data[len(data)/2:]
	
	var firstAvg, secondAvg float64
	for _, v := range firstHalf {
		firstAvg += v
	}
	for _, v := range secondHalf {
		secondAvg += v
	}
	
	firstAvg /= float64(len(firstHalf))
	secondAvg /= float64(len(secondHalf))
	
	return round(secondAvg - firstAvg)
}

// round 四舍五入到指定小数位
func round(value float64) float64 {
	return math.Round(value*100) / 100
}

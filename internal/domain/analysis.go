package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Analysis struct {
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// 保持与前端 FormData 传参一致的 JSON 标签
	ImageId      primitive.ObjectID `json:"imageId" bson:"imageid"`
	AnalysisType string             `json:"analysisType" bson:"analysisType"`
	Faces        []FaceAnalysis     `json:"faces" bson:"faces"`
	ResultUrl    string             `json:"resultUrl" bson:"result_url"`
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
	AnalysisMode string             `json:"analysisMode" bson:"analysisMode"`
	ClassName    string             `json:"className" bson:"className"`
	CourseName   string             `json:"courseName" bson:"courseName"`
	Description  string             `json:"description" bson:"description"`
	FileName     string             `json:"fileName" bson:"filename"`
	AnalysisName string             `json:"analysisName" bson:"analysisName"`
	FileType     string             `json:"fileType" bson:"filetype"`
}
type AnalysisResult struct {
	AvgFocus           float64 `json:"avgFocus"`
	FocusedStudents    int     `json:"focusedStudents"`
	DistractedStudents int     `json:"distractedStudents"`
	//ImageCount         int     `json:"imageCount"`
	//DistributionData   []Distribution `json:"distributionData"`
	//ImageResults       []ImageResult  `json:"imageResults"`
	ResultURL string `json:"resultURL"`
}
type UpdateConfigRequest struct {
	AnalysisId   primitive.ObjectID `json:"analysisId" bson:"analysisId"`     // 分析ID
	AnalysisMode string             `json:"analysisMode" bson:"analysisMode"` // 分析模式
	ClassName    string             `json:"className" bson:"className"`       // 班级名称
	CourseName   string             `json:"courseName" bson:"courseName"`     // 课程名称
	Description  string             `json:"description" bson:"description"`   // 描述信息
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
}

type UpdateAnalysisNameRequest struct {
	FileName string `json:"fileName" bson:"filename"` // 新的文件名
	ImageId  string `json:"imageid" bson:"imageid"`   // 图片ID
}
type UpdateStatus struct {
	TeacherId           primitive.ObjectID `json:"teacherId" bson:"teacherId"`
	TaskId              string             `json:"taskId" bson:"taskId"`
	ImageId             primitive.ObjectID `json:"imageId" bson:"imageId"`
	Status              string             `json:"status" bson:"status"`
	ConfidenceThreshold float64            `json:"confidenceThreshold" bson:"confidenceThreshold"`
	ResultUrl           string             `json:"resultUrl" bson:"resultUrl"`
	ErrorMsg            string             `json:"errorMsg" bson:"errorMsg"`
}

//type Distribution struct {
//	ID         string  `json:"id"`
//	FocusScore float64 `json:"focusScore"`
//	ClassNum   int     `json:"classNum"`
//}
//
//type ImageResult struct {
//	ID       string `json:"id"`
//	ResultURL string `json:"resultURL"`
//}

// FaceAnalysis 人脸分析详情
type FaceAnalysis struct {
	FaceIndex          int      `json:"faceIndex" bson:"face_index"`
	FocusScore         float64  `json:"focusScore" bson:"focus_score"`
	Confidence         float64  `json:"confidence" bson:"confidence"`
	ClassNum           int      `json:"classNum,omitempty" bson:"class_num,omitempty"`
	Emotion            string   `json:"emotion,omitempty" bson:"emotion,omitempty"`
	FatigueScore       float64  `json:"fatigueScore,omitempty" bson:"fatigue_score,omitempty"`
	BlinkRate          float64  `json:"blinkRate,omitempty" bson:"blink_rate,omitempty"`
	YawnCount          int      `json:"yawnCount,omitempty" bson:"yawn_count,omitempty"`
	EmotionFluctuation float64  `json:"emotionFluctuation,omitempty" bson:"emotion_fluctuation,omitempty"`
	RecentEmotions     []string `json:"recentEmotions,omitempty" bson:"recent_emotions,omitempty"`
}

// VideoAnalysis 视频分析详情
type VideoAnalysis struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AnalysisId primitive.ObjectID `json:"analysisId" bson:"analysisId"`
	FocusTrend []FocusPoint       `json:"focusTrend" bson:"focusTrend"`
	Duration   int                `json:"duration" bson:"duration"` // 视频时长(秒)
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

// ImageAnalysis 图片分析详情
type ImageAnalysis struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AnalysisId primitive.ObjectID `json:"analysisId" bson:"analysisId"`
	FocusScore float64            `json:"focusScore" bson:"focusScore"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

// FocusPoint 专注度数据点
type FocusPoint struct {
	Time  string  `json:"time" bson:"time"`   // 时间点 (HH:MM)
	Focus float64 `json:"focus" bson:"focus"` // 专注度分数
}

// Student 学生信息
type Student struct {
	Id        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	StudentId string             `json:"studentId" bson:"studentId"`
	Name      string             `json:"name" bson:"name"`
	Class     string             `json:"class" bson:"class"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// StudentFocus 学生专注度记录
type StudentFocus struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	StudentId  primitive.ObjectID `json:"studentId" bson:"studentId"`
	AnalysisId primitive.ObjectID `json:"analysisId" bson:"analysisId"`
	FocusScore float64            `json:"focusScore" bson:"focusScore"`
	Date       time.Time          `json:"date" bson:"date"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

// DashboardData 仪表盘数据
type DashboardData struct {
	TotalVideos       int64          `json:"totalVideos"`
	TotalImages       int64          `json:"totalImages"`
	TotalStudents     int64          `json:"totalStudents"`
	TotalAnalysis     int64          `json:"totalAnalysis"`
	RecentRecords     []*Analysis    `json:"recentRecords"`
	FocusTrend        []float64      `json:"focusTrend"`
	FocusDistribution map[string]int ` json:"focusDistribution"`
}

// FocusRange 专注度分布
type FocusRange struct {
	Range string `json:"range" bson:"range"`
	Count int    `json:"count" bson:"count"`
}

// RecentRecord 最近记录
type RecentRecord struct {
	Id    primitive.ObjectID `json:"id"`
	Type  string             `json:"type"`
	Title string             `json:"title"`
	Date  time.Time          `json:"date"`
	Focus float64            `json:"focus"`
}

// Report 报告
type Report struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AnalysisId primitive.ObjectID `json:"analysisId" bson:"analysisId"`
	Title      string             `json:"title" bson:"title"`
	FilePath   string             `json:"filePath" bson:"filePath"`
	Type       string             `json:"type" bson:"type"` // pdf, word
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

// ========== 分析报告相关结构体 ==========

// ReportListItem 报告列表项
type ReportListItem struct {
	Id           string  `json:"id"`
	Type         string  `json:"type"`
	Date         string  `json:"date"`
	CourseName   string  `json:"courseName"`
	ClassName    string  `json:"className"`
	AvgFocus     float64 `json:"avgFocus"`
	StudentCount int     `json:"studentCount"`
}

// ReportDetail 报告详情
type ReportDetail struct {
	Id               string                 `json:"id"`
	Type             string                 `json:"type"`
	Date             string                 `json:"date"`
	CourseName       string                 `json:"courseName"`
	ClassName        string                 `json:"className"`
	AvgFocus         float64                `json:"avgFocus"`
	MaxFocus         float64                `json:"maxFocus"`
	MinFocus         float64                `json:"minFocus"`
	MaxFocusTime     int                    `json:"maxFocusTime"`
	MinFocusTime     int                    `json:"minFocusTime"`
	Duration         int                    `json:"duration"`
	StudentCount     int                    `json:"studentCount"`
	FocusTrend       float64                `json:"focusTrend"`
	FocusData        []float64              `json:"focusData"`
	DistributionData []ReportDistribution   `json:"distributionData"`
	StudentRanking   []ReportStudentRanking `json:"studentRanking"`
	DetailedData     []ReportDetailedData   `json:"detailedData"`
	Suggestions      ReportSuggestions      `json:"suggestions"`
}

// ReportDistribution 报告分布数据
type ReportDistribution struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// ReportStudentRanking 学生排名
type ReportStudentRanking struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Focus int    `json:"focus"`
	Trend int    `json:"trend"`
}

// ReportDetailedData 详细数据
type ReportDetailedData struct {
	Time            string `json:"time"`
	Focus           int    `json:"focus"`
	FocusedCount    int    `json:"focusedCount"`
	DistractedCount int    `json:"distractedCount"`
	Notes           string `json:"notes"`
}

// ReportSuggestions 教学建议
type ReportSuggestions struct {
	Overall           string                   `json:"overall"`
	Improvements      []string                 `json:"improvements"`
	Methods           []string                 `json:"methods"`
	AttentionStudents []ReportAttentionStudent `json:"attentionStudents"`
}

// ReportAttentionStudent 需要关注的学生
type ReportAttentionStudent struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Focus int    `json:"focus"`
}

// AnalysisTask 分析任务
type AnalysisTask struct {
	Id           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TaskId       string             `json:"taskId" bson:"taskId"`
	TeacherId    primitive.ObjectID `json:"teacherId" bson:"teacherId"`
	Title        string             `json:"title" bson:"title"`
	ClassName    string             `json:"className" bson:"className"`
	StudentCount int                `json:"studentCount" bson:"studentCount"`
	AnalysisMode string             `json:"analysisMode" bson:"analysisMode"`
	Description  string             `json:"description" bson:"description"`
	Status       string             `json:"status" bson:"status"` // processing, completed, failed
	Progress     int                `json:"progress" bson:"progress"`
	Files        []TaskFile         `json:"files" bson:"files"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// TaskFile 任务文件信息
type TaskFile struct {
	Filename string `json:"filename" bson:"filename"`
	Path     string `json:"path" bson:"path"`
	Url      string `json:"url" bson:"url"`
}

// BoundingBox 边界框
type BoundingBox struct {
	X1 int `json:"x1" bson:"x1"`
	Y1 int `json:"y1" bson:"y1"`
	X2 int `json:"x2" bson:"x2"`
	Y2 int `json:"y2" bson:"y2"`
}

// Landmark 人脸关键点
type Landmark struct {
	X int `json:"x" bson:"x"`
	Y int `json:"y" bson:"y"`
}

// VideoAnalysisDetail 视频分析详情响应
type VideoAnalysisDetail struct {
	Id                 string    `json:"id" bson:"_id"`
	AvgFocus           float64   `json:"avgFocus" bson:"avgFocus"`
	MaxFocus           float64   `json:"maxFocus" bson:"maxFocus"`
	MinFocus           float64   `json:"minFocus" bson:"minFocus"`
	Duration           int       `json:"duration" bson:"duration"`
	FocusedStudents    int       `json:"focusedStudents" bson:"focusedStudents"`
	DistractedStudents int       `json:"distractedStudents" bson:"distractedStudents"`
	ResultURL          string    `json:"ResultURL" bson:"resultURL"`
	FocusData          []float64 `json:"focusData" bson:"focusData"`
	AnalysisTime       time.Time `json:"analysisTime" bson:"analysisTime"`
	CourseName         string    `json:"courseName" bson:"courseName"`
	ClassName          string    `json:"className" bson:"className"`
	StudentCount       int       `json:"studentCount" bson:"studentCount"`
	Accuracy           string    `json:"accuracy" bson:"accuracy"`
}

// 班级分析返回结构体
// 用于 /api/teacher/class-analysis 接口
// focusAvg 为小数，前端需乘以 100 显示百分比
// date 格式为 yyyy-MM-dd
// resultUrl 为分析结果图片地址
type ClassAnalysisItem struct {
	CourseName string  `json:"courseName" bson:"courseName"`
	ClassName  string  `json:"className" bson:"className"`
	FileName   string  `json:"fileName" bson:"fileName"`
	Date       string  `json:"date" bson:"date"`
	ImageId    string  `json:"imageid" bson:"imageid"`
	FocusAvg   float64 `json:"focusAvg" bson:"focusAvg"`
	ResultUrl  string  `json:"resultUrl" bson:"resultUrl"`
}

// ========== 情绪分析相关结构体 ==========

// EmotionAnalysisResponse 情绪分析响应
type EmotionAnalysisResponse struct {
	Summary             EmotionSummary         `json:"summary"`
	EmotionDistribution map[string]EmotionStat `json:"emotionDistribution"`
	Students            []StudentEmotion       `json:"students"`
	TimeSeries          EmotionTimeSeries      `json:"timeSeries"`
	VideoUrl            string                 `json:"videoUrl"`
	OriginalVideoUrl    string                 `json:"originalVideoUrl"`
}

// EmotionSummary 情绪分析摘要
type EmotionSummary struct {
	TotalStudents             int       `json:"totalStudents"`
	CourseName                string    `json:"courseName"`
	ClassName                 string    `json:"className"`
	Timestamp                 time.Time `json:"timestamp"`
	AverageEmotionFluctuation float64   `json:"averageEmotionFluctuation"`
}

// EmotionStat 情绪统计
type EmotionStat struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// StudentEmotion 学生情绪信息
type StudentEmotion struct {
	FaceIndex          int      `json:"faceIndex"`
	CurrentEmotion     string   `json:"currentEmotion"`
	EmotionFluctuation float64  `json:"emotionFluctuation"`
	RecentEmotions     []string `json:"recentEmotions"`
	EmotionStability   string   `json:"emotionStability"`
	AverageFatigue     float64  `json:"averageFatigue"`
}

// EmotionTimeSeries 情绪时间序列
type EmotionTimeSeries struct {
	Timestamps  []string               `json:"timestamps"`
	StudentData []StudentEmotionSeries `json:"studentData"`
}

// StudentEmotionSeries 学生情绪序列
type StudentEmotionSeries struct {
	FaceIndex         int       `json:"faceIndex"`
	Emotions          []string  `json:"emotions"`
	FluctuationValues []float64 `json:"fluctuationValues"`
}

// EmotionHeatmapResponse 情绪热力图响应
type EmotionHeatmapResponse struct {
	Timestamps    []string `json:"timestamps"`
	Emotions      []string `json:"emotions"`
	Data          [][]int  `json:"data"`
	TotalStudents int      `json:"totalStudents"`
}

// ========== 疲劳度分析相关结构体 ==========

// FatigueAnalysisResponse 疲劳度分析响应
type FatigueAnalysisResponse struct {
	Summary             FatigueSummary         `json:"summary"`
	FatigueDistribution map[string]FatigueStat `json:"fatigueDistribution"`
	Students            []StudentFatigue       `json:"students"`
	Alerts              []FatigueAlert         `json:"alerts"`
	TimeSeries          FatigueTimeSeries      `json:"timeSeries"`
	VideoUrl            string                 `json:"videoUrl"`
	OriginalVideoUrl    string                 `json:"originalVideoUrl"`
}

// FatigueSummary 疲劳度分析摘要
type FatigueSummary struct {
	TotalStudents       int       `json:"totalStudents"`
	CourseName          string    `json:"courseName"`
	ClassName           string    `json:"className"`
	Timestamp           time.Time `json:"timestamp"`
	AverageFatigueScore float64   `json:"averageFatigueScore"`
	HighFatigueCount    int       `json:"highFatigueCount"`
	TotalYawnCount      int       `json:"totalYawnCount"`
	AverageBlinkRate    float64   `json:"averageBlinkRate"`
}

// FatigueStat 疲劳度统计
type FatigueStat struct {
	Count      int         `json:"count"`
	Percentage float64     `json:"percentage"`
	Threshold  interface{} `json:"threshold"`
}

// StudentFatigue 学生疲劳度信息
type StudentFatigue struct {
	FaceIndex       int       `json:"faceIndex"`
	FatigueScore    float64   `json:"fatigueScore"`
	FatigueLevel    string    `json:"fatigueLevel"`
	BlinkRate       float64   `json:"blinkRate"`
	YawnCount       int       `json:"yawnCount"`
	FatigueTrend    []float64 `json:"fatigueTrend"`
	AttentionStatus string    `json:"attentionStatus"`
}

// FatigueAlert 疲劳度预警
type FatigueAlert struct {
	FaceIndex      int     `json:"faceIndex"`
	FatigueScore   float64 `json:"fatigueScore"`
	BlinkRate      float64 `json:"blinkRate"`
	YawnCount      int     `json:"yawnCount"`
	AlertLevel     string  `json:"alertLevel"`
	Recommendation string  `json:"recommendation"`
}

// FatigueTimeSeries 疲劳度时间序列
type FatigueTimeSeries struct {
	Timestamps   []string  `json:"timestamps"`
	ClassAverage []float64 `json:"classAverage"`
	ClassMax     []float64 `json:"classMax"`
	ClassMin     []float64 `json:"classMin"`
}

// BlinkAnalysisResponse 眨眼频率分析响应
type BlinkAnalysisResponse struct {
	AverageBlinkRate float64            `json:"averageBlinkRate"`
	NormalRange      []float64          `json:"normalRange"`
	Students         []StudentBlinkRate `json:"students"`
	Distribution     []BlinkRangeStat   `json:"distribution"`
}

// StudentBlinkRate 学生眨眼频率
type StudentBlinkRate struct {
	FaceIndex int     `json:"faceIndex"`
	BlinkRate float64 `json:"blinkRate"`
	Status    string  `json:"status"`
}

// BlinkRangeStat 眨眼频率范围统计
type BlinkRangeStat struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	ID           string  `bson:"_id" json:"id"`
	AvgFatigue   float64 `bson:"avgFatigue" json:"avgFatigue"`
	MaxFatigue   float64 `bson:"maxFatigue" json:"maxFatigue"`
	MinFatigue   float64 `bson:"minFatigue" json:"minFatigue"`
	AvgFocus     float64 `bson:"avgFocus" json:"avgFocus"`
	AvgBlinkRate float64 `bson:"avgBlinkRate" json:"avgBlinkRate"`
	AvgYawnCount float64 `bson:"avgYawnCount" json:"avgYawnCount"`
}

// EmotionTimeSeriesPoint 情绪时间序列数据点
type EmotionTimeSeriesPoint struct {
	ID             string  `bson:"_id" json:"id"`
	AngryCount     int     `bson:"angryCount" json:"angryCount"`
	HappyCount     int     `bson:"happyCount" json:"happyCount"`
	NeutralCount   int     `bson:"neutralCount" json:"neutralCount"`
	SadCount       int     `bson:"sadCount" json:"sadCount"`
	AvgFluctuation float64 `bson:"avgFluctuation" json:"avgFluctuation"`
}

// ========== 全局可视化分析相关结构体 ==========

// GlobalAnalysisResponse 全局分析响应
type GlobalAnalysisResponse struct {
	Overview     GlobalOverview      `json:"overview"`
	ClassStats   []ClassStat         `json:"classStats"`
	CourseStats  []CourseStat        `json:"courseStats"`
	StudentStats []StudentGlobalStat `json:"studentStats"`
	EmotionTrend []EmotionTrendPoint `json:"emotionTrend"`
	FocusTrend   []FocusTrendPoint   `json:"focusTrend"`
	FatigueTrend []FatigueTrendPoint `json:"fatigueTrend"`
}

// GlobalOverview 全局概览统计
type GlobalOverview struct {
	TotalStudents       int64          `json:"totalStudents"`
	TotalAnalysis       int64          `json:"totalAnalysis"`
	TotalVideos         int64          `json:"totalVideos"`
	TotalImages         int64          `json:"totalImages"`
	AverageFocusScore   float64        `json:"averageFocusScore"`
	AverageFatigueScore float64        `json:"averageFatigueScore"`
	HighFatigueStudents int64          `json:"highFatigueStudents"`
	LowFocusStudents    int64          `json:"lowFocusStudents"`
	EmotionDistribution map[string]int `json:"emotionDistribution"`
}

// ClassStat 班级统计
type ClassStat struct {
	ClassName           string         `json:"className"`
	StudentCount        int            `json:"studentCount"`
	AnalysisCount       int            `json:"analysisCount"`
	AverageFocusScore   float64        `json:"averageFocusScore"`
	AverageFatigueScore float64        `json:"averageFatigueScore"`
	EmotionDistribution map[string]int `json:"emotionDistribution"`
	CourseList          []string       `json:"courseList"`
}

// CourseStat 课程统计
type CourseStat struct {
	CourseName          string         `json:"courseName"`
	ClassCount          int            `json:"classCount"`
	AnalysisCount       int            `json:"analysisCount"`
	AverageFocusScore   float64        `json:"averageFocusScore"`
	AverageFatigueScore float64        `json:"averageFatigueScore"`
	EmotionDistribution map[string]int `json:"emotionDistribution"`
}

// StudentGlobalStat 学生全局统计
type StudentGlobalStat struct {
	FaceIndex           int       `json:"faceIndex"`
	ClassName           string    `json:"className"`
	CourseName          string    `json:"courseName"`
	AnalysisCount       int       `json:"analysisCount"`
	AverageFocusScore   float64   `json:"averageFocusScore"`
	AverageFatigueScore float64   `json:"averageFatigueScore"`
	AverageEmotion      string    `json:"averageEmotion"`
	LatestAnalysisTime  time.Time `json:"latestAnalysisTime"`
}

// EmotionTrendPoint 情绪趋势点
type EmotionTrendPoint struct {
	Date     string `json:"date"`
	Happy    int    `json:"happy"`
	Neutral  int    `json:"neutral"`
	Sad      int    `json:"sad"`
	Angry    int    `json:"angry"`
	Surprise int    `json:"surprise"`
	Fear     int    `json:"fear"`
	Disgust  int    `json:"disgust"`
	Total    int    `json:"total"`
}

// FocusTrendPoint 专注度趋势点
type FocusTrendPoint struct {
	Date              string  `json:"date"`
	AverageFocusScore float64 `json:"averageFocusScore"`
	StudentCount      int     `json:"studentCount"`
}

// FatigueTrendPoint 疲劳度趋势点
type FatigueTrendPoint struct {
	Date                string  `json:"date"`
	AverageFatigueScore float64 `json:"averageFatigueScore"`
	HighFatigueCount    int     `json:"highFatigueCount"`
}

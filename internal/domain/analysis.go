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
	FaceIndex  int     `json:"faceIndex" bson:"face_index"`
	FocusScore float64 `json:"focusScore" bson:"focus_score"`
	Confidence float64 `json:"confidence" bson:"confidence"`
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
// focusAvg 为小数，前端需乘以100显示百分比
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

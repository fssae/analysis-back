package domain

import "time"

type TeacherAnalysisRequest struct {
	ImageId             string  `form:"imageId" bson:"imageId" json:"imageId"` // 绑定 form 数据
	ConfidenceThreshold float64 `form:"ConfidenceThreshold" json:"ConfidenceThreshold" bson:"confidenceThreshold"`
	//TaskId       string    `form:"taskId" bson:"taskId" json:"taskId"` //任务唯一标识
	AnalysisType string    `form:"analysisType" bson:"analysisType" json:"analysisType"`
	Url          string    `form:"Url" bson:"Url" json:"Url"`                         // 绑定 form 数据
	ClassName    string    `form:"className" bson:"className" json:"className"`       // 绑定 form 数据
	CourseName   string    `form:"courseName" bson:"courseName" json:"courseName"`    // 课程名称
	Description  string    `form:"description" bson:"description" json:"description"` // 描述信息
	Timestamp    time.Time `form:"timestamp" bson:"timestamp" json:"timestamp"`
}
type KafkaMessage struct {
	Req TeacherAnalysisRequest `json:"req" bson:"req"`
}
type KafkaResp struct {
	ID          string `json:"id"`
	ResultURL   string `json:"result_url"`
	DownloadUrl string `json:"download_url"`
	Status      string `json:"status,omitempty"` // 支持Python端的status字段
}

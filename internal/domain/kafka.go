package domain

import "time"

type TeacherAnalysisRequest struct {
	ImageId             string    `form:"imageId" json:"imageId" bson:"imageId"`
	VideoId             string    `form:"videoId" json:"videoId" bson:"videoId"`
	TaskId              string    `form:"taskId" json:"taskId" bson:"taskId"`
	ConfidenceThreshold float64   `form:"confidenceThreshold" json:"confidenceThreshold" bson:"confidenceThreshold"`
	AnalysisType        string    `form:"analysisType" json:"analysisType" bson:"analysisType"`
	Url                 string    `form:"url" json:"url" bson:"url"`
	ClassName           string    `form:"className" json:"className" bson:"className"`
	CourseName          string    `form:"courseName" json:"courseName" bson:"courseName"`
	Description         string    `form:"description" json:"description" bson:"description"`
	Timestamp           time.Time `form:"timestamp" json:"timestamp" bson:"timestamp"`
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

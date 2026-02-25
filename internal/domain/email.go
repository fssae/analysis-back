package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Email struct {
	TeacherId primitive.ObjectID `json:"teacherId,omitempty" bson:"teacherId" `
	Unread    int                `json:"unRead" bson:"unread"`
	Urls      []EmailUrl         `json:"urls" bson:"urls"`
}
type EmailUrl struct {
	Url         string    `json:"url,omitempty" bson:"url"`
	TaskId      string    `json:"taskId" bson:"taskId"`
	Time        time.Time `json:"time" bson:"time"`
	Confidence  float64   `json:"confidence" bson:"confidence"`
	FileName    string    `json:"fileName,omitempty" bson:"fileName"`       // From Analysis object
	ClassName   string    `json:"className,omitempty" bson:"className"`     // From Analysis object
	CourseName  string    `json:"courseName,omitempty" bson:"courseName"`   // From Analysis object
	Description string    `json:"description,omitempty" bson:"description"` // From Analysis object
	FileType    string    `json:"fileType,omitempty" bson:"fileType"`       // From Analysis object
	ResultUrl   string    `json:"resultUrl,omitempty" bson:"resultUrl"`     // From Analysis object
}
type EmailClearRequest struct {
	TeacherId primitive.ObjectID `json:"teacherId,omitempty" bson:"teacherId" `
}
type EmailClearResponse struct {
	Msg string `json:"msg,omitempty"`
}
type EmailInsertRequest struct {
	TeacherId primitive.ObjectID `json:"teacherId,omitempty" bson:"teacherId" `
}
type GetUrlsRequest struct {
	Page     int `json:"page" bson:"page"`
	PageSize int `json:"pageSize" bson:"pageSize"`
}
type EmailDeleteRequest struct {
	TaskId string `json:"taskId,omitempty" bson:"taskId" `
}

package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Teacher 教师模型
type Teacher struct {
	Id        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TeacherId string             `json:"teacherId" bson:"teacherId"` // 工号
	Name      string             `json:"name" bson:"name"`           // 姓名
	Password  string             `json:"password" bson:"password"`   // 密码
	Email     string             `json:"email" bson:"email"`         // 邮箱
	Phone     string             `json:"phone" bson:"phone"`         // 电话
	Avatar    string             `json:"avatar" bson:"avatar"`       // 头像
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// TeacherLoginRequest 教师登录请求
type TeacherLoginRequest struct {
	TeacherId string `json:"teacherId" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// TeacherRegisterRequest 教师注册请求
type TeacherRegisterRequest struct {
	TeacherId string `json:"teacherId" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// TeacherLoginResponse 教师登录响应
type TeacherLoginResponse struct {
	Token   string   `json:"token"`
	Teacher *Teacher `json:"teacher"`
}

// TeacherSettings 教师设置
type TeacherSettings struct {
	Id           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TeacherId    primitive.ObjectID `json:"teacherId" bson:"teacherId"`
	Email        string             `json:"email" bson:"email"`
	Notification bool               `json:"notification" bson:"notification"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

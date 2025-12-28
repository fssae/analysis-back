package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeacherClaims 教师JWT Claims
type TeacherClaims struct {
	Id        primitive.ObjectID `json:"id"`
	TeacherId string             `json:"teacherId"`
	jwt.RegisteredClaims
}

// NewTeacherClaims 创建教师Claims
func NewTeacherClaims(teacher *Teacher) *TeacherClaims {
	return &TeacherClaims{
		TeacherId: teacher.TeacherId,
		Id:        teacher.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(300 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
}

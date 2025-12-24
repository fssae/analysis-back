package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeacherClaims 教师JWT Claims
type TeacherClaims struct {
	TeacherId string             `json:"teacherId"`
	Id        primitive.ObjectID `json:"id"`
	Name      string             `json:"name"`
	jwt.RegisteredClaims
}

// NewTeacherClaims 创建教师Claims
func NewTeacherClaims(teacher *Teacher) *TeacherClaims {
	return &TeacherClaims{
		TeacherId: teacher.TeacherId,
		Id:        teacher.Id,
		Name:      teacher.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
}

package util

import (
	"classroom-analysis/internal/domain"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetClaims(c *gin.Context) (*domain.TeacherClaims, error) {
	claims, exists := c.Get("claims")
	if !exists {
		zap.String("error", "无法解析token")
		return nil, errors.New("无法解析token")
	}
	claim := claims.(domain.TeacherClaims)
	return &claim, nil
}

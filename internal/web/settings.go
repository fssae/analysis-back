package web

import (
	"classroom-analysis/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SettingsRequest 设置请求结构
type SettingsRequest struct {
	UISettings           domain.UISettings           `json:"uiSettings"`
	AnalysisSettings     domain.AnalysisSettings     `json:"analysisSettings"`
	NotificationSettings domain.NotificationSettings `json:"notificationSettings"`
}

// GetSettings 获取教师设置
func (h *TeacherHandler) GetSettings(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	teacherClaims := claims.(domain.TeacherClaims)
	teacherId := teacherClaims.TeacherId

	// 从数据库获取设置
	settings, err := h.teacherService.GetSettings(c.Request.Context(), teacherId)
	if err != nil {
		// 如果获取失败，返回默认设置
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
			"data": gin.H{
				"uiSettings": gin.H{
					"theme":            "light",
					"sidebarCollapsed": false,
					"enableAnimation":  true,
				},
				"analysisSettings": gin.H{
					"defaultCourse":    "",
					"defaultClass":     "",
					"fatigueThreshold": 70,
					"focusThreshold":   60,
				},
				"notificationSettings": gin.H{
					"enabled":           true,
					"fatigueAlert":      true,
					"focusAlert":        true,
					"emailNotification": false,
				},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": settings,
	})
}

// UpdateSettings 更新教师设置
func (h *TeacherHandler) UpdateSettings(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	teacherClaims := claims.(domain.TeacherClaims)
	teacherId := teacherClaims.TeacherId

	var req SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新设置
	err := h.teacherService.UpdateSettings(c.Request.Context(), teacherId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "设置保存成功",
	})
}

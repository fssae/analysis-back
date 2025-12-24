package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDashboard 获取仪表盘数据
func (h *TeacherHandler) GetDashboard(c *gin.Context) {
	dashboardData, err := h.teacherService.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取仪表盘数据失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": dashboardData,
	})
}

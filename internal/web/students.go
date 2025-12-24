package web

import (
	"classroom-analysis/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStudents 获取学生专注度数据
func (h *TeacherHandler) GetStudents(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	// 验证教师权限（可选，如果需要验证该分析是否属于该教师）
	_ = claims.(domain.TeacherClaims)

	students, err := h.teacherService.GetStudents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取学生数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": students,
	})
}

// FindStudents 获取查询信息的学生信息
func (h *TeacherHandler) FindStudents(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	// 验证教师权限（可选，如果需要验证该分析是否属于该教师）
	_ = claims.(domain.TeacherClaims)

	// 从查询参数获取学生姓名，而不是从请求体
	studentName := c.Query("name")
	if studentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少学生姓名参数",
		})
		return
	}

	students, err := h.teacherService.FindStudents(c.Request.Context(), studentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取学生数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": students,
	})
}

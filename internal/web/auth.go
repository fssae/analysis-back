package web

import (
	"classroom-analysis/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login 教师登录
func (h *TeacherHandler) Login(c *gin.Context) {
	var req domain.TeacherLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}

	resp, err := h.teacherService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "登录成功",
		"data":    resp.Teacher,
		"success": true,
		"token":   resp.Token,
	})
}

// Register 教师注册
func (h *TeacherHandler) Register(c *gin.Context) {
	var req domain.TeacherRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}

	err := h.teacherService.Register(c.Request.Context(), &req)
	if err != nil {
		print("err")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "注册成功",
		"success": true,
	})
}

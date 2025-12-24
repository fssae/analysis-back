package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSettings 获取教师设置
func (h *TeacherHandler) GetSettings(c *gin.Context) {
	//claims, exists := c.Get("claims")
	//if !exists {
	//	c.JSON(http.StatusUnauthorized, gin.H{
	//		"code": 401,
	//		"msg":  "未授权",
	//	})
	//	return
	//}
	//
	//teacherClaims := claims.(domain.TeacherClaims)
	//teacherId, err := primitive.ObjectIDFromHex(teacherClaims.TeacherId)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"code": 400,
	//		"msg":  "教师ID格式错误",
	//	})
	//	return
	//}
	//
	//// 获取教师信息
	//teacher, err := h.teacherService.GetTeacherById(c.Request.Context(), teacherId)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"code": 500,
	//		"msg":  "获取设置失败",
	//	})
	//	return
	//}
	//
	//settings := gin.H{
	//	"email":        teacher.Email,
	//	"notification": true, // 默认开启通知
	//}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "setting",
	})
}

// UpdateSettings 更新教师设置
func (h *TeacherHandler) UpdateSettings(c *gin.Context) {
	//claims, exists := c.Get("claims")
	//if !exists {
	//	c.JSON(http.StatusUnauthorized, gin.H{
	//		"code": 401,
	//		"msg":  "未授权",
	//	})
	//	return
	//}
	//
	//teacherClaims := claims.(domain.TeacherClaims)
	//teacherId, err := primitive.ObjectIDFromHex(teacherClaims.TeacherId)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"code": 400,
	//		"msg":  "教师ID格式错误",
	//	})
	//	return
	//}
	//
	//var settings struct {
	//	Email        string `json:"email"`
	//	Notification bool   `json:"notification"`
	//}
	//
	//if err := c.ShouldBindJSON(&settings); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"code": 400,
	//		"msg":  "请求参数错误",
	//	})
	//	return
	//}
	//
	//// 更新教师设置
	//err = h.teacherService.UpdateSettings(c.Request.Context(), teacherId, settings.Email)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"code": 500,
	//		"msg":  "更新设置失败",
	//	})
	//	return
	//}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

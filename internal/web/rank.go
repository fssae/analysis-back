package web

import (
	"classroom-analysis/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *TeacherHandler) GetRank(c *gin.Context) {
	var req domain.RankRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}
	if req.Page == 0 || req.PageSize == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}
	resp, err := h.teacherService.GetRank(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"msg":   "服务器错误",
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "获取排行榜成功",
		"data":    resp,
		"success": true,
	})
}

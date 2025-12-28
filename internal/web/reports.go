package web

import (
	"classroom-analysis/internal/domain"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetReport 获取分析报告
func (h *TeacherHandler) GetReport(c *gin.Context) {
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

	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少分析ID参数",
		})
		return
	}

	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "分析ID格式错误",
		})
		return
	}

	// 这里应该生成报告文件，暂时返回模拟数据
	reportUrl := fmt.Sprintf("https://server.com/reports/%s.pdf", analysisId.Hex())

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"url": reportUrl,
		},
	})
}

// GetReportList 获取报告列表
func (h *TeacherHandler) GetReportList(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	teacherClaims := claims.(domain.TeacherClaims)
	_, err := primitive.ObjectIDFromHex(teacherClaims.TeacherId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "教师ID格式错误",
		})
		return
	}

	// 获取分析记录作为报告列表
	analyses, err := h.analysisService.GetHistory(c.Request.Context(), "video")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取报告列表失败",
		})
		return
	}

	var reportList []map[string]interface{}
	for _, analysis := range analyses {
		reportList = append(reportList, map[string]interface{}{
			"id":    analysis.Id.Hex(),
			"title": analysis.CourseName,
			"date":  analysis.Timestamp.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": reportList,
	})
}

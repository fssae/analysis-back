package web

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (h *TeacherHandler) Analyze(c *gin.Context) {
	var req domain.TeacherAnalysisRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}
	if req.Url == "" || req.ImageId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "ID或URL不能为空",
		})
		return
	}
	status, _ := h.redis.Get(c, req.ImageId).Result()
	log.Printf("%v", status)
	if status == "STATUS_PROCESSING" {
		//计数器
		domain.IdempotentInterceptTotal.Inc()
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  fmt.Sprintf("正在处理中,请勿重复提交,%+v", req.ImageId),
		})
		return
	}
	// 从JWT claims中获取教师ID
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未找到用户认证信息",
		})
		return
	}

	teacherClaims, ok := claims.(domain.TeacherClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户认证信息格式错误",
		})
		return
	}

	// 异步分析任务，传递教师ID

	go h.performAnalysisAsync(context.Background(), req, teacherClaims.Id)

	c.JSON(http.StatusAccepted, gin.H{
		"code": 202,
		"msg":  "分析任务已提交，正在后台处理中",
		"data": gin.H{
			"image":   req.ImageId,
			"status":  "processing",
			"message": "可通过WebSocket实时获取进度，或使用taskId轮询分析",
			"wsUrl":   "/teacher/ws?taskId=" + req.ImageId,
		},
	})
}
func (h *TeacherHandler) UpdateConfig(c *gin.Context) {
	var req domain.UpdateConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}
	fmt.Print("UpdateConfig", req)
	fmt.Print("id:", req.AnalysisId)
	if err := h.analysisService.UpdateConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    500,
			"message": fmt.Sprintf("服务器错误: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})

}

// GetImageHistory 获取图片分析历史
func (h *TeacherHandler) GetImageHistory(c *gin.Context) {
	history, err := h.analysisService.GetHistory(c.Request.Context(), "image")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取历史记录失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": history,
	})
}
func (h *TeacherHandler) GetVideoHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "history",
	})
}

// GetImageAnalysisResult 获取图片分析结果（汇总）
func (h *TeacherHandler) GetImageAnalysisResult(c *gin.Context) {
	analysisIdStr := c.Query("analysisId")
	if analysisIdStr == "" {
		c.JSON(400, gin.H{"success": false, "msg": "缺少分析ID"})
		return
	}
	analysisId, err := primitive.ObjectIDFromHex(analysisIdStr)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "分析ID格式错误"})
		return
	}
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisId)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": result})
}
func (h *TeacherHandler) GetImageAnalysisDetail(c *gin.Context) {
	id := c.Query("id")
	result, err := h.teacherService.GetImageAnalysisDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": result})
}

// GetVideoAnalysisDetail 获取视频分析详情
func (h *TeacherHandler) GetVideoAnalysisDetail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少视频分析ID",
		})
		return
	}

	result, err := h.teacherService.GetVideoAnalysisDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": result,
	})
}

// GetClassAnalysis 获取班级分析列表
func (h *TeacherHandler) GetClassAnalysis(c *gin.Context) {
	// 解析查询参数
	courseName := c.Query("courseName")
	className := c.Query("className")
	if className != "" {
		className += "班"
	}
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	// 转换分页参数
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	list, total, err := h.teacherService.GetClassAnalysisList(
		c.Request.Context(),
		courseName,
		className,
		startDate,
		endDate,
		pageInt,
		pageSizeInt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":  list,
			"total": total,
		},
	})
}

// performAnalysisAsync 异步执行分析任务，不返回HTTP响应
func (h *TeacherHandler) performAnalysisAsync(c context.Context, req domain.TeacherAnalysisRequest, teacherId primitive.ObjectID) {
	// 使用结构化日志
	h.logger.Info("开始异步分析",
		zap.String("analysisType", req.AnalysisType),
		zap.String("imageId", req.ImageId),
		zap.String("url", req.Url),
		zap.String("className", req.ClassName),
		zap.String("courseName", req.CourseName))

	// 获取WebSocket管理器
	wsManager := GetAnalysisWSManager()
	message := domain.KafkaMessage{
		Req: req,
	}
	taskId := req.ImageId
	teacherIdStr := teacherId.Hex()
	// 检查kafkaWriter是否为nil
	if h.kafkaWriter == nil {
		err := fmt.Errorf("kafkaWriter未初始化")
		h.logger.Error("kafkaWriter为nil", zap.Error(err))
		h.updateTaskStatus(taskId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(teacherIdStr, "error", "分析失败", "nil", err.Error())
		return
	}

	err := h.kafkaWriter.Write(message)
	if err != nil {
		h.logger.Error("发送Kafka消息失败",
			zap.String("imageId", taskId),
			zap.Error(err),
			zap.Any("message", message))
		h.updateTaskStatus(taskId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(teacherIdStr, "error", "分析失败", "nil", err.Error())
		return
	}
	wsManager.SendTaskStatusUpdate(teacherIdStr, "已传入Kafka消息", "异步分析中...", "nil", "nil")
	// 读取结果，传入imageId用于消息匹配
	resultChan, ctx, cancel, err := h.ReadKafka(c, req.ImageId)
	if err != nil {
		fmt.Printf("收到空响应: %v\n", err)
		h.updateTaskStatus(taskId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(teacherIdStr, "创建Kafka消费者失败", "分析失败", "nil", "nil")
		return
	}
	//wsManager.SendTaskStatusUpdate(taskId, "test", "test", "nil", "test")
	defer cancel()
	select {
	case resp := <-resultChan:
		if resp == nil {
			h.updateTaskStatus(taskId, "error", "", "收到空响应", teacherId, req.ConfidenceThreshold)
			wsManager.SendTaskStatusUpdate(teacherIdStr, "收到空响应", "分析失败", "nil", "响应为空")
			return
		}
		//分析完成，幂等处理
		h.redis.Set(c, req.ImageId, "STATUS_DONE", 20*time.Minute)
		resultURL := ""
		if resp.ResultURL != "" {
			resultURL = resp.ResultURL
		}
		//上传status，email,计算平均专注度
		h.updateTaskStatus(taskId, "completed", resultURL, "", teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(teacherIdStr, "completed", "分析成功", resultURL, "")
	case <-ctx.Done():
		h.updateTaskStatus(taskId, "timeout", "", ctx.Err().Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(teacherIdStr, "超时", "分析失败", "nil", ctx.Err().Error())
	}
}

// updateTaskStatus 更新任务状态到数据库
func (h *TeacherHandler) updateTaskStatus(imageId, status, resultUrl, errorMsg string, teacherId primitive.ObjectID, confidenceThreshold float64) {
	h.logger.Info("更新任务状态",
		zap.String("imageId", imageId),
		zap.String("status", status),
		zap.String("resultUrl", resultUrl),
		zap.String("errorMsg", errorMsg),
		zap.Float64("confidenceThreshold", confidenceThreshold))

	// 检查analysisService是否为nil
	if h.analysisService == nil {
		h.logger.Error("analysisService为nil，无法更新任务状态",
			zap.String("taskId", imageId),
			zap.String("status", status))
		return
	}

	updateStatus := domain.UpdateStatus{
		TeacherId:           teacherId, // 使用传入的教师ID
		TaskId:              imageId,
		ConfidenceThreshold: confidenceThreshold,
		Status:              status,
		ResultUrl:           resultUrl,
		ErrorMsg:            errorMsg,
	}

	if err := h.analysisService.UpdateStatus(&updateStatus); err != nil {
		h.logger.Error("更新任务状态失败",
			zap.String("taskId", imageId),
			zap.String("status", status),
			zap.Error(err))
		// 这里可以考虑重试机制或者将失败的状态更新放入队列
		return
	}
	//插入email
	h.logger.Info("任务状态更新成功",
		zap.String("taskId", imageId),
		zap.String("status", status))
}

func (h *TeacherHandler) GetStatus(c *gin.Context) {
	taskId := c.Query("taskId")

	// 检查analysisService是否为nil
	if h.analysisService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "服务不可用",
			"error": "analysisService未初始化",
		})
		return
	}

	status, err := h.analysisService.GetStatus(c, taskId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":   "获取失败",
			"error": err,
		})
		return
	}
	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "status状态为空",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":    "获取成功",
		"status": status,
	})
}

// HandleAnalysisWebSocket 处理分析任务的WebSocket连接
func (h *TeacherHandler) HandleAnalysisWebSocket(c *gin.Context) {
	wsManager := GetAnalysisWSManager()
	claims, err := util.GetClaims(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "id 解析错误",
		})
		return
	}
	//通过token获取教师id
	wsManager.HandleConnection(claims.TeacherId, c.Writer, c.Request)
}

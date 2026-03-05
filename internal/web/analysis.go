package web

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
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

	fileId := ""
	if req.AnalysisType == "video" {
		fileId = req.VideoId
		log.Printf("videoId:%v", req.VideoId)
	} else {
		fileId = req.ImageId
		log.Printf("imageId:%v", req.ImageId)
	}

	if req.Url == "" || fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "ID或URL不能为空",
		})
		return
	}
	status, _ := h.redis.Get(c, fileId).Result()
	log.Printf("%v", status)
	if status == "STATUS_PROCESSING" {
		domain.IdempotentInterceptTotal.Inc()
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  fmt.Sprintf("正在处理中,请勿重复提交,%+v", fileId),
		})
		return
	}
	claims, err := util.GetClaims(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  fmt.Sprintf("请重新登录"),
		})
		return
	}

	if h.kafkaWriter == nil {
		h.logger.Warn("KafkaWriter未初始化，使用Mock模式分析", zap.String("fileId", fileId))
		go h.analysisService.MockAnalyzeImage(
			context.Background(),
			fileId,
			claims.TeacherId,
			req.ConfidenceThreshold,
		)

		c.JSON(http.StatusAccepted, gin.H{
			"code": 202,
			"msg":  "分析任务已提交(Mock模式)",
			"data": gin.H{
				"fileId":  fileId,
				"status":  "processing",
				"message": "正在使用模拟分析服务(Dev Mode)",
				"wsUrl":   "/teacher/ws?taskId=" + fileId,
			},
		})
		return
	}

	go h.performAnalysisAsync(context.Background(), req, claims)

	c.JSON(http.StatusAccepted, gin.H{
		"code": 202,
		"msg":  "分析任务已提交，正在后台处理中",
		"data": gin.H{
			"fileId":  fileId,
			"status":  "processing",
			"message": "可通过WebSocket实时获取进度，或使用taskId轮询分析",
			"wsUrl":   "/teacher/ws?taskId=" + fileId,
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

// UpdateAnalysisName 更新分析文件名
func (h *TeacherHandler) UpdateAnalysisName(c *gin.Context) {
	var req domain.UpdateAnalysisNameRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	if req.ImageId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "imageid 不能为空",
		})
		return
	}

	if err := h.analysisService.UpdateAnalysisName(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    500,
			"message": fmt.Sprintf("服务器错误: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "文件名更新成功",
	})
}

// GetImageHistory 获取图片分析历史
func (h *TeacherHandler) GetImageHistory(c *gin.Context) {
	// 支持分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	history, total, err := h.analysisService.GetHistory(c.Request.Context(), "image", pageInt, pageSizeInt)
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
		"data": gin.H{
			"list":  history,
			"total": total,
		},
	})
}
func (h *TeacherHandler) GetHistory(c *gin.Context) {
	// 通过 type 参数区分图片和视频历史
	// type 为空字符串时返回全部记录
	analysisType := c.Query("type")

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	history, total, err := h.analysisService.GetHistory(c.Request.Context(), analysisType, pageInt, pageSizeInt)
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
		"data": gin.H{
			"list":  history,
			"total": total,
		},
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

	if result == nil {
		c.JSON(404, gin.H{"success": false, "msg": "分析记录不存在"})
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
func (h *TeacherHandler) performAnalysisAsync(c context.Context, req domain.TeacherAnalysisRequest, claims *domain.TeacherClaims) {
	imageId := ""
	if req.AnalysisType == "video" {
		imageId = req.VideoId
		req.ImageId = req.VideoId
	} else {
		imageId = req.ImageId
	}

	h.logger.Info("开始异步分析",
		zap.String("analysisType", req.AnalysisType),
		zap.String("imageId", imageId),
		zap.String("url", req.Url),
		zap.String("className", req.ClassName),
		zap.String("courseName", req.CourseName))
	wsManager := GetAnalysisWSManager()
	teacherId := claims.Id
	// 使用 imageId 作为 WebSocket 任务标识，与前端保持一致
	wsId := imageId

	if h.kafkaWriter == nil {
		err := fmt.Errorf("kafkaWriter未初始化")
		h.logger.Error("kafkaWriter为nil", zap.Error(err))
		h.updateTaskStatus(imageId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(wsId, "error", "分析失败", "nil", err.Error())
		return
	}

	message := domain.KafkaMessage{
		Req: req,
	}
	err := h.kafkaWriter.Write(message)
	if err != nil {
		h.logger.Error("发送Kafka消息失败",
			zap.String("imageId", imageId),
			zap.Error(err),
			zap.Any("message", message))
		h.updateTaskStatus(imageId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(wsId, "error", "分析失败", "nil", err.Error())
		return
	}
	wsManager.SendTaskStatusUpdate(wsId, "已传入Kafka消息", "异步分析中...", "nil", "nil")
	resultChan, ctx, cancel, err := h.ReadKafka(c, imageId)
	if err != nil {
		fmt.Printf("收到空响应: %v\n", err)
		h.updateTaskStatus(imageId, "error", "", err.Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(wsId, "创建Kafka消费者失败", "分析失败", "nil", "nil")
		return
	}
	defer cancel()
	select {
	case resp := <-resultChan:
		if resp == nil {
			h.updateTaskStatus(imageId, "error", "", "收到空响应", teacherId, req.ConfidenceThreshold)
			wsManager.SendTaskStatusUpdate(wsId, "收到空响应", "分析失败", "nil", "响应为空")
			return
		}
		h.logger.Info("收到Kafka响应",
			zap.String("taskId", resp.ID),
			zap.String("respID", resp.ID),
			zap.String("resultURL", resp.ResultURL))

		h.redis.Set(c, imageId, "STATUS_DONE", 20*time.Minute)
		resultURL := ""
		if resp.ResultURL != "" {
			resultURL = resp.ResultURL
		}

		if req.FileName != "" {
			err = h.analysisService.UpdateFileNameByTaskId(c, resp.ID, req.FileName)
			if err != nil {
				h.logger.Error("更新分析文件名失败",
					zap.String("taskId", resp.ID),
					zap.String("fileName", req.FileName),
					zap.Error(err))
			} else {
				h.logger.Info("成功更新分析文件名",
					zap.String("taskId", resp.ID),
					zap.String("fileName", req.FileName))
			}
		}

		h.updateTaskStatus(imageId, "completed", resultURL, "", teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(wsId, "completed", "分析成功", resultURL, "")
	case <-ctx.Done():
		h.updateTaskStatus(imageId, "timeout", "", ctx.Err().Error(), teacherId, req.ConfidenceThreshold)
		wsManager.SendTaskStatusUpdate(wsId, "超时", "分析失败", "nil", ctx.Err().Error())
	}
}

func (h *TeacherHandler) updateTaskStatus(imageId, status, resultUrl, errorMsg string, teacherId primitive.ObjectID, confidenceThreshold float64) {
	h.logger.Info("更新任务状态",
		zap.String("imageId", imageId),
		zap.String("status", status),
		zap.String("resultUrl", resultUrl),
		zap.String("errorMsg", errorMsg),
		zap.Float64("confidenceThreshold", confidenceThreshold))

	if h.analysisService == nil {
		h.logger.Error("analysisService为nil，无法更新任务状态",
			zap.String("imageId", imageId),
			zap.String("status", status))
		return
	}

	oid, err := primitive.ObjectIDFromHex(imageId)
	if err != nil {
		h.logger.Warn("imageId不是有效的ObjectID，生成新的ObjectID",
			zap.String("imageId", imageId),
			zap.Error(err))
		oid = primitive.NewObjectID()
	}

	updateStatus := domain.UpdateStatus{
		TaskId:              imageId,
		ImageId:             oid,
		TeacherId:           teacherId,
		ConfidenceThreshold: confidenceThreshold,
		Status:              status,
		ResultUrl:           resultUrl,
		ErrorMsg:            errorMsg,
	}

	if err := h.analysisService.UpdateStatus(&updateStatus); err != nil {
		h.logger.Error("更新任务状态失败",
			zap.String("imageId", imageId),
			zap.String("status", status),
			zap.Error(err))
		return
	}
	h.logger.Info("任务状态更新成功",
		zap.String("imageId", imageId),
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

	// 从查询参数中获取 imageId（前端使用 imageId 作为任务标识）
	imageId := c.Query("imageId")
	if imageId == "" {
		// 兼容旧版使用 taskId 的情况
		imageId = c.Query("taskId")
	}

	if imageId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少 imageId 或 taskId 参数",
		})
		return
	}

	// 使用 imageId 作为任务标识注册 WebSocket 连接
	wsManager.HandleConnection(imageId, c.Writer, c.Request)
}

// GetGlobalAnalysis 获取全局分析数据
func (h *TeacherHandler) GetGlobalAnalysis(c *gin.Context) {
	data, err := h.analysisService.GetGlobalAnalysisData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取全局分析数据失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	})
}

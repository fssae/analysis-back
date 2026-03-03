package web

import (
	"classroom-analysis/internal/repository"
	"classroom-analysis/internal/service"
	"classroom-analysis/internal/web/middleware"

	"github.com/redis/go-redis/v9"

	"github.com/IBM/sarama"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type KafkaWriter interface {
	Write(message interface{}) error
}
type KafkaReader interface {
	Read(partition int32, offset int64) (<-chan *sarama.ConsumerMessage, error)
}

type ConsumerGroupFactory func(groupId string) (sarama.ConsumerGroup, error)

// TeacherHandler 教师处理器
type TeacherHandler struct {
	teacherService  *service.TeacherService
	analysisService *service.AnalysisService
	fileRepo        *repository.FileRepository
	kafkaWriter     KafkaWriter
	kafkaReader     KafkaReader
	logger          *zap.Logger
	redis           *redis.Client
}

// NewTeacherHandler 创建教师处理器实例
func NewTeacherHandler(
	teacherService *service.TeacherService,
	analysisService *service.AnalysisService,
	fileRepo *repository.FileRepository,
	logger *zap.Logger,
	kafkaWriter KafkaWriter,
	redis *redis.Client,
) *TeacherHandler {
	return &TeacherHandler{
		teacherService:  teacherService,
		analysisService: analysisService,
		fileRepo:        fileRepo,
		logger:          logger,
		kafkaWriter:     kafkaWriter,
		redis:           redis,
	}
}

// RegisterRoutes 注册教师相关路由
func (h *TeacherHandler) RegisterRoutes(server *gin.Engine) {
	// 教师登录  注册 - 不需要JWT验证
	server.POST("/teacher/login", h.Login)
	server.POST("/teacher/register", h.Register)

	// 需要JWT验证的接口
	teacherGroup := server.Group("/teacher")
	teacherGroup.Use(
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/teacher/login").
			IgnorePaths("/teacher/register").
			//ReWritPaths("/teacher/login").
			Build(),
	)
	// 排行
	teacherGroup.GET("/class-rank", h.GetRank)
	// 仪表盘
	teacherGroup.GET("/dashboard", h.GetDashboard)
	// Kafka状态检查 - 暂时注释掉
	// teacherGroup.GET("/kafka/status", h.CheckKafkaStatus)
	//视频或图片分析
	teacherGroup.POST("/analyze", h.Analyze)
	// 分析任务WebSocket连接
	teacherGroup.GET("/ws", h.HandleAnalysisWebSocket)
	teacherGroup.GET("/status", h.GetStatus)

	teacherGroup.GET("/video/analysis-detail", h.GetVideoAnalysisDetail)
	teacherGroup.GET("/video/history", h.GetVideoHistory)
	teacherGroup.POST("/image/analysis-config", h.UpdateConfig)
	teacherGroup.POST("/update-analysis-name", h.UpdateAnalysisName)
	teacherGroup.GET("/image/history", h.GetImageHistory)
	teacherGroup.GET("/image/analysis-result", h.GetImageAnalysisResult)
	teacherGroup.GET("/image/analysis-detail", h.GetImageAnalysisDetail)
	// 报告
	teacherGroup.GET("/report", h.GetReport)
	teacherGroup.GET("/report/list", h.GetReportList)

	// 学生专注度
	teacherGroup.GET("/students", h.GetStudents)

	//根据学生id,姓名,平局专注度范围，分析时间进行查询
	teacherGroup.GET("/findStudents", h.FindStudents)

	// 新增：班级分析接口
	teacherGroup.GET("/class-analysis", h.GetClassAnalysis)

	// 情绪分析接口
	teacherGroup.GET("/emotion-analysis", h.GetEmotionAnalysis)
	teacherGroup.GET("/emotion-heatmap", h.GetEmotionHeatmap)

	// 疲劳度分析接口
	teacherGroup.GET("/fatigue-analysis", h.GetFatigueAnalysis)
	teacherGroup.GET("/blink-analysis", h.GetBlinkAnalysis)

	// 教师设置
	teacherGroup.GET("/settings", h.GetSettings)
	teacherGroup.POST("/settings", h.UpdateSettings)
}

package ioc

import (
	"classroom-analysis/internal/repository"
	"classroom-analysis/internal/repository/dao"
	"classroom-analysis/internal/service"
	"classroom-analysis/internal/web"

	"github.com/google/wire"
)

var MinimalSet = wire.NewSet(
	InitMongodb,
	InitMongoDatabase,
	InitRedis,
	InitLogger,
	NewLogger,
	NewMinioClient,

	// DAO层
	dao.NewTeacherDAO,
	dao.NewAnalysisDAO,
	dao.NewAnalysisTaskDAO,
	dao.NewStudentDAO,
	dao.NewStudentFocusDAO,
	dao.NewVideoAnalysisDAO,
	dao.NewImageAnalysisDAO,
	dao.NewFaceAnalysisDAO,
	dao.NewFileDao,
	dao.ProvideFileDaoInterface,
	dao.NewEmailDao,

	// Repository层
	repository.NewTeacherRepository,
	repository.NewAnalysisRepository,
	repository.NewAnalysisTaskRepository,
	repository.NewStudentRepository,
	repository.NewStudentFocusRepository,
	repository.NewVideoAnalysisRepository,
	repository.NewImageAnalysisRepository,
	repository.NewFaceAnalysisRepository,
	repository.NewFileRepository,
	repository.ProvideFileInterface,
	repository.NewEmailRepository,

	// Service层
	service.NewTeacherService,
	service.NewAnalysisService,
	service.NewFileService,
	service.NewEmailService,

	// Kafka相关
	InitKafkaWriter,
	wire.Bind(new(web.KafkaWriter), new(*Writer)),
	// Web层 - 添加FileHandler
	web.NewTeacherHandler,
	web.NewFileHandler,
	web.NewEmailHandler,
	web.NewFaceApiHandler,

	// Gin引擎
	InitGin,
)

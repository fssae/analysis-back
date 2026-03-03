package service

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository"
	"context"
	"fmt"
	"math/rand"
	"time"

	// "mime/multipart" // 已移除
	// "github.com/google/uuid" // 已移除
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalysisService struct {
	analysisRepo      *repository.AnalysisRepository
	videoAnalysisRepo *repository.VideoAnalysisRepository
	imageAnalysisRepo *repository.ImageAnalysisRepository
	studentFocusRepo  *repository.StudentFocusRepository
	studentRepo       *repository.StudentRepository
	analysisTaskRepo  *repository.AnalysisTaskRepository
	faceAnalysisRepo  *repository.FaceAnalysisRepository
	fileRepo          *repository.FileRepository
}

func NewAnalysisService(
	analysisRepo *repository.AnalysisRepository,
	videoAnalysisRepo *repository.VideoAnalysisRepository,
	imageAnalysisRepo *repository.ImageAnalysisRepository,
	studentFocusRepo *repository.StudentFocusRepository,
	studentRepo *repository.StudentRepository,
	analysisTaskRepo *repository.AnalysisTaskRepository,
	faceAnalysisRepo *repository.FaceAnalysisRepository,
	fileRepo *repository.FileRepository,
) *AnalysisService {
	return &AnalysisService{
		analysisRepo:      analysisRepo,
		videoAnalysisRepo: videoAnalysisRepo,
		imageAnalysisRepo: imageAnalysisRepo,
		studentFocusRepo:  studentFocusRepo,
		studentRepo:       studentRepo,
		analysisTaskRepo:  analysisTaskRepo,
		faceAnalysisRepo:  faceAnalysisRepo,
		fileRepo:          fileRepo,
	}
}

func (s *AnalysisService) UpdateConfig(update *domain.UpdateConfigRequest) error {
	return s.analysisRepo.UpdateConfig(update)
}

func (s *AnalysisService) UpdateAnalysisName(update *domain.UpdateAnalysisNameRequest) error {
	return s.analysisRepo.UpdateAnalysisName(update)
}

func (s *AnalysisService) UpdateStatus(update *domain.UpdateStatus) error {
	return s.analysisRepo.UpdateStatus(update)
}

func (s *AnalysisService) GetStatus(ctx context.Context, taskId string) ([]*domain.UpdateStatus, error) {
	return s.analysisTaskRepo.GetStatus(ctx, taskId)
}

func (s *AnalysisService) FindByTaskId(ctx context.Context, taskId string) (*domain.Analysis, error) {
	return s.analysisRepo.FindByTaskId(ctx, taskId)
}

func (s *AnalysisService) FindByImageIdString(ctx context.Context, imageIdStr string) (*domain.Analysis, error) {
	return s.analysisRepo.FindByImageIdString(ctx, imageIdStr)
}

func (s *AnalysisService) UpdateFileNameByTaskId(ctx context.Context, taskId, fileName string) error {
	return s.analysisRepo.UpdateFileNameByTaskId(ctx, taskId, fileName)
}

// AnalyzeVideo 分析视频 (Legacy/Mock)
func (s *AnalysisService) AnalyzeVideo(ctx context.Context, fileName, fileType string) (*domain.Analysis, error) {
	analysis := &domain.Analysis{
		FileName:  fileName,
		FileType:  fileType, // "video"
		Timestamp: time.Now(),
	}
	err := s.analysisRepo.Create(ctx, analysis)
	if err != nil {
		return nil, err
	}
	go s.processVideoAnalysis(ctx, analysis)
	return analysis, nil
}

func (s *AnalysisService) processVideoAnalysis(ctx context.Context, analysis *domain.Analysis) {
	time.Sleep(5 * time.Second)
	var focusTrend []domain.FocusPoint
	for i := 0; i < 10; i++ {
		focusTrend = append(focusTrend, domain.FocusPoint{
			Time:  fmt.Sprintf("%02d:%02d", i/60, i%60),
			Focus: 70 + rand.Float64()*20,
		})
	}
	// 生成人脸分析数据
	var faces []domain.FaceAnalysis
	for i := 0; i < 5; i++ {
		faces = append(faces, domain.FaceAnalysis{
			FaceIndex:  i + 1,
			FocusScore: 70 + rand.Float64()*20,
			Confidence: 0.9 + rand.Float64()*0.1,
		})
	}
	analysis.Faces = faces
	s.analysisRepo.Update(ctx, analysis)
}

// MockAnalyzeImage 模拟图片分析（当 Kafka 不可用时）
func (s *AnalysisService) MockAnalyzeImage(ctx context.Context, imageId string, teacherIdStr string, threshold float64) {
	tid, _ := primitive.ObjectIDFromHex(teacherIdStr)

	// 1. 设置状态为处理中
	imgObjID, _ := primitive.ObjectIDFromHex(imageId)
	if imgObjID.IsZero() {
		imgObjID = primitive.NewObjectID()
	}
	s.UpdateStatus(&domain.UpdateStatus{
		TaskId:              imageId,
		ImageId:             imgObjID,
		Status:              "processing",
		TeacherId:           tid,
		ConfidenceThreshold: threshold,
	})

	// 异步模拟分析
	go func() {
		// 模拟耗时
		time.Sleep(3 * time.Second)

		// 构造结果
		resultUrl := "https://picsum.photos/800/600" // 模拟结果图

		// 尝试将 imageId 转为 ObjectID，如果失败则生成新的（兼容性）
		imgObjID, err := primitive.ObjectIDFromHex(imageId)
		if err != nil {
			imgObjID = primitive.NewObjectID()
		}

		analysis := &domain.Analysis{
			Id:           primitive.NewObjectID(),
			ImageId:      imgObjID,
			Timestamp:    time.Now(),
			AnalysisMode: "image",
			ResultUrl:    resultUrl,
			Faces: []domain.FaceAnalysis{
				{FaceIndex: 1, FocusScore: 0.92, Confidence: 0.98},
				{FaceIndex: 2, FocusScore: 0.85, Confidence: 0.96},
				{FaceIndex: 3, FocusScore: 0.76, Confidence: 0.91},
			},
			ClassName:  "Mock Class",
			CourseName: "Mock Course",
		}

		// 保存结果到 DB
		// 注意 context.Background() 用于新的 goroutine
		_ = s.analysisRepo.Create(context.Background(), analysis)

		// 更新状态为完成
		s.UpdateStatus(&domain.UpdateStatus{
			TaskId:    imageId,
			ImageId:   imgObjID,
			Status:    "completed",
			ResultUrl: resultUrl,
			TeacherId: tid,
		})
	}()
}

func (s *AnalysisService) AnalyzeImage(ctx context.Context, fileName, fileType string) (*domain.Analysis, error) {
	analysis := &domain.Analysis{
		FileName:  fileName,
		FileType:  fileType, // "image"
		Timestamp: time.Now(),
	}
	err := s.analysisRepo.Create(ctx, analysis)
	if err != nil {
		return nil, err
	}
	go s.processImageAnalysis(ctx, analysis)
	return analysis, nil
}

func (s *AnalysisService) processImageAnalysis(ctx context.Context, analysis *domain.Analysis) {
	time.Sleep(2 * time.Second)
	// 生成人脸分析数据
	var faces []domain.FaceAnalysis
	for i := 0; i < 5; i++ {
		faces = append(faces, domain.FaceAnalysis{
			FaceIndex:  i + 1,
			FocusScore: 75 + rand.Float64()*20,
			Confidence: 0.9 + rand.Float64()*0.1,
		})
	}
	analysis.Faces = faces
	s.analysisRepo.Update(ctx, analysis)
}

// GetHistory 获取分析历史
func (s *AnalysisService) GetHistory(ctx context.Context, fileType string) ([]*domain.Analysis, error) {
	// teacherId 暂时传空，后续应从 Context 获取或作为参数传入
	analyses, err := s.analysisRepo.FindByTeacherId(ctx, primitive.NilObjectID, 20)
	if err != nil {
		return nil, err
	}
	var result []*domain.Analysis
	for _, analysis := range analyses {
		if analysis.AnalysisMode == fileType || (fileType == "image" && analysis.FileType == "image") {
			result = append(result, analysis)
		}
	}
	return result, nil
}

// GetAnalysisResult 获取单条分析结果
func (s *AnalysisService) GetAnalysisResult(ctx context.Context, analysisId primitive.ObjectID) (*domain.Analysis, error) {
	return s.analysisRepo.FindById(ctx, analysisId)
}

// 批量图片分析任务、状态、结果等接口建议继续用 AnalysisTask/FaceAnalysis 相关结构体，不建议再用 Analysis 结构体存储任务型数据。

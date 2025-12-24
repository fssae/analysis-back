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
}

func NewAnalysisService(
	analysisRepo *repository.AnalysisRepository,
	videoAnalysisRepo *repository.VideoAnalysisRepository,
	imageAnalysisRepo *repository.ImageAnalysisRepository,
	studentFocusRepo *repository.StudentFocusRepository,
	studentRepo *repository.StudentRepository,
	analysisTaskRepo *repository.AnalysisTaskRepository,
	faceAnalysisRepo *repository.FaceAnalysisRepository,
) *AnalysisService {
	return &AnalysisService{
		analysisRepo:      analysisRepo,
		videoAnalysisRepo: videoAnalysisRepo,
		imageAnalysisRepo: imageAnalysisRepo,
		studentFocusRepo:  studentFocusRepo,
		studentRepo:       studentRepo,
		analysisTaskRepo:  analysisTaskRepo,
		faceAnalysisRepo:  faceAnalysisRepo,
	}
}

func (s *AnalysisService) UpdateConfig(update *domain.UpdateConfigRequest) error {
	return s.analysisRepo.UpdateConfig(update)
}
func (s *AnalysisService) UpdateStatus(update *domain.UpdateStatus) error {
	return s.analysisRepo.UpdateStatus(update)
}
func (s *AnalysisService) GetStatus(ctx context.Context, taskId string) ([]*domain.UpdateStatus, error) {
	return s.analysisTaskRepo.GetStatus(ctx, taskId)
}

// AnalyzeVideo 分析视频
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

// 获取分析历史（只返回基本信息）
func (s *AnalysisService) GetHistory(ctx context.Context, fileType string) ([]map[string]interface{}, error) {
	// 这里只能用 fileType 区分图片/视频
	// teacherId 相关已去除
	analyses, err := s.analysisRepo.FindByTeacherId(ctx, primitive.NilObjectID, 20) // teacherId 传空
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, analysis := range analyses {
		if analysis.AnalysisMode == fileType {
			result = append(result, map[string]interface{}{
				"id":        analysis.Id.Hex(),
				"fileName":  analysis.CourseName,
				"fileType":  analysis.ClassName,
				"timestamp": analysis.Timestamp.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return result, nil
}

// 获取单条分析结果（只返回人脸分析详情）
func (s *AnalysisService) GetAnalysisResult(ctx context.Context, analysisId primitive.ObjectID) (map[string]interface{}, error) {
	analysis, err := s.analysisRepo.FindById(ctx, analysisId)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":           analysis.Id.Hex(),
		"fileName":     analysis.FileName,
		"fileType":     analysis.FileType,
		"timestamp":    analysis.Timestamp.Format("2006-01-02 15:04:05"),
		"faceAnalysis": analysis.Faces,
	}, nil
}

// 批量图片分析任务、状态、结果等接口建议继续用 AnalysisTask/FaceAnalysis 相关结构体，不建议再用 Analysis 结构体存储任务型数据。

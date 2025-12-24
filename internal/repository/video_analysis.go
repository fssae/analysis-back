package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VideoAnalysisRepository struct {
	videoAnalysisDAO *dao.VideoAnalysisDAO
}

func NewVideoAnalysisRepository(videoAnalysisDAO *dao.VideoAnalysisDAO) *VideoAnalysisRepository {
	return &VideoAnalysisRepository{
		videoAnalysisDAO: videoAnalysisDAO,
	}
}

// Create 创建视频分析
func (r *VideoAnalysisRepository) Create(ctx context.Context, videoAnalysis *domain.VideoAnalysis) error {
	return r.videoAnalysisDAO.Create(ctx, videoAnalysis)
}

// FindByAnalysisId 根据分析ID查找视频分析
func (r *VideoAnalysisRepository) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) (*domain.VideoAnalysis, error) {
	return r.videoAnalysisDAO.FindByAnalysisId(ctx, analysisId)
}

// Update 更新视频分析
func (r *VideoAnalysisRepository) Update(ctx context.Context, videoAnalysis *domain.VideoAnalysis) error {
	return r.videoAnalysisDAO.Update(ctx, videoAnalysis)
}

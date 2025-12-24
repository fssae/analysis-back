package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImageAnalysisRepository struct {
	imageAnalysisDAO *dao.ImageAnalysisDAO
}

func NewImageAnalysisRepository(imageAnalysisDAO *dao.ImageAnalysisDAO) *ImageAnalysisRepository {
	return &ImageAnalysisRepository{
		imageAnalysisDAO: imageAnalysisDAO,
	}
}

// Create 创建图片分析
func (r *ImageAnalysisRepository) Create(ctx context.Context, imageAnalysis *domain.ImageAnalysis) error {
	return r.imageAnalysisDAO.Create(ctx, imageAnalysis)
}

// FindByAnalysisId 根据分析ID查找图片分析
func (r *ImageAnalysisRepository) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) (*domain.ImageAnalysis, error) {
	return r.imageAnalysisDAO.FindByAnalysisId(ctx, analysisId)
}

// Update 更新图片分析
func (r *ImageAnalysisRepository) Update(ctx context.Context, imageAnalysis *domain.ImageAnalysis) error {
	return r.imageAnalysisDAO.Update(ctx, imageAnalysis)
}

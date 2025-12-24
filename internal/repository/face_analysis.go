package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"
)

type FaceAnalysisRepository struct {
	faceAnalysisDAO *dao.FaceAnalysisDAO
}

func NewFaceAnalysisRepository(faceAnalysisDAO *dao.FaceAnalysisDAO) *FaceAnalysisRepository {
	return &FaceAnalysisRepository{
		faceAnalysisDAO: faceAnalysisDAO,
	}
}

// FindByTaskId 根据任务ID查找所有人脸分析记录
func (r *FaceAnalysisRepository) FindByTaskId(ctx context.Context, taskId string) ([]*domain.FaceAnalysis, error) {
	return r.faceAnalysisDAO.FindByTaskId(ctx, taskId)
}

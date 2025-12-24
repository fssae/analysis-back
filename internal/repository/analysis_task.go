package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"
)

type AnalysisTaskRepository struct {
	analysisTaskDAO *dao.AnalysisTaskDAO
}

func NewAnalysisTaskRepository(analysisTaskDAO *dao.AnalysisTaskDAO) *AnalysisTaskRepository {
	return &AnalysisTaskRepository{
		analysisTaskDAO: analysisTaskDAO,
	}
}

// Create 创建分析任务
func (r *AnalysisTaskRepository) Create(ctx context.Context, task *domain.AnalysisTask) error {
	return r.analysisTaskDAO.Create(ctx, task)
}
func (r *AnalysisTaskRepository) GetStatus(ctx context.Context, taskId string) ([]*domain.UpdateStatus, error) {
	return r.analysisTaskDAO.FindByTaskId(ctx, taskId)
}

// Update 更新任务
func (r *AnalysisTaskRepository) Update(ctx context.Context, task *domain.AnalysisTask) error {
	return r.analysisTaskDAO.Update(ctx, task)
}

// UpdateStatus 更新任务状态
func (r *AnalysisTaskRepository) UpdateStatus(ctx context.Context, taskId string, status string, progress int) error {
	return r.analysisTaskDAO.UpdateStatus(ctx, taskId, status, progress)
}

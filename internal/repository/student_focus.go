package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudentFocusRepository struct {
	studentFocusDAO *dao.StudentFocusDAO
}

func NewStudentFocusRepository(studentFocusDAO *dao.StudentFocusDAO) *StudentFocusRepository {
	return &StudentFocusRepository{
		studentFocusDAO: studentFocusDAO,
	}
}

// Create 创建学生专注度记录
func (r *StudentFocusRepository) Create(ctx context.Context, studentFocus *domain.StudentFocus) error {
	return r.studentFocusDAO.Create(ctx, studentFocus)
}

// FindByStudentId 根据学生ID查找专注度记录
func (r *StudentFocusRepository) FindByStudentId(ctx context.Context, studentId primitive.ObjectID, limit int64) ([]*domain.StudentFocus, error) {
	return r.studentFocusDAO.FindByStudentId(ctx, studentId, limit)
}

// FindByAnalysisId 根据分析ID查找专注度记录
func (r *StudentFocusRepository) FindByAnalysisId(ctx context.Context, analysisId primitive.ObjectID) ([]*domain.StudentFocus, error) {
	return r.studentFocusDAO.FindByAnalysisId(ctx, analysisId)
}

// GetAverageFocusByStudentId 获取学生平均专注度
func (r *StudentFocusRepository) GetAverageFocusByStudentId(ctx context.Context, studentId primitive.ObjectID) (float64, error) {
	return r.studentFocusDAO.GetAverageFocusByStudentId(ctx, studentId)
}

// GetFocusTrend 获取专注度趋势数据
func (r *StudentFocusRepository) GetFocusTrend(ctx context.Context) ([]float64, error) {
	return r.studentFocusDAO.GetAverageFocusScores(ctx)
}
func (r *StudentFocusRepository) GetFocusDistribution(ctx context.Context) (map[string]int, error) {
	return r.studentFocusDAO.GetFocusDistribution(ctx)
}

func (r *StudentFocusRepository) GetFocusByStudentId(ctx context.Context, id string) (float64, int, int, error) {
	return r.studentFocusDAO.GetFocusMetricsById(ctx, id)
}

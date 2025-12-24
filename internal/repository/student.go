package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudentRepository struct {
	studentDAO *dao.StudentDAO
}

func NewStudentRepository(studentDAO *dao.StudentDAO) *StudentRepository {
	return &StudentRepository{
		studentDAO: studentDAO,
	}
}

// Create 创建学生
func (r *StudentRepository) Create(ctx context.Context, student *domain.Student) error {
	return r.studentDAO.Create(ctx, student)
}

// FindById 根据ID查找学生
func (r *StudentRepository) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Student, error) {
	return r.studentDAO.FindById(ctx, id)
}

// FindAll 查找所有学生
func (r *StudentRepository) FindAll(ctx context.Context) ([]*domain.Student, error) {
	return r.studentDAO.FindAll(ctx)
}

// FindByName 根据姓名查找学生
func (r *StudentRepository) FindByName(ctx context.Context, name string) ([]*domain.Student, error) {
	return r.studentDAO.FindByName(ctx, name)
}

// Count 统计学生数量
func (r *StudentRepository) Count(ctx context.Context) (int64, error) {
	return r.studentDAO.Count(ctx)
}


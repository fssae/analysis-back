package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type TeacherRepository struct {
	teacherDAO *dao.TeacherDAO
}

func NewTeacherRepository(teacherDAO *dao.TeacherDAO) *TeacherRepository {
	return &TeacherRepository{
		teacherDAO: teacherDAO,
	}
}

// Create 创建教师
func (r *TeacherRepository) Create(ctx context.Context, teacher *domain.Teacher) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(teacher.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	teacher.Password = string(hashedPassword)
	return r.teacherDAO.Create(ctx, teacher)
}

// Login 教师登录
func (r *TeacherRepository) Login(ctx context.Context, teacherId, password string) (*domain.Teacher, error) {
	teacher, err := r.teacherDAO.FindByTeacherAccount(ctx, teacherId)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(password))
	if err != nil {
		return nil, err
	}
	return teacher, nil
}

// Register 教师注册
func (r *TeacherRepository) Register(ctx context.Context, teacherId, password string) error {
	//判断教师是否存在
	teacher, err := r.teacherDAO.FindByTeacherAccount(ctx, teacherId)

	//教师存在不可以在创建
	if err == nil {
		return errors.New("教师已存在")
	}

	//创建教师
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	teacher = &domain.Teacher{
		TeacherId: teacherId,
		Password:  string(hashedPassword),
	}
	return r.teacherDAO.Create(ctx, teacher)
}

// FindById 根据ID查找教师
func (r *TeacherRepository) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Teacher, error) {
	return r.teacherDAO.FindById(ctx, id)
}

// Update 更新教师信息
func (r *TeacherRepository) Update(ctx context.Context, teacher *domain.Teacher) error {
	return r.teacherDAO.Update(ctx, teacher)
}

// GetSettings 获取教师设置
func (r *TeacherRepository) GetSettings(ctx context.Context, teacherId string) (*domain.SettingsData, error) {
	return r.teacherDAO.GetSettings(ctx, teacherId)
}

// SaveSettings 保存教师设置
func (r *TeacherRepository) SaveSettings(ctx context.Context, teacherId string, settings domain.SettingsData) error {
	return r.teacherDAO.SaveSettings(ctx, teacherId, settings)
}

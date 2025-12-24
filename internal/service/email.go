package service

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailServiceInterface interface {
	ClearEmail(*gin.Context, primitive.ObjectID) error
	DeleteEmail(primitive.ObjectID, string) error
	GetUnread(*gin.Context, primitive.ObjectID) (int, error)
	GetUrlsByTeacherId(*gin.Context, primitive.ObjectID) ([]domain.EmailUrl, error)
}
type EmailService struct {
	repo repository.EmailInterface
}

func NewEmailService(repo repository.EmailInterface) EmailServiceInterface {
	return &EmailService{
		repo: repo,
	}
}
func (e *EmailService) GetUrlsByTeacherId(c *gin.Context, teacherId primitive.ObjectID) ([]domain.EmailUrl, error) {
	return e.repo.GetUrlsByTeacherId(c.Request.Context(), teacherId)
}
func (e *EmailService) GetUnread(c *gin.Context, teacherId primitive.ObjectID) (int, error) {
	return e.repo.GetUnread(c.Request.Context(), teacherId)
}
func (e *EmailService) ClearEmail(c *gin.Context, teacherId primitive.ObjectID) error {
	return e.repo.DeleteUnread(c.Request.Context(), teacherId)
}
func (e *EmailService) DeleteEmail(teacherID primitive.ObjectID, taskID string) error {
	return e.repo.DeleteEmailByTaskIdAndTeacherID(teacherID, taskID)
}

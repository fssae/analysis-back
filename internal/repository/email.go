package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailInterface interface {
	DeleteUnread(context.Context, primitive.ObjectID) error
	DeleteEmailByTaskIdAndTeacherID(teacherID primitive.ObjectID, taskID string) error
	GetUnread(context.Context, primitive.ObjectID) (int, error)
	GetUrlsByTeacherId(context.Context, primitive.ObjectID) ([]domain.EmailUrl, error)
}
type EmailRepository struct {
	emailDao dao.EmailDaoInterface
}

func NewEmailRepository(email dao.EmailDaoInterface) EmailInterface {
	return &EmailRepository{
		emailDao: email,
	}

}
func (e *EmailRepository) GetUrlsByTeacherId(c context.Context, teacherId primitive.ObjectID) ([]domain.EmailUrl, error) {
	email, err := e.emailDao.GetAllByTeacherId(c, teacherId)
	var emailUrls []domain.EmailUrl
	for _, url := range email.Urls {
		if url.Url == "" {
			continue
		}
		emailUrls = append(emailUrls, domain.EmailUrl{
			TaskId:     url.TaskId,
			Time:       url.Time,
			Url:        url.Url,
			Confidence: url.Confidence,
		})
	}
	return emailUrls, err
}
func (e *EmailRepository) GetUnread(c context.Context, teacherId primitive.ObjectID) (int, error) {
	return e.emailDao.GetUnread(c, teacherId)
}
func (e *EmailRepository) DeleteUnread(ctx context.Context, teacherId primitive.ObjectID) error {
	return e.emailDao.DeleteUnread(ctx, teacherId)
}
func (e *EmailRepository) DeleteEmailByTaskIdAndTeacherID(teacherID primitive.ObjectID, taskID string) error {
	return e.emailDao.DeleteByTaskId(teacherID, taskID)
}

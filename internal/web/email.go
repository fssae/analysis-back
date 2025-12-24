package web

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/service"
	"classroom-analysis/internal/util"
	"classroom-analysis/internal/web/middleware"
	"errors"
	"gitee.com/huahua20414/pkgx/ginx"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailHandler struct {
	svc service.EmailServiceInterface
}

func NewEmailHandler(svc service.EmailServiceInterface) *EmailHandler {
	return &EmailHandler{
		svc: svc,
	}
}

func (m *EmailHandler) RegisterRoutes(server *gin.Engine) {
	authorized := server.Group("/")
	authorized.Use(
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/teacher/login").
			IgnorePaths("/teacher/register").
			//ReWritPaths("/teacher/login").
			Build(),
	)
	authorized.GET("/email/clearUnread", ginx.Wrap(m.ClearUnread))
	authorized.POST("/email/deleteByTaskId", ginx.WrapBody[domain.EmailDeleteRequest](m.DeleteEmail))
	authorized.GET("/email/getUnread", ginx.Wrap(m.GetUnread))
	authorized.GET("/email/getUrlsByTeacherId", ginx.Wrap(m.GetUrlsByTeacherId))
}
func (m *EmailHandler) GetUrlsByTeacherId(c *gin.Context) (ginx.Response, error) {
	claims, err := util.GetClaims(c)
	if err != nil {
		return ginx.ErrorMess("token过期，请重新登录", err), nil
	}
	urls, err := m.svc.GetUrlsByTeacherId(c, claims.Id)
	if err != nil {
		ginx.ErrorMess("获取错误", err)
	}
	return ginx.SuccessMess("成功", gin.H{"urls": urls}), nil
}
func (m *EmailHandler) GetUnread(c *gin.Context) (ginx.Response, error) {
	claims, err := util.GetClaims(c)
	if err != nil {
		return ginx.ErrorMess("token过期，请重新登录", err), nil
	}
	unread, err := m.svc.GetUnread(c, claims.Id)
	if err != nil {
		ginx.ErrorMess("获取错误", err)
	}
	return ginx.SuccessMess("成功", gin.H{"unreadNumber": unread}), nil
}
func (m *EmailHandler) ClearUnread(c *gin.Context) (ginx.Response, error) {
	claims, err := util.GetClaims(c)
	if err != nil {
		return ginx.ErrorMess("jwt过期，请重新登录", nil), nil
	}
	err = m.svc.ClearEmail(c, claims.Id)
	if err != nil {
		return ginx.ErrorMess("清理邮箱出错", err.Error()), err
	}
	return ginx.SuccessMess("删除成功", nil), nil
}
func (m *EmailHandler) DeleteEmail(c *gin.Context, req domain.EmailDeleteRequest) (ginx.Response, error) {
	claims, err := util.GetClaims(c)
	if err != nil {
		return ginx.ErrorMess("jwt过期，请重新登录", nil), nil
	}
	err = m.svc.DeleteEmail(claims.Id, req.TaskId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ginx.SuccessMess("结果为空", err.Error()), nil
		}
		return ginx.ErrorMess("删除错误", err.Error()), nil
	}
	return ginx.SuccessMess("删除成功", nil), nil
}

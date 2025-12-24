package web

import (
	"classroom-analysis/internal/service"
	"classroom-analysis/internal/web/middleware"
	"gitee.com/huahua20414/pkgx/ginx"
	"github.com/gin-gonic/gin"
)

type FaceApiHandler struct {
	svc service.EmailServiceInterface
}

func NewFaceApiHandler(svc service.EmailServiceInterface) *FaceApiHandler {
	return &FaceApiHandler{
		svc: svc,
	}
}

func (m *FaceApiHandler) RegisterRoutes(server *gin.Engine) {
	faceAPi := server.Group("/")
	faceAPi.Use(
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/teacher/login").
			IgnorePaths("/teacher/register").
			//ReWritPaths("/teacher/login").
			Build(),
	)
	faceAPi.GET("/faceApi/call", ginx.Wrap(m.Call))
}
func (m *FaceApiHandler) Call(c *gin.Context) (ginx.Response, error) {
	//TODO
	// 1. 获取上传的人脸图片
	// 2. 调用第三方专注度检测API
	// 3. 解析返回的专注度数据
	// 4. 返回分析结果
	return ginx.Response{
		Code:    200,
		Message: "success",
		Data:    nil,
	}, nil
}

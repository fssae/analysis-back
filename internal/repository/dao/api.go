package dao

import (
	"fmt"
	"gitee.com/huahua20414/pkgx/ginx"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApiDaoInterface interface {
	FindByFilter(ctx *gin.Context, page ginx.Page, url string, method string) ([]Api, int64, error)
	FindAllApis(ctx *gin.Context) ([]Api, int64, error)
}
type ApiDao struct {
	col *mongo.Collection
}

func (a *ApiDao) FindAllApis(ctx *gin.Context) ([]Api, int64, error) {
	filter := bson.M{}
	cur, err := a.col.Find(ctx, filter)

	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var apis []Api
	for cur.Next(ctx) {
		var api Api
		if err := cur.Decode(&api); err != nil {
			return nil, 0, err
		}
		apis = append(apis, api)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}
	count, err := a.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return apis, count, nil
}

func (a *ApiDao) FindByFilter(ctx *gin.Context, page ginx.Page, url string, method string) ([]Api, int64, error) {
	filter := bson.M{}
	if url != "" {
		regex := fmt.Sprintf(".*%v.*", url)
		filter["url"] = primitive.Regex{Pattern: regex}
	}
	if method != "" {
		filter["method"] = method
	}
	opts := page.GetOpts()
	cur, err := a.col.Find(ctx, filter, opts)

	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var apis []Api
	for cur.Next(ctx) {
		var api Api
		if err := cur.Decode(&api); err != nil {
			return nil, 0, err
		}
		apis = append(apis, api)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}
	count, err := a.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return apis, count, nil
}

func NewApiDao(client *mongo.Client) ApiDaoInterface {
	return &ApiDao{
		col: client.Database("shared").Collection("api"),
	}
}

type Api struct {
	Url    string `json:"url" bson:"url"`
	Method string `json:"method" bson:"method"`
}

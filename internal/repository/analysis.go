package repository

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository/dao"
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalysisRepository struct {
	analysisDAO *dao.AnalysisDAO
	emailDAO    dao.EmailDaoInterface
	fileRepo    FileInterface
}

func NewAnalysisRepository(analysisDAO *dao.AnalysisDAO,
	emailDAO dao.EmailDaoInterface,
	fileRepo FileInterface) *AnalysisRepository {
	return &AnalysisRepository{
		analysisDAO: analysisDAO,
		emailDAO:    emailDAO,
		fileRepo:    fileRepo,
	}
}

func (r *AnalysisRepository) UpdateConfig(update *domain.UpdateConfigRequest) error {
	return r.analysisDAO.UpdateConfig(update)
}

func (r *AnalysisRepository) UpdateAnalysisName(update *domain.UpdateAnalysisNameRequest) error {
	return r.analysisDAO.UpdateAnalysisName(update)
}

func (r *AnalysisRepository) UpdateStatus(update *domain.UpdateStatus) error {
	err := r.analysisDAO.UpdateStatus(update)
	if err != nil {
		return err
	}
	err = r.emailDAO.EmailUpdate(update.TeacherId, update.TaskId, update.ResultUrl, update.ConfidenceThreshold)
	if err != nil {
		return err
	}
	return nil
}

func (r *AnalysisRepository) FindByTaskId(ctx context.Context, taskId string) (*domain.Analysis, error) {
	return r.analysisDAO.FindByTaskId(ctx, taskId)
}

func (r *AnalysisRepository) FindByImageIdString(ctx context.Context, imageIdStr string) (*domain.Analysis, error) {
	return r.analysisDAO.FindByImageIdString(ctx, imageIdStr)
}

func (r *AnalysisRepository) UpdateFileNameByTaskId(ctx context.Context, taskId, fileName string) error {
	return r.analysisDAO.UpdateFileNameByTaskId(ctx, taskId, fileName)
}
func (r *AnalysisRepository) GetRank(ctx context.Context, req *domain.RankRequest) (*domain.RankItem, error) {
	return r.analysisDAO.GetRankDao(ctx, req)
}
func (r *AnalysisRepository) FindRecentAnalysis(ctx context.Context) ([]*domain.Analysis, error) {
	return r.analysisDAO.FindRecentAnalysis(ctx)
}

// Create 创建分析记录
func (r *AnalysisRepository) Create(ctx context.Context, analysis *domain.Analysis) error {
	return r.analysisDAO.Create(ctx, analysis)
}

// FindById 根据ID查找分析记录
func (r *AnalysisRepository) FindById(ctx context.Context, id primitive.ObjectID) (*domain.Analysis, error) {
	return r.analysisDAO.FindById(ctx, id)
}

// FindByTeacherId 根据教师ID查找分析记录
func (r *AnalysisRepository) FindByTeacherId(ctx context.Context, limit int64) ([]*domain.Analysis, error) {
	return r.analysisDAO.FindByTeacherId(ctx, limit)
}

// Update 更新分析记录
func (r *AnalysisRepository) Update(ctx context.Context, analysis *domain.Analysis) error {
	return r.analysisDAO.Update(ctx, analysis)
}

// UpdateAnalysisStatus 更新分析记录状态
func (r *AnalysisRepository) UpdateAnalysisStatus(ctx context.Context, imageId primitive.ObjectID, status string) error {
	return r.analysisDAO.UpdateAnalysisStatus(ctx, imageId, status)
}

// CountByTeacherId 统计教师的分析数量
func (r *AnalysisRepository) CountByTeacherId(ctx context.Context) (int64, int64, int64, int64, error) {
	return r.analysisDAO.CountByTeacherId(ctx)
}

// GetLastAnalysisTime 获取最后分析时间
func (r *AnalysisRepository) GetLastAnalysisTime(ctx context.Context, teacherId primitive.ObjectID) (*time.Time, error) {
	return r.analysisDAO.GetLastAnalysisTime(ctx, teacherId)
}

func (r *AnalysisRepository) GetAnalysisDAO() *dao.AnalysisDAO {
	return r.analysisDAO
}

// DeleteById 根据ID删除分析记录
func (r *AnalysisRepository) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.analysisDAO.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// 获取班级分析列表
func (r *AnalysisRepository) GetClassAnalysisList(
	ctx context.Context,
	courseName, className, startDate, endDate string,
	page, pageSize int,
) ([]domain.ClassAnalysisItem, int, error) {
	analyses, total, err := r.analysisDAO.FindClassAnalysisList(ctx, courseName, className, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var result []domain.ClassAnalysisItem
	for _, a := range analyses {
		var sum float64
		for _, f := range a.Faces {
			sum += f.FocusScore
		}
		focusAvg := 0.0
		if len(a.Faces) > 0 {
			focusAvg = sum / float64(len(a.Faces))
		}
		// 返回结果时给班级名称加上"班"字
		className := a.ClassName
		if className != "" && !strings.HasSuffix(className, "班") {
			className += "班"
		}
		item := domain.ClassAnalysisItem{
			ImageId:    string(a.ImageId.Hex()),
			CourseName: a.CourseName,
			FileName:   a.FileName,
			ClassName:  className,
			Date:       a.Timestamp.Format("2006-01-02 15:04"),
			FocusAvg:   focusAvg,
			ResultUrl:  a.ResultUrl,
		}
		result = append(result, item)
	}
	return result, int(total), nil
}

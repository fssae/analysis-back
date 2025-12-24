package service

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeacherService struct {
	teacherRepo       *repository.TeacherRepository
	analysisRepo      *repository.AnalysisRepository
	videoAnalysisRepo *repository.VideoAnalysisRepository
	studentRepo       *repository.StudentRepository
	studentFocusRepo  *repository.StudentFocusRepository
}

func NewTeacherService(
	teacherRepo *repository.TeacherRepository,
	analysisRepo *repository.AnalysisRepository,
	videoAnalysisRepo *repository.VideoAnalysisRepository,
	studentRepo *repository.StudentRepository,
	studentFocusRepo *repository.StudentFocusRepository,
) *TeacherService {
	return &TeacherService{
		teacherRepo:       teacherRepo,
		analysisRepo:      analysisRepo,
		videoAnalysisRepo: videoAnalysisRepo,
		studentRepo:       studentRepo,
		studentFocusRepo:  studentFocusRepo,
	}
}
func (s *TeacherService) GetRank(ctx context.Context, req *domain.RankRequest) (*domain.RankItem, error) {
	return s.analysisRepo.GetRank(ctx, req)
}

// Login 教师登录
func (s *TeacherService) Login(ctx context.Context, req *domain.TeacherLoginRequest) (*domain.TeacherLoginResponse, error) {
	teacher, err := s.teacherRepo.Login(ctx, req.TeacherId, req.Password)
	if err != nil {
		return nil, errors.New("工号或密码错误")
	}
	// 生成JWT
	tokenString, err := createToken(teacher.Id, teacher.TeacherId)
	return &domain.TeacherLoginResponse{
		Token:   tokenString,
		Teacher: teacher,
	}, nil
}

// Register 教师注册
func (s *TeacherService) Register(ctx context.Context, req *domain.TeacherRegisterRequest) error {
	err := s.teacherRepo.Register(ctx, req.TeacherId, req.Password)

	if err != nil { // 处理注册失败的情况
		return err
	}

	//注册成功
	return nil
}

// GetDashboard 获取仪表盘数据
func (s *TeacherService) GetDashboard(ctx context.Context) (*domain.DashboardData, error) {
	// 获取分析总数
	totalImages, totalVideos, totalFaces, totalDocs, err := s.analysisRepo.CountByTeacherId(ctx)
	if err != nil {
		return nil, err
	}
	//获取最近分析记录
	recentAnalysis, err := s.analysisRepo.FindRecentAnalysis(ctx)
	if err != nil {
		return nil, err
	}
	//获取专注度趋势数据，默认本周
	focusTrend, err := s.studentFocusRepo.GetFocusTrend(ctx)
	if err != nil {
		return nil, err
	}
	//获取专注度分布数据
	focusDistribution, err := s.studentFocusRepo.GetFocusDistribution(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.DashboardData{
		TotalVideos:       totalVideos,
		TotalImages:       totalImages,
		TotalStudents:     totalFaces,
		TotalAnalysis:     totalDocs,
		RecentRecords:     recentAnalysis,
		FocusTrend:        focusTrend,
		FocusDistribution: focusDistribution,
	}, nil
}
func (s *TeacherService) GetImageAnalysisDetail(ctx context.Context, id string) (domain.AnalysisResult, error) {
	//平均专注度 (avgFocus)
	//专注学生数 (focusedStudents)，分心学生数 (distractedStudents),专注度分布数据 (distributionData)
	avgFocus, focusedStudents, distractedStudents, err := s.studentFocusRepo.GetFocusByStudentId(ctx, id)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	return domain.AnalysisResult{
		AvgFocus:           avgFocus,
		DistractedStudents: distractedStudents,
		FocusedStudents:    focusedStudents,
	}, nil
}

// GetVideoAnalysisDetail 获取视频分析详情
func (s *TeacherService) GetVideoAnalysisDetail(ctx context.Context, id string) (*domain.VideoAnalysisDetail, error) {
	// 根据ID获取视频分析详情
	// 这里需要从多个数据源聚合数据
	analysisId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// 获取基础分析信息
	analysis, err := s.analysisRepo.FindById(ctx, analysisId)
	if err != nil {
		return nil, err
	}

	// 获取视频分析详情
	videoAnalysis, err := s.videoAnalysisRepo.FindByAnalysisId(ctx, analysisId)
	if err != nil {
		return nil, err
	}

	// 获取学生专注度数据
	studentFocuses, err := s.studentFocusRepo.FindByAnalysisId(ctx, analysisId)
	if err != nil {
		return nil, err
	}

	// 计算统计数据
	var totalFocus float64
	var maxFocus, minFocus float64
	var focusedStudents, distractedStudents int
	var focusData []float64

	if len(studentFocuses) > 0 {
		maxFocus = studentFocuses[0].FocusScore
		minFocus = studentFocuses[0].FocusScore

		for _, focus := range studentFocuses {
			totalFocus += focus.FocusScore
			focusData = append(focusData, focus.FocusScore)

			if focus.FocusScore > maxFocus {
				maxFocus = focus.FocusScore
			}
			if focus.FocusScore < minFocus {
				minFocus = focus.FocusScore
			}

			// 判断专注/分心学生
			if focus.FocusScore >= 0.7 {
				focusedStudents++
			} else {
				distractedStudents++
			}
		}
	}

	avgFocus := 0.0
	if len(studentFocuses) > 0 {
		avgFocus = totalFocus / float64(len(studentFocuses))
	}

	// 构建响应数据
	result := &domain.VideoAnalysisDetail{
		Id:                 id,
		AvgFocus:           avgFocus,
		MaxFocus:           maxFocus,
		MinFocus:           minFocus,
		Duration:           videoAnalysis.Duration,
		FocusedStudents:    focusedStudents,
		DistractedStudents: distractedStudents,
		ResultURL:          fmt.Sprintf("http://82.156.64.69:9000/analysis/%s.mp4", id),
		FocusData:          focusData,
		AnalysisTime:       analysis.Timestamp,
		CourseName:         "数学课",  // 这里可以从其他地方获取
		ClassName:          "高三一班", // 这里可以从其他地方获取
		StudentCount:       len(studentFocuses),
		Accuracy:           "medium",
	}

	return result, nil
}

// GetStudents 获取学生专注度数据
func (s *TeacherService) GetStudents(ctx context.Context) ([]map[string]interface{}, error) {
	students, err := s.studentRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, student := range students {
		avgFocus, err := s.studentFocusRepo.GetAverageFocusByStudentId(ctx, student.Id)
		if err != nil {
			avgFocus = 0
		}

		// 获取最后分析时间
		lastFocus, err := s.studentFocusRepo.FindByStudentId(ctx, student.Id, 1)
		var lastAnalysis string
		if err == nil && len(lastFocus) > 0 {
			lastAnalysis = lastFocus[0].Date.Format("2006-01-02")
		}

		result = append(result, map[string]interface{}{
			"studentId":    student.Id.Hex(),
			"name":         student.Name,
			"focusAvg":     avgFocus,
			"lastAnalysis": lastAnalysis,
		})
	}

	return result, nil
}

// 根据信息查询学生
func (s *TeacherService) FindStudents(ctx context.Context, name string) ([]map[string]interface{}, error) {
	students, err := s.studentRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, student := range students {
		avgFocus, err := s.studentFocusRepo.GetAverageFocusByStudentId(ctx, student.Id)
		if err != nil {
			avgFocus = 0
		}

		// 获取最后分析时间
		lastFocus, err := s.studentFocusRepo.FindByStudentId(ctx, student.Id, 1)
		var lastAnalysis string
		if err == nil && len(lastFocus) > 0 {
			lastAnalysis = lastFocus[0].Date.Format("2006-01-02")
		}

		result = append(result, map[string]interface{}{
			"studentId":    student.Id.Hex(),
			"name":         student.Name,
			"focusAvg":     avgFocus,
			"lastAnalysis": lastAnalysis,
		})
	}
	return result, nil
}

// GetTeacherById 根据ID获取教师信息
func (s *TeacherService) GetTeacherById(ctx context.Context, teacherId primitive.ObjectID) (*domain.Teacher, error) {
	return s.teacherRepo.FindById(ctx, teacherId)
}

// UpdateSettings 更新教师设置
func (s *TeacherService) UpdateSettings(ctx context.Context, teacherId primitive.ObjectID, email string) error {
	teacher, err := s.teacherRepo.FindById(ctx, teacherId)
	if err != nil {
		return err
	}

	teacher.Email = email
	// 注意：这里简化处理，实际应该有一个单独的settings表
	// 或者扩展Teacher结构体来包含notification字段

	return s.teacherRepo.Update(ctx, teacher)
}

// 获取班级分析列表
func (s *TeacherService) GetClassAnalysisList(
	ctx context.Context,
	courseName, className, startDate, endDate string,
	page, pageSize int,
) ([]domain.ClassAnalysisItem, int, error) {
	return s.analysisRepo.GetClassAnalysisList(ctx, courseName, className, startDate, endDate, page, pageSize)
}

func createToken(Id primitive.ObjectID, teacherId string) (tokenString string, err error) {
	secret := viper.GetString("general.jwt")
	claims := domain.TeacherClaims{
		//设置参数
		RegisteredClaims: jwt.RegisteredClaims{
			//设置300天的过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 300)),
		},
		Id:        Id,
		TeacherId: teacherId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//加密
	tokenStr, err := token.SignedString([]byte(secret))
	return tokenStr, err
}

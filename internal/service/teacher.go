package service

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository"
	"classroom-analysis/internal/util"
	"context"
	"errors"
	"fmt"

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
	tokenString, err := createToken(teacher)
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
	// 验证输入参数
	if id == "" {
		return nil, errors.New("analysis ID is required")
	}

	// 首先通过 imageId 获取 analysis 记录
	analysisResult, err := s.analysisRepo.FindByImageIdString(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("analysis not found for imageId: %s", id)
	}

	// 初始化统计数据
	var totalFocus float64
	var maxFocus, minFocus float64
	var focusedStudents, distractedStudents int
	var focusData []float64

	// 使用 analysis.faces 中的数据进行统计计算
	if len(analysisResult.Faces) > 0 {
		// 初始化最大值和最小值为第一个元素的专注度分数
		maxFocus = analysisResult.Faces[0].FocusScore
		minFocus = analysisResult.Faces[0].FocusScore

		for _, face := range analysisResult.Faces {
			totalFocus += face.FocusScore
			focusData = append(focusData, face.FocusScore*100)

			if face.FocusScore > maxFocus {
				maxFocus = face.FocusScore
			}
			if face.FocusScore < minFocus {
				minFocus = face.FocusScore
			}

			// 判断专注/分心学生
			if face.FocusScore >= 0.7 {
				focusedStudents++
			} else {
				distractedStudents++
			}
		}
	} else {
		// 如果没有 faces 数据，则使用默认值
		maxFocus = 0
		minFocus = 0
	}

	// 计算平均专注度
	avgFocus := 0.0
	studentCount := len(analysisResult.Faces)
	if studentCount > 0 {
		avgFocus = totalFocus / float64(studentCount)
	}

	// 构建响应数据
	// 注意：Duration 默认设置为 0，因为 analysis 集合中没有视频时长字段
	response := &domain.VideoAnalysisDetail{
		Id:                 id,
		AvgFocus:           avgFocus,
		MaxFocus:           maxFocus,
		MinFocus:           minFocus,
		Duration:           0, // 从 analysis 集合无法获取视频时长，设置为默认值
		FocusedStudents:    focusedStudents,
		DistractedStudents: distractedStudents,
		ResultURL:          analysisResult.ResultUrl, // 使用数据库中保存的真实结果URL
		FocusData:          focusData,
		AnalysisTime:       analysisResult.Timestamp,
		CourseName:         analysisResult.CourseName, // 使用数据库中的真实课程名
		ClassName:          analysisResult.ClassName,  // 使用数据库中的真实班级名
		StudentCount:       studentCount,
		Accuracy:           "medium", // 这个值可以根据实际算法调整
	}

	return response, nil
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
			lastAnalysis = util.FormatBeijingTime(lastFocus[0].Date, "2006-01-02")
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
			lastAnalysis = util.FormatBeijingTime(lastFocus[0].Date, "2006-01-02")
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

// GetSettings 获取教师设置
func (s *TeacherService) GetSettings(ctx context.Context, teacherId string) (*domain.SettingsData, error) {
	// 从数据库获取设置，如果不存在则返回默认值
	settings, err := s.teacherRepo.GetSettings(ctx, teacherId)
	if err != nil {
		// 返回默认设置
		return &domain.SettingsData{
			UISettings: domain.UISettings{
				Theme:            "light",
				SidebarCollapsed: false,
				EnableAnimation:  true,
			},
			AnalysisSettings: domain.AnalysisSettings{
				DefaultCourse:    "",
				DefaultClass:     "",
				FatigueThreshold: 70,
				FocusThreshold:   60,
			},
			NotificationSettings: domain.NotificationSettings{
				Enabled:           true,
				FatigueAlert:      true,
				FocusAlert:        true,
				EmailNotification: false,
			},
		}, nil
	}
	return settings, nil
}

// UpdateSettings 更新教师设置
func (s *TeacherService) UpdateSettings(ctx context.Context, teacherId string, settings domain.SettingsData) error {
	return s.teacherRepo.SaveSettings(ctx, teacherId, settings)
}

// UpdateTeacherEmail 更新教师邮箱
func (s *TeacherService) UpdateTeacherEmail(ctx context.Context, teacherId primitive.ObjectID, email string) error {
	teacher, err := s.teacherRepo.FindById(ctx, teacherId)
	if err != nil {
		return err
	}

	teacher.Email = email
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

func createToken(teacher *domain.Teacher) (tokenString string, err error) {
	secret := viper.GetString("general.jwt")
	claims := domain.NewTeacherClaims(teacher)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//加密
	tokenStr, err := token.SignedString([]byte(secret))
	return tokenStr, err
}

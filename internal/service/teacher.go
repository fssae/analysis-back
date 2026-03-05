package service

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// 获取到 analysis 的 _id，用于查询关联表
	analysisId := analysisResult.Id

	// 并行获取视频分析详情和学生专注度数据以提高性能
	type result struct {
		videoAnalysis  *domain.VideoAnalysis
		studentFocuses []*domain.StudentFocus // 修改为指针类型
		err            error
		index          int
	}

	// 创建带缓冲区的通道，用于接收两个goroutine的结果
	ch := make(chan result, 2)

	// 获取视频分析详情
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- result{err: fmt.Errorf("panic occurred while getting video analysis: %v", r), index: 0}
			}
		}()

		videoAnalysis, err := s.videoAnalysisRepo.FindByAnalysisId(ctx, analysisId)
		ch <- result{videoAnalysis: videoAnalysis, err: err, index: 0}
	}()

	// 获取学生专注度数据
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- result{err: fmt.Errorf("panic occurred while getting student focuses: %v", r), index: 1}
			}
		}()

		studentFocuses, err := s.studentFocusRepo.FindByAnalysisId(ctx, analysisId)
		ch <- result{studentFocuses: studentFocuses, err: err, index: 1}
	}()

	// 收集结果
	var videoAnalysis *domain.VideoAnalysis
	var studentFocusPtrs []*domain.StudentFocus // 修改变量名为更清晰的名称

	var resultsReceived int
	var criticalErr error // 关键错误（非"未找到"错误）

	// 等待所有goroutine完成（现在只有2个）
	for resultsReceived < 2 {
		select {
		case res := <-ch:
			resultsReceived++

			switch res.index {
			case 0: // videoAnalysis
				if res.err != nil {
					if res.err == mongo.ErrNoDocuments {
						// 对于视频分析不存在的情况，我们返回错误，因为这是必须的
						criticalErr = fmt.Errorf("video analysis not found for analysis ID: %s", id)
					} else {
						criticalErr = fmt.Errorf("failed to get video analysis: %v", res.err)
					}
				} else {
					videoAnalysis = res.videoAnalysis
				}
			case 1: // studentFocuses
				if res.err != nil && res.err != mongo.ErrNoDocuments {
					// 对于学生专注度数据，如果只是没找到文档，这是正常的（可能没有学生数据）
					// 只有其他错误才视为关键错误
					criticalErr = fmt.Errorf("failed to get student focuses: %v", res.err)
				} else {
					studentFocusPtrs = res.studentFocuses
				}
			}
		case <-ctx.Done():
			// 如果上下文被取消，立即返回
			return nil, ctx.Err()
		}
	}

	// 检查是否存在关键错误
	if criticalErr != nil {
		return nil, criticalErr
	}

	// 检查视频分析是否存在（这也是必须的，因为视频分析详情依赖于它）
	if videoAnalysis == nil {
		return nil, fmt.Errorf("video analysis not found for analysis ID: %s", id)
	}

	// studentFocusPtrs 可能为空数组，这是正常情况，表示没有学生专注度数据

	// 初始化统计数据
	var totalFocus float64
	var maxFocus, minFocus float64
	var focusedStudents, distractedStudents int
	var focusData []float64
	var studentFocuses []domain.StudentFocus // 值切片用于后续处理

	// 优先使用 analysis.faces 中的数据进行统计计算（因为这是直接的视频分析数据）
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
		// 如果没有 analysis.faces 数据，则使用学生专注度数据作为备选
		if len(studentFocusPtrs) > 0 {
			// 初始化最大值和最小值为第一个元素
			maxFocus = studentFocusPtrs[0].FocusScore
			minFocus = studentFocusPtrs[0].FocusScore

			for _, focusPtr := range studentFocusPtrs {
				// 添加到值切片中
				studentFocuses = append(studentFocuses, *focusPtr)

				totalFocus += focusPtr.FocusScore
				focusData = append(focusData, focusPtr.FocusScore)

				if focusPtr.FocusScore > maxFocus {
					maxFocus = focusPtr.FocusScore
				}
				if focusPtr.FocusScore < minFocus {
					minFocus = focusPtr.FocusScore
				}

				// 判断专注/分心学生
				if focusPtr.FocusScore >= 0.7 {
					focusedStudents++
				} else {
					distractedStudents++
				}
			}
		} else {
			// 如果都没有数据，则使用默认值
			maxFocus = 0
			minFocus = 0
		}
	}

	// 计算平均专注度
	avgFocus := 0.0
	var studentCount int
	if len(studentFocuses) > 0 {
		studentCount = len(studentFocuses)
		avgFocus = totalFocus / float64(len(studentFocuses))
	} else if len(analysisResult.Faces) > 0 {
		studentCount = len(analysisResult.Faces)
		avgFocus = totalFocus / float64(len(analysisResult.Faces))
	}

	// 构建响应数据
	response := &domain.VideoAnalysisDetail{
		Id:                 id,
		AvgFocus:           avgFocus,
		MaxFocus:           maxFocus,
		MinFocus:           minFocus,
		Duration:           videoAnalysis.Duration,
		FocusedStudents:    focusedStudents,
		DistractedStudents: distractedStudents,
		ResultURL:          fmt.Sprintf("http://82.156.64.69:9000/analysis/%s.mp4", id),
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

func createToken(teacher *domain.Teacher) (tokenString string, err error) {
	secret := viper.GetString("general.jwt")
	claims := domain.NewTeacherClaims(teacher)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//加密
	tokenStr, err := token.SignedString([]byte(secret))
	return tokenStr, err
}

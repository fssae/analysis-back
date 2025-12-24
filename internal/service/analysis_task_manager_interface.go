package service

import (
	"classroom-analysis/internal/domain"
	"context"
)

// AnalysisTaskManagerInterface 分析任务管理器接口
// 定义接口便于测试和依赖注入
type AnalysisTaskManagerInterface interface {
	// ProcessAnalysisTask 处理分析任务
	ProcessAnalysisTask(ctx context.Context, req domain.TeacherAnalysisRequest) (*domain.KafkaResp, error)
}

// 确保AnalysisTaskManager实现了接口
var _ AnalysisTaskManagerInterface = (*AnalysisTaskManager)(nil)

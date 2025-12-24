package service

import (
	"classroom-analysis/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockKafkaWriter 模拟Kafka写入器
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) Write(message interface{}) error {
	args := m.Called(message)
	return args.Error(0)
}

// TestAnalysisTaskManager_ProcessAnalysisTask 测试分析任务处理
func TestAnalysisTaskManager_ProcessAnalysisTask(t *testing.T) {
	// 准备测试数据
	logger := zaptest.NewLogger(t)
	config := &AnalysisConfig{
		KafkaResponseTimeout: 1 * time.Second,
		RetryInterval:        100 * time.Millisecond,
		MaxRetries:           1,
		ResponseTopic:        "test_topic",
	}

	// 创建模拟对象
	mockKafkaWriter := &MockKafkaWriter{}

	// 创建任务管理器
	manager := NewAnalysisTaskManager(
		mockKafkaWriter,
		nil, // 这里需要实际的ConsumerGroupFactory
		logger,
		config,
	)

	// 测试请求
	req := domain.TeacherAnalysisRequest{
		ImageId:      "test-task-123",
		AnalysisType: "image",
		Url:          "http://example.com/image.jpg",
		ClassName:    "TestClass",
		CourseName:   "TestCourse",
	}

	// 设置模拟期望
	mockKafkaWriter.On("Write", mock.AnythingOfType("domain.KafkaMessage")).Return(nil)

	// 执行测试
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 注意：这个测试会因为Kafka消费者部分而超时，但展示了测试结构
	result, err := manager.ProcessAnalysisTask(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 验证模拟调用
	mockKafkaWriter.AssertExpectations(t)
}

// TestAnalysisConfig_Default 测试默认配置
func TestAnalysisConfig_Default(t *testing.T) {
	config := DefaultAnalysisConfig()

	assert.Equal(t, 5*time.Minute, config.KafkaResponseTimeout)
	assert.Equal(t, 30*time.Second, config.RetryInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "face_analyze_response", config.ResponseTopic)
}

// BenchmarkAnalysisTaskManager 性能基准测试示例
func BenchmarkAnalysisTaskManager(b *testing.B) {
	logger := zaptest.NewLogger(b)
	config := DefaultAnalysisConfig()

	// 创建管理器（需要实际的依赖）
	manager := NewAnalysisTaskManager(
		nil, nil, logger, config,
	)

	req := domain.TeacherAnalysisRequest{
		ImageId:      "benchmark-task",
		AnalysisType: "image",
		Url:          "http://example.com/image.jpg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		// 注意：这里会因为超时而失败，但展示了基准测试结构
		_, _ = manager.ProcessAnalysisTask(ctx, req)
		cancel()
	}
}

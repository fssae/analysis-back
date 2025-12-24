package service

import "time"

// AnalysisConfig 分析任务配置
type AnalysisConfig struct {
	// KafkaResponseTimeout Kafka响应超时时间
	KafkaResponseTimeout time.Duration
	
	// RetryInterval 重试间隔
	RetryInterval time.Duration
	
	// MaxRetries 最大重试次数
	MaxRetries int
	
	// ResponseTopic Kafka响应主题
	ResponseTopic string
}

// DefaultAnalysisConfig 返回默认配置
func DefaultAnalysisConfig() *AnalysisConfig {
	return &AnalysisConfig{
		KafkaResponseTimeout: 5 * time.Minute,
		RetryInterval:        30 * time.Second,
		MaxRetries:           3,
		ResponseTopic:        "face_analyze_response",
	}
}

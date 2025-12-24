package service

import (
	"classroom-analysis/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaWriter 定义用于写入Kafka的接口
type KafkaWriter interface {
	Write(message interface{}) error
}

// AnalysisTaskManager 分析任务管理器
type AnalysisTaskManager struct {
	kafkaWriter          KafkaWriter
	consumerGroup        sarama.ConsumerGroup
	logger               *zap.Logger
	config               *AnalysisConfig
	consumerGroupFactory func(groupId string) (sarama.ConsumerGroup, error)
}

// NewAnalysisTaskManager 创建分析任务管理器
func NewAnalysisTaskManager(
	kafkaWriter KafkaWriter,
	consumerGroupFactory func(groupId string) (sarama.ConsumerGroup, error),
	logger *zap.Logger,
	config *AnalysisConfig,
) *AnalysisTaskManager {
	if config == nil {
		config = DefaultAnalysisConfig()
	}

	return &AnalysisTaskManager{
		kafkaWriter:          kafkaWriter,
		consumerGroupFactory: consumerGroupFactory,
		logger:               logger,
		config:               config,
	}
}

// ProcessAnalysisTask 处理分析任务
func (atm *AnalysisTaskManager) ProcessAnalysisTask(ctx context.Context, req domain.TeacherAnalysisRequest) (*domain.KafkaResp, error) {
	// 1. 发送消息到Kafka
	if err := atm.sendKafkaMessage(req); err != nil {
		return nil, fmt.Errorf("发送Kafka消息失败: %w", err)
	}

	// 2. 等待响应
	return atm.waitForResponse(ctx, req.ImageId)
}

// sendKafkaMessage 发送Kafka消息
func (atm *AnalysisTaskManager) sendKafkaMessage(req domain.TeacherAnalysisRequest) error {
	message := domain.KafkaMessage{Req: req}

	atm.logger.Info("发送Kafka消息",
		zap.String("analysisType", req.AnalysisType),
		zap.String("taskId", req.ImageId),
		zap.String("url", req.Url))

	if err := atm.kafkaWriter.Write(message); err != nil {
		atm.logger.Error("Kafka发送失败",
			zap.String("taskId", req.ImageId),
			zap.Error(err))
		return err
	}

	atm.logger.Info("Kafka消息发送成功",
		zap.String("analysisType", req.AnalysisType),
		zap.String("taskId", req.ImageId),
		zap.String("url", req.Url))

	return nil
}

// waitForResponse 等待Kafka响应
func (atm *AnalysisTaskManager) waitForResponse(ctx context.Context, taskId string) (*domain.KafkaResp, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, atm.config.KafkaResponseTimeout)
	defer cancel()

	// 创建结果通道
	resultChan := make(chan *domain.KafkaResp, 1)
	errorChan := make(chan error, 1)

	// 启动消费者
	go atm.consumeResponse(ctx, taskId, resultChan, errorChan)

	// 等待结果
	select {
	case resp := <-resultChan:
		if resp == nil {
			return nil, fmt.Errorf("收到空响应")
		}
		atm.logger.Info("收到分析响应",
			zap.String("taskId", resp.ID),
			zap.String("resultURL", resp.ResultURL))
		return resp, nil

	case err := <-errorChan:
		return nil, fmt.Errorf("消费响应时出错: %w", err)

	case <-ctx.Done():
		return nil, fmt.Errorf("等待响应超时: %w", ctx.Err())
	}
}

// consumeResponse 消费Kafka响应消息
func (atm *AnalysisTaskManager) consumeResponse(ctx context.Context, taskId string, resultChan chan<- *domain.KafkaResp, errorChan chan<- error) {
	defer close(resultChan)
	defer close(errorChan)

	// 创建消费者组
	groupId := fmt.Sprintf("analysis-task-%s-%d", taskId, time.Now().UnixNano())
	consumerGroup, err := atm.consumerGroupFactory(groupId)
	if err != nil {
		atm.logger.Error("创建ConsumerGroup失败",
			zap.String("taskId", taskId),
			zap.String("groupId", groupId),
			zap.Error(err))
		errorChan <- fmt.Errorf("创建ConsumerGroup失败: %w", err)
		return
	}
	defer func() {
		if err := consumerGroup.Close(); err != nil {
			atm.logger.Error("关闭ConsumerGroup失败",
				zap.String("taskId", taskId),
				zap.String("groupId", groupId),
				zap.Error(err))
		}
	}()

	// 创建处理器
	handler := &AnalysisResponseHandler{
		taskId:     taskId,
		resultChan: resultChan,
		ctx:        ctx,
		logger:     atm.logger,
	}

	// 开始消费
	for {
		select {
		case <-ctx.Done():
			atm.logger.Info("消费上下文已取消",
				zap.String("taskId", taskId),
				zap.Error(ctx.Err()))
			return
		default:
			if err := consumerGroup.Consume(ctx, []string{atm.config.ResponseTopic}, handler); err != nil {
				atm.logger.Error("消费出错",
					zap.String("taskId", taskId),
					zap.String("topic", atm.config.ResponseTopic),
					zap.Error(err))

				if ctx.Err() != nil {
					return
				}

				// 重试机制
				select {
				case <-time.After(atm.config.RetryInterval):
					continue
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// AnalysisResponseHandler Kafka响应处理器
type AnalysisResponseHandler struct {
	taskId     string
	resultChan chan<- *domain.KafkaResp
	ctx        context.Context
	logger     *zap.Logger
}

func (h *AnalysisResponseHandler) Setup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("设置Kafka消费者处理器",
		zap.String("taskId", h.taskId))
	return nil
}

func (h *AnalysisResponseHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("清理Kafka消费者处理器",
		zap.String("taskId", h.taskId))
	return nil
}

func (h *AnalysisResponseHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	h.logger.Info("开始消费Kafka消息",
		zap.String("taskId", h.taskId),
		zap.String("topic", claim.Topic()),
		zap.Int32("partition", claim.Partition()))

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				h.logger.Info("Kafka消息通道已关闭",
					zap.String("taskId", h.taskId))
				return nil
			}

			h.logger.Debug("收到Kafka消息",
				zap.String("taskId", h.taskId),
				zap.String("topic", msg.Topic),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset))

			// 解析响应消息
			resp, err := h.parseResponseMessage(msg.Value)
			if err != nil {
				h.logger.Error("解析响应消息失败",
					zap.String("taskId", h.taskId),
					zap.Error(err))
				session.MarkMessage(msg, "")
				continue
			}

			// 检查是否为目标任务的响应
			if resp.ID == h.taskId {
				h.logger.Info("找到匹配的响应消息",
					zap.String("taskId", resp.ID),
					zap.String("resultURL", resp.ResultURL))

				select {
				case h.resultChan <- resp:
					h.logger.Info("成功发送响应到结果通道",
						zap.String("taskId", resp.ID))
				case <-h.ctx.Done():
					h.logger.Info("上下文已取消，停止发送响应",
						zap.String("taskId", resp.ID))
					return nil
				default:
					h.logger.Warn("结果通道已满，忽略响应",
						zap.String("taskId", resp.ID))
				}
				session.MarkMessage(msg, "")
				return nil // 找到匹配响应后直接返回
			}

			// 标记消息已处理（但不是目标消息）
			session.MarkMessage(msg, "")

		case <-h.ctx.Done():
			h.logger.Info("消费上下文已取消",
				zap.String("taskId", h.taskId),
				zap.Error(h.ctx.Err()))
			return nil
		}
	}
}

// parseResponseMessage 解析响应消息
func (h *AnalysisResponseHandler) parseResponseMessage(data []byte) (*domain.KafkaResp, error) {
	var resp domain.KafkaResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("反序列化响应消息失败: %w, 原始消息: %s", err, string(data))
	}
	return &resp, nil
}

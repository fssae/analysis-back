package ioc

import (
	"classroom-analysis/internal/events"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"classroom-analysis/internal/domain"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

// 创建一个handler
type KafkaGroupHandler struct {
	Waiter *events.ResponseWaiter
}

// 实现 sarama.ConsumerGroupHandler 接口
func (h *KafkaGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *KafkaGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// 解析函数

func (h *KafkaGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var resp domain.KafkaResp
		if err := json.Unmarshal(msg.Value, &resp); err == nil {
			// 核心逻辑：分发消息给正在等待的 HTTP 请求
			h.Waiter.Notify(resp.ID, &resp)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

type Writer struct {
	producer sarama.SyncProducer
	topic    string
}

type Reader struct {
	consumer sarama.Consumer
	topic    string
}

func StartKafkaResponseConsumer() {
	brokers := viper.GetStringSlice("kafka.brokers")
	topic := viper.GetString("kafka.ResponseTopic")
	groupId := "face-analyze-service-group"

	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.User = viper.GetString("kafka.username")
	config.Net.SASL.Password = viper.GetString("kafka.password")
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.Handshake = true
	config.Net.TLS.Enable = false // 如果用 SASL_PLAINTEXT
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		panic(fmt.Sprintf("创建消费者组失败: %v", err))
	}

	handler := &KafkaGroupHandler{Waiter: events.GlobalWaiter}

	// 启动全局独立携程
	go func() {
		for {
			ctx := context.Background()
			// 这里的 Consume 会阻塞，直到发生错误或 ctx 取消
			if err := group.Consume(ctx, []string{topic}, handler); err != nil {
				fmt.Printf("消费组报错，3秒后重试: %v\n", err)
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func InitKafkaWriter() *Writer {
	brokers := viper.GetStringSlice("kafka.brokers")
	if len(brokers) == 0 {
		brokers = []string{"82.156.64.69:9092"} // 默认值
	}

	topic := viper.GetString("kafka.RequestTopic")
	if topic == "" {
		panic("Kafka RequestTopic 配置缺失")
	}

	writer, err := NewWriter(brokers, topic)
	if err != nil {
		panic(fmt.Sprintf("初始化Kafka Writer失败: %v", err))
	}
	return writer
}

// Write 发送一条消息到 Kafka
// message: 要发送的消息内容（字符串）
func (w *Writer) Write(message interface{}) error {
	if message == nil {
		return fmt.Errorf("消息不能为空")
	}

	// 将结构体序列化为 JSON 字节数组
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("结构体序列化失败: %v", err)
	}

	// 构建 Kafka 消息对象
	msg := &sarama.ProducerMessage{
		Topic: w.topic,                       // 消息目标主题
		Value: sarama.ByteEncoder(jsonBytes), // 使用字节编码器发送 JSON 数据
	}

	// 发送消息并获取响应（同步）
	_, _, err = w.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送消息到 Kafka 失败: %v", err)
	}
	return nil
}

func NewWriter(brokers []string, topic string) (*Writer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("brokers 列表不能为空")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic 不能为空")
	}

	// 创建默认配置
	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.User = viper.GetString("kafka.username")
	config.Net.SASL.Password = viper.GetString("kafka.password")
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.Handshake = true
	config.Net.TLS.Enable = false // 如果用 SASL_PLAINTEXT
	// 设置生产者成功发送消息后返回响应
	config.Producer.Return.Successes = true
	// 设置重试次数
	config.Producer.Retry.Max = 3
	// 增加生产者重试间隔
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	// 设置更长的超时时间
	config.Producer.Timeout = 30 * time.Second

	// 设置确认模式
	config.Producer.RequiredAcks = sarama.WaitForAll

	// 创建同步生产者实例
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kafka 生产者失败: %v", err)
	}

	// 返回封装好的 Writer 实例
	return &Writer{
		producer: producer,
		topic:    topic,
	}, nil
}

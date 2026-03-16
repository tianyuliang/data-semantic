// Package logic Kafka消费者逻辑
package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

// KafkaConsumer Kafka消费者结构
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handlers map[string]MessageHandler
}

// MessageHandler 消息处理接口
type MessageHandler interface {
	Handle(ctx context.Context, message *sarama.ConsumerMessage) error
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(brokers []string, groupID string, topics []string) (*KafkaConsumer, error) {
	return NewKafkaConsumerWithAuth(brokers, groupID, topics, "", "")
}

// NewKafkaConsumerWithAuth 创建带认证的Kafka消费者
func NewKafkaConsumerWithAuth(brokers []string, groupID string, topics []string, username, password string) (*KafkaConsumer, error) {
	// Kafka 配置
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// 重试配置
	config.Consumer.Retry.Backoff = 3 * time.Second // 重试间隔

	// 关键配置：防止会话超时导致消费停止
	config.Consumer.Group.Session.Timeout = 30 * time.Second     // 会话超时
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second   // 心跳间隔
	config.Consumer.MaxProcessingTime = 30 * time.Second        // 单条消息最大处理时间

	// SASL 认证配置
	if username != "" && password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = username
		config.Net.SASL.Password = password
	}

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("创建消费者失败: %w", err)
	}

	kafkaConsumer := &KafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]MessageHandler),
	}

	return kafkaConsumer, nil
}

// RegisterHandler 注册消息处理器
func (kc *KafkaConsumer) RegisterHandler(topic string, handler MessageHandler) {
	kc.handlers[topic] = handler
}

// Start 启动消费者
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	// 订阅主题
	topics := make([]string, 0, len(kc.handlers))
	for topic := range kc.handlers {
		topics = append(topics, topic)
	}

	logx.Infof("启动Kafka消费者，订阅主题: %v", topics)

	// 启动消费
	if err := kc.consumer.Consume(ctx, topics, kc); err != nil {
		return fmt.Errorf("消费失败: %w", err)
	}

	return nil
}

// Setup 会话开始时调用
func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	logx.Info("Kafka消费者会话开始")
	return nil
}

// Cleanup 会话结束时调用
func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	logx.Info("Kafka消费者会话结束")
	return nil
}

// ConsumeClaim 消费消息 (必须实现 ConsumerGroupHandler 接口)
func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// claim.Messages() 返回一个消息通道，持续消费直到通道关闭
	for msg := range claim.Messages() {
		topic := msg.Topic
		handler, ok := kc.handlers[topic]
		if !ok {
			// 未找到处理器：记录错误并确认消息（避免无限重试）
			logx.Errorf("未找到topic处理器，跳过消息: topic=%s partition=%d offset=%d key=%s",
				topic, msg.Partition, msg.Offset, string(msg.Key))
			// 确认消息以避免无限重试
			session.MarkMessage(msg, "")
			continue
		}

		// 处理消息：只有成功才标记消息已处理
		if err := handler.Handle(session.Context(), msg); err != nil {
			logx.Errorf("处理消息失败，等待重试: topic=%s partition=%d offset=%d error=%v",
				topic, msg.Partition, msg.Offset, err)
			// 处理失败时不确认消息，让 Kafka 重投递
			continue
		}

		// 标记消息已处理 (提交偏移量)
		session.MarkMessage(msg, "")
	}

	return nil
}

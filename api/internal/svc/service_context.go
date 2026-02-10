// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/aiservice"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/time/rate"
)

type ServiceContext struct {
	Config       config.Config
	DB           sqlx.SqlConn       // 数据库连接
	Kafka        sarama.SyncProducer // Kafka 生产者
	Redis        *redis.Redis       // Redis 客户端
	AIClient     *aiservice.Client  // AI 服务客户端
	rateLimiters sync.Map           // formViewId -> *rate.Limiter (限流器缓存)
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	db := sqlx.NewMysql(c.DB.Default.DataSource())

	// 初始化 Kafka Producer
	kafkaProducer, err := initKafkaProducer(c)
	if err != nil {
		log.Printf("初始化 Kafka Producer 失败: %v", err)
	}

	// 初始化 Redis
	redisClient := initRedis(c)

	// 初始化 AI 服务客户端
	timeout := time.Duration(c.AIService.TimeoutSeconds) * time.Second
	aiClient := aiservice.NewClient(c.AIService.URL, timeout)

	return &ServiceContext{
		Config:   c,
		DB:       db,
		Kafka:    kafkaProducer,
		Redis:    redisClient,
		AIClient: aiClient,
	}
}

// initKafkaProducer 初始化 Kafka 同步生产者
func initKafkaProducer(c config.Config) (sarama.SyncProducer, error) {
	if len(c.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("Kafka brokers 未配置")
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有副本确认
	config.Producer.Retry.Max = 3                   // 重试次数
	config.Producer.Return.Successes = true          // 启用成功通道

	producer, err := sarama.NewSyncProducer(c.Kafka.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kafka Producer 失败: %w", err)
	}

	return producer, nil
}

// initRedis 初始化 Redis 客户端
func initRedis(c config.Config) *redis.Redis {
	return redis.MustNewRedis(redis.RedisConf{
		Host: c.Redis.Addr(),
		Pass: c.Redis.Password,
		Type: redis.NodeType,
	})
}

// GetRateLimiter 获取或创建指定 formViewId 的限流器
// 使用 1 秒窗口，允许 1 次请求
func (s *ServiceContext) GetRateLimiter(formViewId string) *rate.Limiter {
	// 尝试从缓存中获取
	if limiter, ok := s.rateLimiters.Load(formViewId); ok {
		return limiter.(*rate.Limiter)
	}

	// 创建新的限流器：1 秒内最多 1 次请求
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	// 存入缓存（如果已存在则使用已存在的）
	actual, _ := s.rateLimiters.LoadOrStore(formViewId, limiter)
	return actual.(*rate.Limiter)
}

// SendKafkaMessage 发送 Kafka 消息
func (s *ServiceContext) SendKafkaMessage(topic string, message map[string]interface{}) error {
	if s.Kafka == nil {
		return fmt.Errorf("Kafka Producer 未初始化")
	}

	// 将消息序列化为 JSON
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 创建 Kafka 消息
	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonBytes),
	}

	// 发送消息
	partition, offset, err := s.Kafka.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("发送 Kafka 消息失败: %w", err)
	}

	log.Printf("Kafka 消息发送成功: topic=%s, partition=%d, offset=%d", topic, partition, offset)
	return nil
}

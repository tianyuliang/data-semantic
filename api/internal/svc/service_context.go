// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	DB            sqlx.SqlConn       // 数据库连接
	Kafka         sarama.SyncProducer // Kafka 生产者
	Redis         *redis.Redis       // Redis 客户端
	HttpClient    *http.Client       // HTTP 客户端（用于调用 AI 服务）
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

	// 初始化 HTTP 客户端（用于调用 AI 服务）
	timeout := time.Duration(c.AIService.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second // 默认 10 秒
	}
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &ServiceContext{
		Config:     c,
		DB:         db,
		Kafka:      kafkaProducer,
		Redis:      redisClient,
		HttpClient: httpClient,
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

// AIServiceResponse AI 服务响应结构
type AIServiceResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	MessageID string `json:"message_id"`
}

// CallAIService 调用 AI 服务 HTTP API
func (s *ServiceContext) CallAIService(requestType string, requestBody map[string]interface{}) (*AIServiceResponse, error) {
	if s.Config.AIService.URL == "" {
		return nil, fmt.Errorf("AI 服务 URL 未配置")
	}

	// 构建 JSON 请求体
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/af-sailor-agent/v1/data_understand/view_semantic_and_business_analysis", s.Config.AIService.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI 服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var aiResponse AIServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return nil, fmt.Errorf("解析 AI 服务响应失败: %w", err)
	}

	return &aiResponse, nil
}

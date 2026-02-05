// Package data_understanding Kafka消费者测试
package data_understanding

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaConsumer Mock Kafka消费者
type MockKafkaConsumer struct {
	mock.Mock
}

// TestCheckMessageId 测试消息去重
func TestCheckMessageId(t *testing.T) {
	t.Skip("需要数据库连接")

	handler := NewDataUnderstandingHandler()

	ctx := context.Background()
	messageId := "test-message-id"

	exists, err := handler.checkMessageId(ctx, messageId)
	assert.NoError(t, err)
	assert.False(t, exists) // 临时返回 false
}

// TestProcessSuccessResponse 测试成功响应处理
func TestProcessSuccessResponse(t *testing.T) {
	t.Skip("需要数据库连接")

	handler := NewDataUnderstandingHandler()
	ctx := context.Background()

	messageId := "test-message-id"
	formViewId := "test-form-view-id"
	msg := map[string]interface{}{
		"message_id":       messageId,
		"form_view_id":     formViewId,
		"request_time":     "2026-02-04T10:00:00Z",
		"business_objects": []interface{}{},
	}

	err := handler.processSuccessResponse(ctx, messageId, formViewId, msg)
	assert.NoError(t, err)
}

// TestRecordFailure 测试失败记录
func TestRecordFailure(t *testing.T) {
	handler := NewDataUnderstandingHandler()
	ctx := context.Background()

	msg := map[string]interface{}{
		"message_id":   "test-message-id",
		"form_view_id": "test-form-view-id",
	}

	err := handler.recordFailure(ctx, msg, assert.AnError)
	assert.NoError(t, err) // recordFailure 总是返回 nil
}

// MockMessage 模拟消息
type MockMessage struct {
	sarama.ConsumerMessage
	value []byte
}

// NewMockMessage 创建模拟消息
func NewMockMessage(topic string, value []byte) *MockMessage {
	return &MockMessage{
		value: value,
	}
}

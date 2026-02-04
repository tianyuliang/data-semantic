// Package data_understanding Kafka消息处理逻辑
package data_understanding

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// DataUnderstandingHandler 数据理解消息处理器
type DataUnderstandingHandler struct {
	// TODO: 添加依赖 (数据库连接、Redis等)
}

// NewDataUnderstandingHandler 创建消息处理器
func NewDataUnderstandingHandler() *DataUnderstandingHandler {
	return &DataUnderstandingHandler{}
}

// Handle 处理Kafka消息
func (h *DataUnderstandingHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	logx.Infof("收到Kafka消息: topic=%s partition=%d offset=%d",
		message.Topic, message.Partition, message.Offset)

	// 解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		logx.Errorf("解析消息失败: %v", err)
		return h.recordFailure(ctx, msg, err)
	}

	// 提取字段
	messageId, _ := msg["message_id"].(string)
	formViewId, _ := msg["form_view_id"].(string)

	// 去重检查
	if exists, err := h.checkMessageId(ctx, messageId); err != nil {
		logx.Errorf("检查消息去重失败: %v", err)
		return err
	} else if exists {
		logx.Infof("消息已处理，跳过: message_id=%s", messageId)
		return nil
	}

	// 处理成功响应
	if err := h.processSuccessResponse(ctx, messageId, formViewId, msg); err != nil {
		return err
	}

	return nil
}

// checkMessageId 检查消息是否已处理
func (h *DataUnderstandingHandler) checkMessageId(ctx context.Context, messageId string) (bool, error) {
	// TODO: 查询 t_kafka_message_log 表
	// SELECT COUNT(*) FROM t_kafka_message_log WHERE message_id = ?
	// return count > 0, nil
	return false, nil
}

// processSuccessResponse 处理成功响应
func (h *DataUnderstandingHandler) processSuccessResponse(ctx context.Context, messageId, formViewId string, msg map[string]interface{}) error {
	// 1. 记录消息处理日志
	logId := uuid.New().String()
	// TODO: INSERT INTO t_kafka_message_log (id, message_id, form_view_id, status=1)

	// 2. 解析AI返回的数据并保存到临时表
	// TODO: 从 msg 中提取 business_objects 和 fields
	// TODO: 保存到 t_form_view_info_temp, t_form_view_field_info_temp
	// TODO: 保存到 t_business_object_temp, t_business_object_attributes_temp

	// 3. 更新 form_view 状态为 2（待确认）
	// TODO: UPDATE form_view SET understand_status = 2 WHERE id = ?

	logx.Infof("处理成功响应: message_id=%s, form_view_id=%s, log_id=%s",
		messageId, formViewId, logId)

	return nil
}

// recordFailure 记录处理失败
func (h *DataUnderstandingHandler) recordFailure(ctx context.Context, msg map[string]interface{}, err error) error {
	// 记录结构化日志
	logData := map[string]interface{}{
		"timestamp": msg["request_time"],
		"level":     "error",
		"message":   "Kafka message processing failed",
		"context": map[string]interface{}{
			"message_id":   msg["message_id"],
			"form_view_id": msg["form_view_id"],
		},
		"error": map[string]interface{}{
			"type":    "AIAnalysisError",
			"message": err.Error(),
		},
	}
	logJson, _ := json.Marshal(logData)
	logx.Error(string(logJson))

	// TODO: 记录到 t_kafka_message_log (status=2, error_msg=...)

	return nil
}

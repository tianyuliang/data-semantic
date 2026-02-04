// Package kafka_message_log Kafka消息处理记录Model
package kafka_message_log

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewKafkaMessageLogModel 创建KafkaMessageLogModel实例
func NewKafkaMessageLogModel(db *sqlx.Tx) *KafkaMessageLogModelImpl {
	return &KafkaMessageLogModelImpl{db: db}
}

// KafkaMessageLogModelImpl KafkaMessageLogModel实现
type KafkaMessageLogModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入Kafka消息处理记录
func (m *KafkaMessageLogModelImpl) Insert(ctx context.Context, data *KafkaMessageLog) (*KafkaMessageLog, error) {
	query := `INSERT INTO t_kafka_message_log (id, message_id, form_view_id, processed_at, status, error_msg)
	           VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.MessageId, data.FormViewId, data.ProcessedAt, data.Status, data.ErrorMsg)
	if err != nil {
		return nil, fmt.Errorf("insert kafka message log failed: %w", err)
	}
	return data, nil
}

// FindOneByMessageId 根据消息ID查询记录
func (m *KafkaMessageLogModelImpl) FindOneByMessageId(ctx context.Context, messageId string) (*KafkaMessageLog, error) {
	var resp KafkaMessageLog
	query := `SELECT id, message_id, form_view_id, processed_at, status, error_msg
	           FROM t_kafka_message_log
	           WHERE message_id = ? LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, messageId)
	if err != nil {
		return nil, fmt.Errorf("find kafka message log by message_id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新处理状态
func (m *KafkaMessageLogModelImpl) Update(ctx context.Context, data *KafkaMessageLog) error {
	query := `UPDATE t_kafka_message_log
	           SET status = ?, error_msg = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.Status, data.ErrorMsg, data.Id)
	if err != nil {
		return fmt.Errorf("update kafka message log failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *KafkaMessageLogModelImpl) WithTx(tx interface{}) KafkaMessageLogModel {
	return &KafkaMessageLogModelImpl{db: tx.(*sqlx.Tx)}
}

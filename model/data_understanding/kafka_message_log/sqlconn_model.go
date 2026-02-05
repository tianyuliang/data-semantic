// Package kafka_message_log Kafka消息处理记录Model (SqlConn实现)
package kafka_message_log

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewKafkaMessageLogModelSqlConn 创建KafkaMessageLogModelSqlConn实例
func NewKafkaMessageLogModelSqlConn(conn sqlx.SqlConn) *KafkaMessageLogModelSqlConn {
	return &KafkaMessageLogModelSqlConn{conn: conn}
}

// NewKafkaMessageLogModelSession 创建KafkaMessageLogModelSqlConn实例 (使用 Session)
func NewKafkaMessageLogModelSession(session sqlx.Session) *KafkaMessageLogModelSqlConn {
	return &KafkaMessageLogModelSqlConn{conn: session}
}

// KafkaMessageLogModelSqlConn KafkaMessageLogModel实现 (基于 go-zero SqlConn)
type KafkaMessageLogModelSqlConn struct {
	conn sqlx.Session
}

// Insert 插入Kafka消息处理记录
func (m *KafkaMessageLogModelSqlConn) Insert(ctx context.Context, data *KafkaMessageLog) (*KafkaMessageLog, error) {
	query := `INSERT INTO t_kafka_message_log (id, message_id, form_view_id, processed_at, status, error_msg)
	           VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, data.Id, data.MessageId, data.FormViewId, data.ProcessedAt, data.Status, data.ErrorMsg)
	if err != nil {
		return nil, fmt.Errorf("insert kafka message log failed: %w", err)
	}
	return data, nil
}

// FindOneByMessageId 根据消息ID查询记录
func (m *KafkaMessageLogModelSqlConn) FindOneByMessageId(ctx context.Context, messageId string) (*KafkaMessageLog, error) {
	var resp KafkaMessageLog
	query := `SELECT id, message_id, form_view_id, processed_at, status, error_msg
	           FROM t_kafka_message_log
	           WHERE message_id = ? LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, messageId)
	if err != nil {
		return nil, fmt.Errorf("find kafka message log by message_id failed: %w", err)
	}
	return &resp, nil
}

// ExistsByMessageId 检查消息ID是否已存在
func (m *KafkaMessageLogModelSqlConn) ExistsByMessageId(ctx context.Context, messageId string) (bool, error) {
	var count int64
	query := `SELECT COUNT(*) FROM t_kafka_message_log WHERE message_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, messageId)
	if err != nil {
		return false, fmt.Errorf("exists kafka message log by message_id failed: %w", err)
	}
	return count > 0, nil
}

// InsertSuccess 插入成功处理记录
func (m *KafkaMessageLogModelSqlConn) InsertSuccess(ctx context.Context, messageId, formViewId string) (*KafkaMessageLog, error) {
	data := &KafkaMessageLog{
		Id:          uuid.New().String(),
		MessageId:   messageId,
		FormViewId:  formViewId,
		ProcessedAt: time.Now(),
		Status:      1, // 处理成功
	}
	return m.Insert(ctx, data)
}

// InsertFailure 插入失败处理记录
func (m *KafkaMessageLogModelSqlConn) InsertFailure(ctx context.Context, messageId, formViewId string, errMsg string) (*KafkaMessageLog, error) {
	data := &KafkaMessageLog{
		Id:          uuid.New().String(),
		MessageId:   messageId,
		FormViewId:  formViewId,
		ProcessedAt: time.Now(),
		Status:      2, // 处理失败
		ErrorMsg:    &errMsg,
	}
	return m.Insert(ctx, data)
}

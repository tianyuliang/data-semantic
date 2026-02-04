// Package kafka_message_log Kafka消息处理记录Model
package kafka_message_log

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

// 测试数据库连接字符串
const testDSN = "root:password@tcp(localhost:3306)/test_db?parseTime=true"

// TestKafkaMessageLogModel_Insert 测试插入记录
func TestKafkaMessageLogModel_Insert(t *testing.T) {
	// 跳过集成测试（需要数据库）
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewKafkaMessageLogModel(tx)

	data := &KafkaMessageLog{
		Id:         "test-id",
		MessageId:  "test-message-id",
		FormViewId: "test-form-view-id",
		Status:     StatusProcessed,
	}

	result, err := model.Insert(context.Background(), data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestKafkaMessageLogModel_FindOneByMessageId 测试根据消息ID查询
func TestKafkaMessageLogModel_FindOneByMessageId(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewKafkaMessageLogModel(tx)

	result, err := model.FindOneByMessageId(context.Background(), "test-message-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Package form_view_info_temp 库表信息临时表Model测试
package form_view_info_temp

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// 测试数据库连接字符串
const testDSN = "root:password@tcp(localhost:3306)/test_db?parseTime=true"

// TestFormViewInfoTempModel_Insert 测试插入记录
func TestFormViewInfoTempModel_Insert(t *testing.T) {
	// 跳过集成测试（需要数据库）
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewInfoTempModel(tx)

	data := &FormViewInfoTemp{
		Id:                "test-id",
		FormViewId:        "test-form-view-id",
		Version:           InitialVersion,
		TableBusinessName: stringPtr("测试业务表"),
		TableDescription:  stringPtr("测试描述"),
	}

	result, err := model.Insert(context.Background(), data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestFormViewInfoTempModel_FindOneByFormViewAndVersion 测试根据form_view_id和version查询
func TestFormViewInfoTempModel_FindOneByFormViewAndVersion(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewInfoTempModel(tx)

	result, err := model.FindOneByFormViewAndVersion(context.Background(), "test-form-view-id", InitialVersion)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestFormViewInfoTempModel_FindLatestByFormViewId 测试查询最新版本
func TestFormViewInfoTempModel_FindLatestByFormViewId(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewInfoTempModel(tx)

	result, err := model.FindLatestByFormViewId(context.Background(), "test-form-view-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}

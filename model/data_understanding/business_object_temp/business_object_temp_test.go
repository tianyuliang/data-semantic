// Package business_object_temp 业务对象临时表Model测试
package business_object_temp

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

// 测试数据库连接字符串
const testDSN = "root:password@tcp(localhost:3306)/test_db?parseTime=true"

// TestBusinessObjectTempModel_Insert 测试插入记录
func TestBusinessObjectTempModel_Insert(t *testing.T) {
	// 跳过集成测试（需要数据库）
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectTempModel(tx)

	data := &BusinessObjectTemp{
		Id:         "test-id",
		FormViewId: "test-form-view-id",
		Version:    InitialVersion,
		ObjectName: "测试业务对象",
	}

	result, err := model.Insert(context.Background(), data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestBusinessObjectTempModel_FindOneByFormViewAndVersion 测试根据form_view_id和version查询
func TestBusinessObjectTempModel_FindOneByFormViewAndVersion(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectTempModel(tx)

	result, err := model.FindByFormViewAndVersion(context.Background(), "test-form-view-id", InitialVersion)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestBusinessObjectTempModel_FindOneById 测试根据id查询
func TestBusinessObjectTempModel_FindOneById(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectTempModel(tx)

	result, err := model.FindOneById(context.Background(), "test-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

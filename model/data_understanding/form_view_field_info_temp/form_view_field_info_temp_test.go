// Package form_view_field_info_temp 库表字段信息临时表Model测试
package form_view_field_info_temp

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// 测试数据库连接字符串
const testDSN = "root:password@tcp(localhost:3306)/test_db?parseTime=true"

// TestFormViewFieldInfoTempModel_Insert 测试插入记录
func TestFormViewFieldInfoTempModel_Insert(t *testing.T) {
	// 跳过集成测试（需要数据库）
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewFieldInfoTempModel(tx)

	role := int8(FieldRoleBusinessKey)
	data := &FormViewFieldInfoTemp{
		Id:                "test-id",
		FormViewId:        "test-form-view-id",
		FormViewFieldId:   "test-field-id",
		Version:           10,
		FieldBusinessName: stringPtr("测试字段"),
		FieldRole:         &role,
		FieldDescription:  stringPtr("测试字段描述"),
	}

	result, err := model.Insert(context.Background(), data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestFormViewFieldInfoTempModel_FindByFormViewAndVersion 测试根据form_view_id和version查询字段列表
func TestFormViewFieldInfoTempModel_FindByFormViewAndVersion(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewFieldInfoTempModel(tx)

	result, err := model.FindByFormViewAndVersion(context.Background(), "test-form-view-id", 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestFormViewFieldInfoTempModel_FindOneByFormFieldId 测试根据form_view_field_id查询
func TestFormViewFieldInfoTempModel_FindOneByFormFieldId(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	model := NewFormViewFieldInfoTempModel(tx)

	result, err := model.FindOneByFormFieldId(context.Background(), "test-field-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}

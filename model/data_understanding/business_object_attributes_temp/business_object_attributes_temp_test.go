// Package business_object_attributes_temp 业务对象属性临时表Model测试
package business_object_attributes_temp

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

// 测试数据库连接字符串
const testDSN = "root:password@tcp(localhost:3306)/test_db?parseTime=true"

// TestBusinessObjectAttributesTempModel_Insert 测试插入记录
func TestBusinessObjectAttributesTempModel_Insert(t *testing.T) {
	// 跳过集成测试（需要数据库）
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectAttributesTempModel(tx)

	data := &BusinessObjectAttributesTemp{
		Id:               "test-attr-id",
		FormViewId:       "test-form-view-id",
		BusinessObjectId: "test-object-id",
		Version:          10,
		FormViewFieldId:  "test-field-id",
		AttrName:         "测试属性",
	}

	result, err := model.Insert(context.Background(), data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestBusinessObjectAttributesTempModel_FindByBusinessObjectId 测试根据business_object_id查询属性列表
func TestBusinessObjectAttributesTempModel_FindByBusinessObjectId(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectAttributesTempModel(tx)

	result, err := model.FindByBusinessObjectId(context.Background(), "test-object-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestBusinessObjectAttributesTempModel_FindByFormViewAndVersion 测试根据form_view_id和version查询所有属性
func TestBusinessObjectAttributesTempModel_FindByFormViewAndVersion(t *testing.T) {
	t.Skip("需要数据库连接")

	db, err := sqlx.Connect("mysql", testDSN)
	assert.NoError(t, err)
	defer db.Close()

	tx, err := db.Beginx()
	assert.NoError(t, err)
	defer tx.Rollback()

	model := NewBusinessObjectAttributesTempModel(tx)

	result, err := model.FindByFormViewAndVersion(context.Background(), "test-form-view-id", 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

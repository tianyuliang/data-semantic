// Package business_object_attributes 业务对象属性正式表Model测试
package business_object_attributes

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestBusinessObjectAttributesModel_Insert(t *testing.T) {
	// TODO: 实现插入测试
	// tx := testDB.Beginx()
	// defer tx.Rollback()
	// model := NewBusinessObjectAttributesModel(tx)
	// data := &BusinessObjectAttributes{
	//     Id:               "test-id",
	//     FormViewId:       "test-form-view-id",
	//     BusinessObjectId: "test-business-object-id",
	//     FormViewFieldId:  "test-field-id",
	//     AttrName:         "测试属性",
	// }
	// result, err := model.Insert(context.Background(), data)
	// assert.NoError(t, err)
	// assert.NotNil(t, result)
}

func TestBusinessObjectAttributesModel_FindByBusinessObjectId(t *testing.T) {
	// TODO: 实现按业务对象查询测试
}

func TestBusinessObjectAttributesModel_FindByFormViewId(t *testing.T) {
	// TODO: 实现按视图查询测试
}

func TestBusinessObjectAttributesModel_Update(t *testing.T) {
	// TODO: 实现更新测试
}

func TestBusinessObjectAttributesModel_Delete(t *testing.T) {
	// TODO: 实现删除测试
}

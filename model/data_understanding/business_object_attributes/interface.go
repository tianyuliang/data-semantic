// Package business_object_attributes 业务对象属性正式表Model
package business_object_attributes

import "context"

// BusinessObjectAttributesModel 业务对象属性正式表Model接口
type BusinessObjectAttributesModel interface {
	// Insert 插入业务对象属性记录
	Insert(ctx context.Context, data *BusinessObjectAttributes) (*BusinessObjectAttributes, error)

	// FindByBusinessObjectId 根据business_object_id查询属性列表
	FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributes, error)

	// FindByFormViewId 根据form_view_id查询所有属性
	FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObjectAttributes, error)

	// Update 更新属性
	Update(ctx context.Context, data *BusinessObjectAttributes) error

	// Delete 逻辑删除属性
	Delete(ctx context.Context, id string) error

	// DeleteByFormViewId 根据form_view_id删除所有属性
	DeleteByFormViewId(ctx context.Context, formViewId string) error

	// BatchInsertFromTemp 从临时表批量插入属性
	BatchInsertFromTemp(ctx context.Context, formViewId string, version int) (int, error)

	// WithTx 设置事务
	WithTx(tx interface{}) BusinessObjectAttributesModel
}

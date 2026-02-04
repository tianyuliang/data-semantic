// Package business_object_attributes_temp 业务对象属性临时表Model
package business_object_attributes_temp

import "context"

// BusinessObjectAttributesTempModel 业务对象属性临时表Model接口
type BusinessObjectAttributesTempModel interface {
	// Insert 插入业务对象属性记录
	Insert(ctx context.Context, data *BusinessObjectAttributesTemp) (*BusinessObjectAttributesTemp, error)

	// FindByBusinessObjectId 根据business_object_id查询属性列表
	FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributesTemp, error)

	// FindByFormViewAndVersion 根据form_view_id和version查询所有属性
	FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectAttributesTemp, error)

	// Update 更新属性名称
	Update(ctx context.Context, data *BusinessObjectAttributesTemp) error

	// WithTx 设置事务
	WithTx(tx interface{}) BusinessObjectAttributesTempModel
}

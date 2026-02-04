// Package business_object_temp 业务对象临时表Model
package business_object_temp

import "context"

// BusinessObjectTempModel 业务对象临时表Model接口
type BusinessObjectTempModel interface {
	// Insert 插入业务对象记录
	Insert(ctx context.Context, data *BusinessObjectTemp) (*BusinessObjectTemp, error)

	// FindOneByFormViewAndVersion 根据form_view_id和version查询业务对象列表
	FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectTemp, error)

	// FindOneById 根据id查询业务对象
	FindOneById(ctx context.Context, id string) (*BusinessObjectTemp, error)

	// Update 更新业务对象名称
	Update(ctx context.Context, data *BusinessObjectTemp) error

	// WithTx 设置事务
	WithTx(tx interface{}) BusinessObjectTempModel
}

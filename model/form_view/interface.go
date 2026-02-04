// Package form_view 库表视图Model
package form_view

import "context"

// FormViewModel 库表视图Model接口
type FormViewModel interface {
	// FindOneById 根据id查询库表视图
	FindOneById(ctx context.Context, id string) (*FormView, error)

	// UpdateUnderstandStatus 更新理解状态
	UpdateUnderstandStatus(ctx context.Context, id string, status int8) error

	// WithTx 设置事务
	WithTx(tx interface{}) FormViewModel
}

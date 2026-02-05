// Package business_object_temp 业务对象临时表Model
package business_object_temp

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewBusinessObjectTempModel 创建BusinessObjectTempModel实例
func NewBusinessObjectTempModel(db *sqlx.Tx) *BusinessObjectTempModelImpl {
	return &BusinessObjectTempModelImpl{db: db}
}

// BusinessObjectTempModelImpl BusinessObjectTempModel实现
type BusinessObjectTempModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入业务对象记录
func (m *BusinessObjectTempModelImpl) Insert(ctx context.Context, data *BusinessObjectTemp) (*BusinessObjectTemp, error) {
	query := `INSERT INTO t_business_object_temp (id, form_view_id, user_id, version, object_name)
	           VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.FormViewId, data.UserId, data.Version, data.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("insert business_object_temp failed: %w", err)
	}
	return data, nil
}

// FindOneByFormViewAndVersion 根据form_view_id和version查询业务对象列表
func (m *BusinessObjectTempModelImpl) FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectTemp, error) {
	var resp []*BusinessObjectTemp
	query := `SELECT id, form_view_id, user_id, version, object_name, created_at, updated_at, deleted_at
	           FROM t_business_object_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find business_object_temp by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// FindOneById 根据id查询业务对象
func (m *BusinessObjectTempModelImpl) FindOneById(ctx context.Context, id string) (*BusinessObjectTemp, error) {
	var resp BusinessObjectTemp
	query := `SELECT id, form_view_id, user_id, version, object_name, created_at, updated_at, deleted_at
	           FROM t_business_object_temp
	           WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find business_object_temp by id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新业务对象名称
func (m *BusinessObjectTempModelImpl) Update(ctx context.Context, data *BusinessObjectTemp) error {
	query := `UPDATE t_business_object_temp
	           SET object_name = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.ObjectName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_temp failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectTempModelImpl) WithTx(tx interface{}) BusinessObjectTempModel {
	return &BusinessObjectTempModelImpl{db: tx.(*sqlx.Tx)}
}

// DeleteByFormViewId 根据form_view_id删除所有业务对象
func (m *BusinessObjectTempModelImpl) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object_temp SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.db.ExecContext(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object_temp by form_view_id failed: %w", err)
	}
	return nil
}

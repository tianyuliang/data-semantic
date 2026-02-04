// Package business_object 业务对象正式表Model
package business_object

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewBusinessObjectModel 创建BusinessObjectModel实例
func NewBusinessObjectModel(db *sqlx.Tx) *BusinessObjectModelImpl {
	return &BusinessObjectModelImpl{db: db}
}

// BusinessObjectModelImpl BusinessObjectModel实现
type BusinessObjectModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入业务对象记录
func (m *BusinessObjectModelImpl) Insert(ctx context.Context, data *BusinessObject) (*BusinessObject, error) {
	query := `INSERT INTO t_business_object (id, object_name, object_type, form_view_id, status)
	           VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.ObjectName, data.ObjectType, data.FormViewId, data.Status)
	if err != nil {
		return nil, fmt.Errorf("insert business_object failed: %w", err)
	}
	return data, nil
}

// FindByFormViewId 根据form_view_id查询业务对象列表
func (m *BusinessObjectModelImpl) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObject, error) {
	var resp []*BusinessObject
	query := `SELECT id, object_name, object_type, form_view_id, status, created_at, updated_at, deleted_at
	           FROM t_business_object
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find business_object by form_view_id failed: %w", err)
	}
	return resp, nil
}

// FindOneById 根据id查询业务对象
func (m *BusinessObjectModelImpl) FindOneById(ctx context.Context, id string) (*BusinessObject, error) {
	var resp BusinessObject
	query := `SELECT id, object_name, object_type, form_view_id, status, created_at, updated_at, deleted_at
	           FROM t_business_object
	           WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find business_object by id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新业务对象
func (m *BusinessObjectModelImpl) Update(ctx context.Context, data *BusinessObject) error {
	query := `UPDATE t_business_object
	           SET object_name = ?, object_type = ?, status = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.ObjectName, data.ObjectType, data.Status, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object failed: %w", err)
	}
	return nil
}

// Delete 逻辑删除业务对象
func (m *BusinessObjectModelImpl) Delete(ctx context.Context, id string) error {
	query := `UPDATE t_business_object SET deleted_at = NOW(3) WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete business_object failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectModelImpl) WithTx(tx interface{}) BusinessObjectModel {
	return &BusinessObjectModelImpl{db: tx.(*sqlx.Tx)}
}

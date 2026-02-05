// Package business_object_attributes 业务对象属性正式表Model
package business_object_attributes

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewBusinessObjectAttributesModel 创建BusinessObjectAttributesModel实例
func NewBusinessObjectAttributesModel(db *sqlx.Tx) *BusinessObjectAttributesModelImpl {
	return &BusinessObjectAttributesModelImpl{db: db}
}

// BusinessObjectAttributesModelImpl BusinessObjectAttributesModel实现
type BusinessObjectAttributesModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入业务对象属性记录
func (m *BusinessObjectAttributesModelImpl) Insert(ctx context.Context, data *BusinessObjectAttributes) (*BusinessObjectAttributes, error) {
	query := `INSERT INTO t_business_object_attributes (id, form_view_id, business_object_id, form_view_field_id, attr_name)
	           VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.FormViewId, data.BusinessObjectId, data.FormViewFieldId, data.AttrName)
	if err != nil {
		return nil, fmt.Errorf("insert business_object_attributes failed: %w", err)
	}
	return data, nil
}

// FindByBusinessObjectId 根据business_object_id查询属性列表
func (m *BusinessObjectAttributesModelImpl) FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributes, error) {
	var resp []*BusinessObjectAttributes
	query := `SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes
	           WHERE business_object_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes by business_object_id failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewId 根据form_view_id查询所有属性
func (m *BusinessObjectAttributesModelImpl) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObjectAttributes, error) {
	var resp []*BusinessObjectAttributes
	query := `SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes by form_view_id failed: %w", err)
	}
	return resp, nil
}

// Update 更新属性
func (m *BusinessObjectAttributesModelImpl) Update(ctx context.Context, data *BusinessObjectAttributes) error {
	query := `UPDATE t_business_object_attributes
	           SET attr_name = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.AttrName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_attributes failed: %w", err)
	}
	return nil
}

// Delete 逻辑删除属性
func (m *BusinessObjectAttributesModelImpl) Delete(ctx context.Context, id string) error {
	query := `UPDATE t_business_object_attributes SET deleted_at = NOW(3) WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete business_object_attributes failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectAttributesModelImpl) WithTx(tx interface{}) BusinessObjectAttributesModel {
	return &BusinessObjectAttributesModelImpl{db: tx.(*sqlx.Tx)}
}

// DeleteByFormViewId 根据form_view_id删除所有属性
func (m *BusinessObjectAttributesModelImpl) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object_attributes SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.db.ExecContext(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object_attributes by form_view_id failed: %w", err)
	}
	return nil
}

// BatchInsertFromTemp 从临时表批量插入属性
func (m *BusinessObjectAttributesModelImpl) BatchInsertFromTemp(ctx context.Context, formViewId string, version int) (int, error) {
	query := `INSERT INTO t_business_object_attributes (id, form_view_id, business_object_id, form_view_field_id, attr_name)
	           SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name
	           FROM t_business_object_attributes_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL`
	result, err := m.db.ExecContext(ctx, query, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("batch insert business_object_attributes from temp failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

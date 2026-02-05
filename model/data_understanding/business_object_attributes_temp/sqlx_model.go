// Package business_object_attributes_temp 业务对象属性临时表Model
package business_object_attributes_temp

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewBusinessObjectAttributesTempModel 创建BusinessObjectAttributesTempModel实例
func NewBusinessObjectAttributesTempModel(db *sqlx.Tx) *BusinessObjectAttributesTempModelImpl {
	return &BusinessObjectAttributesTempModelImpl{db: db}
}

// BusinessObjectAttributesTempModelImpl BusinessObjectAttributesTempModel实现
type BusinessObjectAttributesTempModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入业务对象属性记录
func (m *BusinessObjectAttributesTempModelImpl) Insert(ctx context.Context, data *BusinessObjectAttributesTemp) (*BusinessObjectAttributesTemp, error) {
	query := `INSERT INTO t_business_object_attributes_temp (id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name)
	           VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.FormViewId, data.BusinessObjectId, data.UserId, data.Version, data.FormViewFieldId, data.AttrName)
	if err != nil {
		return nil, fmt.Errorf("insert business_object_attributes_temp failed: %w", err)
	}
	return data, nil
}

// FindByBusinessObjectId 根据business_object_id查询属性列表
func (m *BusinessObjectAttributesTempModelImpl) FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributesTemp, error) {
	var resp []*BusinessObjectAttributesTemp
	query := `SELECT id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes_temp
	           WHERE business_object_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp by business_object_id failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewAndVersion 根据form_view_id和version查询所有属性
func (m *BusinessObjectAttributesTempModelImpl) FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectAttributesTemp, error) {
	var resp []*BusinessObjectAttributesTemp
	query := `SELECT id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// Update 更新属性名称
func (m *BusinessObjectAttributesTempModelImpl) Update(ctx context.Context, data *BusinessObjectAttributesTemp) error {
	query := `UPDATE t_business_object_attributes_temp
	           SET attr_name = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.AttrName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_attributes_temp failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectAttributesTempModelImpl) WithTx(tx interface{}) BusinessObjectAttributesTempModel {
	return &BusinessObjectAttributesTempModelImpl{db: tx.(*sqlx.Tx)}
}

// DeleteByFormViewId 根据form_view_id删除所有属性
func (m *BusinessObjectAttributesTempModelImpl) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object_attributes_temp SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.db.ExecContext(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object_attributes_temp by form_view_id failed: %w", err)
	}
	return nil
}

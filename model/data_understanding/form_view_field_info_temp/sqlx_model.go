// Package form_view_field_info_temp 库表字段信息临时表Model
package form_view_field_info_temp

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewFormViewFieldInfoTempModel 创建FormViewFieldInfoTempModel实例
func NewFormViewFieldInfoTempModel(db *sqlx.Tx) *FormViewFieldInfoTempModelImpl {
	return &FormViewFieldInfoTempModelImpl{db: db}
}

// FormViewFieldInfoTempModelImpl FormViewFieldInfoTempModel实现
type FormViewFieldInfoTempModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入字段信息临时记录
func (m *FormViewFieldInfoTempModelImpl) Insert(ctx context.Context, data *FormViewFieldInfoTemp) (*FormViewFieldInfoTemp, error) {
	query := `INSERT INTO t_form_view_field_info_temp (id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.FormViewId, data.FormViewFieldId, data.UserId, data.Version, data.FieldBusinessName, data.FieldRole, data.FieldDescription)
	if err != nil {
		return nil, fmt.Errorf("insert form_view_field_info_temp failed: %w", err)
	}
	return data, nil
}

// FindByFormViewAndVersion 根据form_view_id和version查询字段列表
func (m *FormViewFieldInfoTempModelImpl) FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*FormViewFieldInfoTemp, error) {
	var resp []*FormViewFieldInfoTemp
	query := `SELECT id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description, created_at, updated_at, deleted_at
	           FROM t_form_view_field_info_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.db.SelectContext(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find form_view_field_info_temp by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// FindOneByFormFieldId 根据form_view_field_id查询字段信息
func (m *FormViewFieldInfoTempModelImpl) FindOneByFormFieldId(ctx context.Context, formViewFieldId string) (*FormViewFieldInfoTemp, error) {
	var resp FormViewFieldInfoTemp
	query := `SELECT id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description, created_at, updated_at, deleted_at
	           FROM t_form_view_field_info_temp
	           WHERE form_view_field_id = ? AND deleted_at IS NULL ORDER BY version DESC LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, formViewFieldId)
	if err != nil {
		return nil, fmt.Errorf("find form_view_field_info_temp by form_view_field_id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新字段信息
func (m *FormViewFieldInfoTempModelImpl) Update(ctx context.Context, data *FormViewFieldInfoTemp) error {
	query := `UPDATE t_form_view_field_info_temp
	           SET field_business_name = ?, field_role = ?, field_description = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.FieldBusinessName, data.FieldRole, data.FieldDescription, data.Id)
	if err != nil {
		return fmt.Errorf("update form_view_field_info_temp failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *FormViewFieldInfoTempModelImpl) WithTx(tx interface{}) FormViewFieldInfoTempModel {
	return &FormViewFieldInfoTempModelImpl{db: tx.(*sqlx.Tx)}
}

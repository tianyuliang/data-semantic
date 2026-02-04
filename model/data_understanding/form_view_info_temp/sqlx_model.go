// Package form_view_info_temp 库表信息临时表Model
package form_view_info_temp

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewFormViewInfoTempModel 创建FormViewInfoTempModel实例
func NewFormViewInfoTempModel(db *sqlx.Tx) *FormViewInfoTempModelImpl {
	return &FormViewInfoTempModelImpl{db: db}
}

// FormViewInfoTempModelImpl FormViewInfoTempModel实现
type FormViewInfoTempModelImpl struct {
	db *sqlx.Tx
}

// Insert 插入库表信息临时记录
func (m *FormViewInfoTempModelImpl) Insert(ctx context.Context, data *FormViewInfoTemp) (*FormViewInfoTemp, error) {
	query := `INSERT INTO t_form_view_info_temp (id, form_view_id, user_id, version, table_business_name, table_description)
	           VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, data.Id, data.FormViewId, data.UserId, data.Version, data.TableBusinessName, data.TableDescription)
	if err != nil {
		return nil, fmt.Errorf("insert form_view_info_temp failed: %w", err)
	}
	return data, nil
}

// FindOneByFormViewAndVersion 根据form_view_id和version查询记录
func (m *FormViewInfoTempModelImpl) FindOneByFormViewAndVersion(ctx context.Context, formViewId string, version int) (*FormViewInfoTemp, error) {
	var resp FormViewInfoTemp
	query := `SELECT id, form_view_id, user_id, version, table_business_name, table_description, created_at, updated_at, deleted_at
	           FROM t_form_view_info_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find form_view_info_temp by form_view_id and version failed: %w", err)
	}
	return &resp, nil
}

// FindLatestByFormViewId 查询指定form_view_id的最新版本记录
func (m *FormViewInfoTempModelImpl) FindLatestByFormViewId(ctx context.Context, formViewId string) (*FormViewInfoTemp, error) {
	var resp FormViewInfoTemp
	query := `SELECT id, form_view_id, user_id, version, table_business_name, table_description, created_at, updated_at, deleted_at
	           FROM t_form_view_info_temp
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY version DESC LIMIT 1`
	err := m.db.GetContext(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find latest form_view_info_temp by form_view_id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新库表信息
func (m *FormViewInfoTempModelImpl) Update(ctx context.Context, data *FormViewInfoTemp) error {
	query := `UPDATE t_form_view_info_temp
	           SET table_business_name = ?, table_description = ?
	           WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, data.TableBusinessName, data.TableDescription, data.Id)
	if err != nil {
		return fmt.Errorf("update form_view_info_temp failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *FormViewInfoTempModelImpl) WithTx(tx interface{}) FormViewInfoTempModel {
	return &FormViewInfoTempModelImpl{db: tx.(*sqlx.Tx)}
}

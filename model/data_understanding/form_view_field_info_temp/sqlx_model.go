// Package form_view_field_info_temp 库表字段信息临时表Model (Sqlx实现)
package form_view_field_info_temp

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewFormViewFieldInfoTempModelSqlx 创建FormViewFieldInfoTempModelSqlx实例
func NewFormViewFieldInfoTempModelSqlx(conn sqlx.SqlConn) *FormViewFieldInfoTempModelSqlx {
	return &FormViewFieldInfoTempModelSqlx{conn: conn}
}

// NewFormViewFieldInfoTempModelSession 创建FormViewFieldInfoTempModelSqlx实例 (使用 Session)
func NewFormViewFieldInfoTempModelSession(session sqlx.Session) *FormViewFieldInfoTempModelSqlx {
	return &FormViewFieldInfoTempModelSqlx{conn: session}
}

// FormViewFieldInfoTempModelSqlx FormViewFieldInfoTempModel实现 (基于 go-zero Sqlx)
type FormViewFieldInfoTempModelSqlx struct {
	conn sqlx.Session
}

// Insert 插入字段信息临时记录
func (m *FormViewFieldInfoTempModelSqlx) Insert(ctx context.Context, data *FormViewFieldInfoTemp) (*FormViewFieldInfoTemp, error) {
	query := `INSERT INTO t_form_view_field_info_temp (id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, data.Id, data.FormViewId, data.FormViewFieldId, data.UserId, data.Version, data.FieldBusinessName, data.FieldRole, data.FieldDescription)
	if err != nil {
		return nil, fmt.Errorf("insert form_view_field_info_temp failed: %w", err)
	}
	return data, nil
}

// FindOneByFormFieldId 根据form_view_field_id查询字段信息
func (m *FormViewFieldInfoTempModelSqlx) FindOneByFormFieldId(ctx context.Context, formViewFieldId string) (*FormViewFieldInfoTemp, error) {
	var resp FormViewFieldInfoTemp
	query := `SELECT id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description, created_at, updated_at, deleted_at
	           FROM t_form_view_field_info_temp
	           WHERE form_view_field_id = ? AND deleted_at IS NULL ORDER BY version DESC LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, formViewFieldId)
	if err != nil {
		return nil, fmt.Errorf("find form_view_field_info_temp by form_view_field_id failed: %w", err)
	}
	return &resp, nil
}

// FindLatestByFormViewId 查询指定form_view_id的最新版本字段列表
func (m *FormViewFieldInfoTempModelSqlx) FindLatestByFormViewId(ctx context.Context, formViewId string) ([]*FormViewFieldInfoTemp, error) {
	var resp []*FormViewFieldInfoTemp
	query := `SELECT id, form_view_id, form_view_field_id, user_id, version, field_business_name, field_role, field_description, created_at, updated_at, deleted_at
	           FROM t_form_view_field_info_temp
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY version DESC, id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find latest form_view_field_info_temp by form_view_id failed: %w", err)
	}
	// 去重：取每个form_view_field_id的最新版本
	return m.deduplicateByFieldId(resp), nil
}

// deduplicateByFieldId 去重，保留每个form_view_field_id的最新版本
func (m *FormViewFieldInfoTempModelSqlx) deduplicateByFieldId(fields []*FormViewFieldInfoTemp) []*FormViewFieldInfoTemp {
	fieldMap := make(map[string]*FormViewFieldInfoTemp)
	for _, f := range fields {
		if existing, ok := fieldMap[f.FormViewFieldId]; !ok || f.Version > existing.Version {
			fieldMap[f.FormViewFieldId] = f
		}
	}
	result := make([]*FormViewFieldInfoTemp, 0, len(fieldMap))
	for _, f := range fieldMap {
		result = append(result, f)
	}
	// 按id排序以保证稳定输出
	// 简单排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Id > result[j].Id {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// Update 更新字段信息
func (m *FormViewFieldInfoTempModelSqlx) Update(ctx context.Context, data *FormViewFieldInfoTemp) error {
	query := `UPDATE t_form_view_field_info_temp
	           SET field_business_name = ?, field_role = ?, field_description = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.FieldBusinessName, data.FieldRole, data.FieldDescription, data.Id)
	if err != nil {
		return fmt.Errorf("update form_view_field_info_temp failed: %w", err)
	}
	return nil
}

// DeleteByFormViewId 逻辑删除指定form_view_id的所有记录
func (m *FormViewFieldInfoTempModelSqlx) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_form_view_field_info_temp SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete form_view_field_info_temp by form_view_id failed: %w", err)
	}
	return nil
}

// DeleteByFormFieldId 逻辑删除指定form_view_field_id的所有记录
func (m *FormViewFieldInfoTempModelSqlx) DeleteByFormFieldId(ctx context.Context, formViewFieldId string) error {
	query := `UPDATE t_form_view_field_info_temp SET deleted_at = NOW(3) WHERE form_view_field_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, formViewFieldId)
	if err != nil {
		return fmt.Errorf("delete form_view_field_info_temp by form_view_field_id failed: %w", err)
	}
	return nil
}

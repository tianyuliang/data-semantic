// Package form_view_field_info_temp 库表字段信息临时表Model (SqlConn实现)
package form_view_field_info_temp

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewFormViewFieldInfoTempModelSqlConn 创建FormViewFieldInfoTempModelSqlConn实例
func NewFormViewFieldInfoTempModelSqlConn(conn sqlx.SqlConn) *FormViewFieldInfoTempModelSqlConn {
	return &FormViewFieldInfoTempModelSqlConn{conn: conn}
}

// FormViewFieldInfoTempModelSqlConn FormViewFieldInfoTempModel实现 (基于 go-zero SqlConn)
type FormViewFieldInfoTempModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindLatestByFormViewId 查询指定form_view_id的最新版本字段列表
func (m *FormViewFieldInfoTempModelSqlConn) FindLatestByFormViewId(ctx context.Context, formViewId string) ([]*FormViewFieldInfoTemp, error) {
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
func (m *FormViewFieldInfoTempModelSqlConn) deduplicateByFieldId(fields []*FormViewFieldInfoTemp) []*FormViewFieldInfoTemp {
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
func (m *FormViewFieldInfoTempModelSqlConn) Update(ctx context.Context, data *FormViewFieldInfoTemp) error {
	query := `UPDATE t_form_view_field_info_temp
	           SET field_business_name = ?, field_role = ?, field_description = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.FieldBusinessName, data.FieldRole, data.FieldDescription, data.Id)
	if err != nil {
		return fmt.Errorf("update form_view_field_info_temp failed: %w", err)
	}
	return nil
}

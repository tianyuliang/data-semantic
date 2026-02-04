// Package form_view_field 字段Model
package form_view_field

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewFormViewFieldModel 创建FormViewFieldModel实例 (使用 go-zero SqlConn)
func NewFormViewFieldModel(conn sqlx.SqlConn) *FormViewFieldModelSqlConn {
	return &FormViewFieldModelSqlConn{conn: conn}
}

// FormViewFieldModelSqlConn FormViewFieldModel实现 (基于 go-zero SqlConn)
type FormViewFieldModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindByFormViewId 根据form_view_id查询字段列表
func (m *FormViewFieldModelSqlConn) FindByFormViewId(ctx context.Context, formViewId string) ([]*FormViewFieldBase, error) {
	var resp []*FormViewFieldBase
	query := `SELECT id, field_tech_name, field_type FROM form_view_field WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find form_view_field by form_view_id failed: %w", err)
	}
	return resp, nil
}

// FindFullByFormViewId 根据form_view_id查询字段完整信息 (包含语义信息)
func (m *FormViewFieldModelSqlConn) FindFullByFormViewId(ctx context.Context, formViewId string) ([]*FormViewField, error) {
	var resp []*FormViewField
	query := `SELECT id, form_view_id, field_tech_name, field_type, business_name, field_role, field_description
	          FROM form_view_field WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find form_view_field full info by form_view_id failed: %w", err)
	}
	return resp, nil
}

// FormViewFieldBase 字段基础信息结构
type FormViewFieldBase struct {
	Id           string `db:"id"`
	FieldTechName string `db:"field_tech_name"`
	FieldType    string `db:"field_type"`
}

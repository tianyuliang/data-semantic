// Package form_view_info_temp 库表信息临时表Model (SqlConn实现)
package form_view_info_temp

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewFormViewInfoTempModelSqlConn 创建FormViewInfoTempModelSqlConn实例
func NewFormViewInfoTempModelSqlConn(conn sqlx.SqlConn) *FormViewInfoTempModelSqlConn {
	return &FormViewInfoTempModelSqlConn{conn: conn}
}

// FormViewInfoTempModelSqlConn FormViewInfoTempModel实现 (基于 go-zero SqlConn)
type FormViewInfoTempModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindLatestByFormViewId 查询指定form_view_id的最新版本记录
func (m *FormViewInfoTempModelSqlConn) FindLatestByFormViewId(ctx context.Context, formViewId string) (*FormViewInfoTemp, error) {
	var resp FormViewInfoTemp
	query := `SELECT id, form_view_id, user_id, version, table_business_name, table_description, created_at, updated_at, deleted_at
	           FROM t_form_view_info_temp
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY version DESC LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find latest form_view_info_temp by form_view_id failed: %w", err)
	}
	return &resp, nil
}

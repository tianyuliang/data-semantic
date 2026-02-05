// Package business_object 业务对象正式表Model (SqlConn实现)
package business_object

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectModelSqlConn 创建BusinessObjectModelSqlConn实例
func NewBusinessObjectModelSqlConn(conn sqlx.SqlConn) *BusinessObjectModelSqlConn {
	return &BusinessObjectModelSqlConn{conn: conn}
}

// BusinessObjectModelSqlConn BusinessObjectModel实现 (基于 go-zero SqlConn)
type BusinessObjectModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindByFormViewId 根据form_view_id查询业务对象列表
func (m *BusinessObjectModelSqlConn) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObject, error) {
	var resp []*BusinessObject
	query := `SELECT id, object_name, object_type, form_view_id, status, created_at, updated_at, deleted_at
	           FROM t_business_object
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find business_object by form_view_id failed: %w", err)
	}
	return resp, nil
}

// FindOneById 根据id查询业务对象
func (m *BusinessObjectModelSqlConn) FindOneById(ctx context.Context, id string) (*BusinessObject, error) {
	var resp BusinessObject
	query := `SELECT id, object_name, object_type, form_view_id, status, created_at, updated_at, deleted_at
	           FROM t_business_object
	           WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find business_object by id failed: %w", err)
	}
	return &resp, nil
}

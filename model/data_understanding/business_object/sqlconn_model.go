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

// NewBusinessObjectModelSession 创建BusinessObjectModelSqlConn实例 (使用 Session)
func NewBusinessObjectModelSession(session sqlx.Session) *BusinessObjectModelSqlConn {
	return &BusinessObjectModelSqlConn{conn: session}
}

// BusinessObjectModelSqlConn BusinessObjectModel实现 (基于 go-zero SqlConn)
type BusinessObjectModelSqlConn struct {
	conn sqlx.Session
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

// DeleteByFormViewId 根据form_view_id删除所有业务对象
func (m *BusinessObjectModelSqlConn) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object by form_view_id failed: %w", err)
	}
	return nil
}

// BatchInsertFromTemp 从临时表批量插入业务对象
func (m *BusinessObjectModelSqlConn) BatchInsertFromTemp(ctx context.Context, formViewId string, version int) (int, error) {
	query := `INSERT INTO t_business_object (id, object_name, object_type, form_view_id, status)
	           SELECT id, object_name, 0, form_view_id, 1
	           FROM t_business_object_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL`
	result, err := m.conn.ExecCtx(ctx, query, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("batch insert business_object from temp failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

// CountByFormViewId 根据form_view_id统计业务对象数量
func (m *BusinessObjectModelSqlConn) CountByFormViewId(ctx context.Context, formViewId string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM t_business_object WHERE form_view_id = ? AND deleted_at IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, formViewId)
	if err != nil {
		return 0, fmt.Errorf("count business_object by form_view_id failed: %w", err)
	}
	return count, nil
}

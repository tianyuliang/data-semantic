// Package business_object_temp 业务对象临时表Model (SqlConn实现)
package business_object_temp

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectTempModelSqlConn 创建BusinessObjectTempModelSqlConn实例
func NewBusinessObjectTempModelSqlConn(conn sqlx.SqlConn) *BusinessObjectTempModelSqlConn {
	return &BusinessObjectTempModelSqlConn{conn: conn}
}

// BusinessObjectTempModelSqlConn BusinessObjectTempModel实现 (基于 go-zero SqlConn)
type BusinessObjectTempModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindByFormViewAndVersion 根据form_view_id和version查询业务对象列表
func (m *BusinessObjectTempModelSqlConn) FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectTemp, error) {
	var resp []*BusinessObjectTemp
	query := `SELECT id, form_view_id, user_id, version, object_name, created_at, updated_at, deleted_at
	           FROM t_business_object_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find business_object_temp by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// FindOneById 根据id查询业务对象
func (m *BusinessObjectTempModelSqlConn) FindOneById(ctx context.Context, id string) (*BusinessObjectTemp, error) {
	var resp BusinessObjectTemp
	query := `SELECT id, form_view_id, user_id, version, object_name, created_at, updated_at, deleted_at
	           FROM t_business_object_temp
	           WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find business_object_temp by id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新业务对象名称
func (m *BusinessObjectTempModelSqlConn) Update(ctx context.Context, data *BusinessObjectTemp) error {
	query := `UPDATE t_business_object_temp
	           SET object_name = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.ObjectName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_temp failed: %w", err)
	}
	return nil
}

// FindLatestVersionByFormViewId 查询指定form_view_id的最新版本号
func (m *BusinessObjectTempModelSqlConn) FindLatestVersionByFormViewId(ctx context.Context, formViewId string) (int, error) {
	var result struct {
		LatestVersion int `db:"latest_version"`
	}
	query := `SELECT COALESCE(MAX(version), 0) AS latest_version
	           FROM t_business_object_temp
	           WHERE form_view_id = ? AND deleted_at IS NULL`
	err := m.conn.QueryRowCtx(ctx, &result, query, formViewId)
	if err != nil {
		return 0, fmt.Errorf("find latest version by form_view_id failed: %w", err)
	}
	return result.LatestVersion, nil
}

// FindByFormViewIdLatest 查询指定form_view_id的最新版本业务对象列表
func (m *BusinessObjectTempModelSqlConn) FindByFormViewIdLatest(ctx context.Context, formViewId string) ([]*BusinessObjectTemp, error) {
	// 先获取最新版本号
	latestVersion, err := m.FindLatestVersionByFormViewId(ctx, formViewId)
	if err != nil {
		return nil, err
	}
	// 如果没有数据，返回空列表
	if latestVersion == 0 {
		return []*BusinessObjectTemp{}, nil
	}
	return m.FindByFormViewAndVersion(ctx, formViewId, latestVersion)
}

// Package form_view 库表视图Model
package form_view

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewFormViewModel 创建FormViewModel实例 (使用 go-zero SqlConn)
func NewFormViewModel(conn sqlx.SqlConn) *FormViewModelSqlConn {
	return &FormViewModelSqlConn{conn: conn}
}

// FormViewModelSqlConn FormViewModel实现 (基于 go-zero SqlConn)
type FormViewModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindOneById 根据id查询库表视图
func (m *FormViewModelSqlConn) FindOneById(ctx context.Context, id string) (*FormView, error) {
	var resp FormView
	query := `SELECT id, understand_status FROM form_view WHERE id = ? LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find form_view by id failed: %w", err)
	}
	return &resp, nil
}

// GetTableTechName 获取表技术名称
func (m *FormViewModelSqlConn) GetTableTechName(ctx context.Context, id string) (string, error) {
	var result struct {
		TableTechName string `db:"table_tech_name"`
	}
	query := `SELECT table_tech_name FROM form_view WHERE id = ? LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &result, query, id)
	if err != nil {
		return "", fmt.Errorf("get table_tech_name failed: %w", err)
	}
	return result.TableTechName, nil
}

// GetTableInfo 获取表信息 (包含状态和技术名称)
func (m *FormViewModelSqlConn) GetTableInfo(ctx context.Context, id string) (*FormViewTableInfo, error) {
	var resp FormViewTableInfo
	query := `SELECT id, understand_status, table_tech_name FROM form_view WHERE id = ? LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("get table info failed: %w", err)
	}
	return &resp, nil
}

// UpdateUnderstandStatus 更新理解状态
func (m *FormViewModelSqlConn) UpdateUnderstandStatus(ctx context.Context, id string, status int8) error {
	query := `UPDATE form_view SET understand_status = ? WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update form_view understand_status failed: %w", err)
	}
	return nil
}

// FormViewTableInfo 表信息结构
type FormViewTableInfo struct {
	Id              string `db:"id"`
	UnderstandStatus int8  `db:"understand_status"`
	TableTechName   string `db:"table_tech_name"`
}

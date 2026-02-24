// Package business_object 业务对象正式表Model (Sqlx实现)
package business_object

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectModelSqlx 创建BusinessObjectModelSqlx实例
func NewBusinessObjectModelSqlx(conn sqlx.SqlConn) *BusinessObjectModelSqlx {
	return &BusinessObjectModelSqlx{conn: conn}
}

// NewBusinessObjectModelSession 创建BusinessObjectModelSqlx实例 (使用 Session)
func NewBusinessObjectModelSession(session sqlx.Session) *BusinessObjectModelSqlx {
	return &BusinessObjectModelSqlx{conn: session}
}

// BusinessObjectModelSqlx BusinessObjectModel实现 (基于 go-zero Sqlx)
type BusinessObjectModelSqlx struct {
	conn sqlx.Session
}

// Insert 插入业务对象记录
func (m *BusinessObjectModelSqlx) Insert(ctx context.Context, data *BusinessObject) (*BusinessObject, error) {
	query := `INSERT INTO t_business_object (id, object_name, object_type, form_view_id, status)
	           VALUES (?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, data.Id, data.ObjectName, data.ObjectType, data.FormViewId, data.Status)
	if err != nil {
		return nil, fmt.Errorf("insert business_object failed: %w", err)
	}
	return data, nil
}

// Update 更新业务对象
func (m *BusinessObjectModelSqlx) Update(ctx context.Context, data *BusinessObject) error {
	query := `UPDATE t_business_object
	           SET object_name = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.ObjectName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object failed: %w", err)
	}
	return nil
}

// Delete 逻辑删除业务对象
func (m *BusinessObjectModelSqlx) Delete(ctx context.Context, id string) error {
	query := `UPDATE t_business_object SET deleted_at = NOW(3) WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete business_object failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectModelSqlx) WithTx(tx interface{}) BusinessObjectModel {
	session, ok := tx.(sqlx.Session)
	if !ok {
		return nil
	}
	return &BusinessObjectModelSqlx{conn: session}
}

// FindByFormViewId 根据form_view_id查询业务对象列表
func (m *BusinessObjectModelSqlx) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObject, error) {
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

// FindByFormViewIdAndObjectName 根据form_view_id和object_name查询单个业务对象
func (m *BusinessObjectModelSqlx) FindByFormViewIdAndObjectName(ctx context.Context, formViewId string, objectName string) (*BusinessObject, error) {
	var resp BusinessObject
	query := `SELECT id, object_name, object_type, form_view_id, status, created_at, updated_at, deleted_at
	           FROM t_business_object
	           WHERE form_view_id = ? AND object_name = ? AND deleted_at IS NULL LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, formViewId, objectName)
	if err != nil {
		return nil, fmt.Errorf("find business_object by form_view_id and object_name failed: %w", err)
	}
	return &resp, nil
}

// FindOneById 根据id查询业务对象
func (m *BusinessObjectModelSqlx) FindOneById(ctx context.Context, id string) (*BusinessObject, error) {
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
func (m *BusinessObjectModelSqlx) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object by form_view_id failed: %w", err)
	}
	return nil
}

// BatchInsertFromTemp 从临时表批量插入业务对象
func (m *BusinessObjectModelSqlx) BatchInsertFromTemp(ctx context.Context, formViewId string, version int) (int, error) {
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
func (m *BusinessObjectModelSqlx) CountByFormViewId(ctx context.Context, formViewId string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM t_business_object WHERE form_view_id = ? AND deleted_at IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, formViewId)
	if err != nil {
		return 0, fmt.Errorf("count business_object by form_view_id failed: %w", err)
	}
	return count, nil
}

// ========== 增量更新相关方法实现 ==========

// UpdateByFormalId 根据formal_id更新业务对象（增量更新）
// 更新临时表中有formal_id的记录到正式表
func (m *BusinessObjectModelSqlx) UpdateByFormalId(ctx context.Context, formViewId string, version int) (int, error) {
	query := `UPDATE t_business_object bo
	           JOIN t_business_object_temp bot ON bo.id = bot.formal_id
	           SET bo.object_name = bot.object_name,
	               bo.updated_at = NOW(3)
	           WHERE bot.form_view_id = ?
	             AND bot.version = ?
	             AND bot.formal_id IS NOT NULL
	             AND bot.deleted_at IS NULL`
	result, err := m.conn.ExecCtx(ctx, query, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("update business_object by formal_id failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

// InsertFromTempWithoutFormalId 从临时表插入formal_id为NULL的记录（增量更新）
func (m *BusinessObjectModelSqlx) InsertFromTempWithoutFormalId(ctx context.Context, formViewId string, version int) (int, error) {
	query := `INSERT INTO t_business_object (id, object_name, object_type, form_view_id, status, created_at, updated_at)
	           SELECT id, object_name, 1, form_view_id, 1, NOW(3), NOW(3)
	           FROM t_business_object_temp
	           WHERE form_view_id = ?
	             AND version = ?
	             AND formal_id IS NULL
	             AND deleted_at IS NULL`
	result, err := m.conn.ExecCtx(ctx, query, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("insert business_object from temp without formal_id failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

// DeleteNotInFormalIdList 删除不在temp表formal_id列表中的记录（增量更新）
// 逻辑删除正式表中有但临时表中没有的记录（通过formal_id判断）
func (m *BusinessObjectModelSqlx) DeleteNotInFormalIdList(ctx context.Context, formViewId string, version int) (int, error) {
	query := `UPDATE t_business_object
	           SET deleted_at = NOW(3)
	           WHERE form_view_id = ?
	             AND deleted_at IS NULL
	             AND id NOT IN (
	               SELECT formal_id FROM t_business_object_temp
	               WHERE form_view_id = ?
	                 AND version = ?
	                 AND formal_id IS NOT NULL
	                 AND deleted_at IS NULL
	             )`
	result, err := m.conn.ExecCtx(ctx, query, formViewId, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("delete business_object not in formal_id list failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

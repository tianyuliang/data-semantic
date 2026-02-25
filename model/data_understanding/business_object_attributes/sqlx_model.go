// Package business_object_attributes 业务对象属性正式表Model (Sqlx实现)
package business_object_attributes

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectAttributesModelSqlx 创建BusinessObjectAttributesModelSqlx实例
func NewBusinessObjectAttributesModelSqlx(conn sqlx.SqlConn) *BusinessObjectAttributesModelSqlx {
	return &BusinessObjectAttributesModelSqlx{conn: conn}
}

// NewBusinessObjectAttributesModelSession 创建BusinessObjectAttributesModelSqlx实例 (使用 Session)
func NewBusinessObjectAttributesModelSession(session sqlx.Session) *BusinessObjectAttributesModelSqlx {
	return &BusinessObjectAttributesModelSqlx{conn: session}
}

// BusinessObjectAttributesModelSqlx BusinessObjectAttributesModel实现 (基于 go-zero Sqlx)
type BusinessObjectAttributesModelSqlx struct {
	conn sqlx.Session
}

// Insert 插入业务对象属性记录
func (m *BusinessObjectAttributesModelSqlx) Insert(ctx context.Context, data *BusinessObjectAttributes) (*BusinessObjectAttributes, error) {
	query := `INSERT INTO t_business_object_attributes (id, form_view_id, business_object_id, form_view_field_id, attr_name)
	           VALUES (?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, data.Id, data.FormViewId, data.BusinessObjectId, data.FormViewFieldId, data.AttrName)
	if err != nil {
		return nil, fmt.Errorf("insert business_object_attributes failed: %w", err)
	}
	return data, nil
}

// Update 更新属性
func (m *BusinessObjectAttributesModelSqlx) Update(ctx context.Context, data *BusinessObjectAttributes) error {
	query := `UPDATE t_business_object_attributes
	           SET attr_name = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.AttrName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_attributes failed: %w", err)
	}
	return nil
}

// Delete 逻辑删除属性
func (m *BusinessObjectAttributesModelSqlx) Delete(ctx context.Context, id string) error {
	query := `UPDATE t_business_object_attributes SET deleted_at = NOW(3) WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete business_object_attributes failed: %w", err)
	}
	return nil
}

// WithTx 设置事务
func (m *BusinessObjectAttributesModelSqlx) WithTx(tx interface{}) BusinessObjectAttributesModel {
	session, ok := tx.(sqlx.Session)
	if !ok {
		return nil
	}
	return &BusinessObjectAttributesModelSqlx{conn: session}
}

// FindByBusinessObjectId 根据business_object_id查询属性列表
func (m *BusinessObjectAttributesModelSqlx) FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributes, error) {
	var resp []*BusinessObjectAttributes
	query := `SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes
	           WHERE business_object_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes by business_object_id failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewId 根据form_view_id查询所有属性
func (m *BusinessObjectAttributesModelSqlx) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObjectAttributes, error) {
	var resp []*BusinessObjectAttributes
	query := `SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes
	           WHERE form_view_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes by form_view_id failed: %w", err)
	}
	return resp, nil
}

// FieldWithAttrInfo 属性关联字段信息
type FieldWithAttrInfo struct {
	Id                string  `db:"id"`
	BusinessObjectId  string  `db:"business_object_id"`
	FormViewFieldId   string  `db:"form_view_field_id"`
	AttrName          string  `db:"attr_name"`
	FieldTechName     string  `db:"field_tech_name"`
	FieldBusinessName *string `db:"field_business_name"`
	FieldRole         *int8   `db:"field_role"`
	FieldType         string  `db:"field_type"`
}

// FindByBusinessObjectIdWithFieldInfo 根据business_object_id查询属性列表（包含字段信息）
func (m *BusinessObjectAttributesModelSqlx) FindByBusinessObjectIdWithFieldInfo(ctx context.Context, businessObjectId string) ([]*FieldWithAttrInfo, error) {
	var resp []*FieldWithAttrInfo
	query := `SELECT boa.id, boa.business_object_id, boa.form_view_field_id, boa.attr_name,
	           fvf.field_tech_name, fvf.business_name AS field_business_name, fvf.field_role, fvf.field_type
	           FROM t_business_object_attributes boa
	           INNER JOIN t_form_view_field fvf ON boa.form_view_field_id = fvf.id
	           WHERE boa.business_object_id = ? AND boa.deleted_at IS NULL AND fvf.deleted_at IS NULL
	           ORDER BY boa.id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes with field info failed: %w", err)
	}
	return resp, nil
}

// DeleteByFormViewId 根据form_view_id删除所有属性
func (m *BusinessObjectAttributesModelSqlx) DeleteByFormViewId(ctx context.Context, formViewId string) error {
	query := `UPDATE t_business_object_attributes SET deleted_at = NOW(3) WHERE form_view_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, formViewId)
	if err != nil {
		return fmt.Errorf("delete business_object_attributes by form_view_id failed: %w", err)
	}
	return nil
}

// BatchInsertFromTemp 从临时表批量插入属性
func (m *BusinessObjectAttributesModelSqlx) BatchInsertFromTemp(ctx context.Context, formViewId string, version int) (int, error) {
	query := `INSERT INTO t_business_object_attributes (id, form_view_id, business_object_id, form_view_field_id, attr_name)
	           SELECT id, form_view_id, business_object_id, form_view_field_id, attr_name
	           FROM t_business_object_attributes_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL`
	result, err := m.conn.ExecCtx(ctx, query, formViewId, version)
	if err != nil {
		return 0, fmt.Errorf("batch insert business_object_attributes from temp failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), nil
}

// ========== 增量更新相关方法实现 ==========

// MergeFromTemp 从临时表合并数据到正式表（基于业务对象匹配 + form_view_field_id）
// 返回：inserted=新增数量, updated=更新数量, deleted=删除数量
func (m *BusinessObjectAttributesModelSqlx) MergeFromTemp(ctx context.Context, formViewId string, version int) (inserted, updated, deleted int, err error) {
	// 1. 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现合并
	// 通过业务对象名称关联正式表的 business_object_id
	query := `INSERT INTO t_business_object_attributes (id, form_view_id, business_object_id, form_view_field_id, attr_name, created_at, updated_at)
	           SELECT boat.id, boat.form_view_id, bo.id AS business_object_id, boat.form_view_field_id, boat.attr_name, NOW(3), NOW(3)
	           FROM t_business_object_attributes_temp boat
	           INNER JOIN t_business_object_temp bot ON boat.business_object_id = bot.id
	           INNER JOIN t_business_object bo ON bo.form_view_id = bot.form_view_id AND bo.object_name = bot.object_name
	           WHERE boat.form_view_id = ?
	             AND bot.version = ?
	             AND boat.deleted_at IS NULL
	             AND bo.deleted_at IS NULL
	           ON DUPLICATE KEY UPDATE
	              attr_name = VALUES(attr_name),
	              updated_at = NOW(3)`

	result, err := m.conn.ExecCtx(ctx, query, formViewId, version)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("merge business_object_attributes from temp failed: %w", err)
	}

	totalAffected, err := result.RowsAffected()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get affected rows failed: %w", err)
	}

	// 2. 删除正式表中不在临时表的属性（融合规则）
	deleteQuery := `UPDATE t_business_object_attributes boa
	                SET boa.deleted_at = NOW(3)
	                WHERE boa.form_view_id = ?
	                  AND boa.deleted_at IS NULL
	                  AND boa.business_object_id IN (
	                    SELECT DISTINCT bo.id
	                    FROM t_business_object bo
	                    INNER JOIN t_business_object_temp bot ON bo.form_view_id = bot.form_view_id AND bo.object_name = bot.object_name
	                    WHERE bot.form_view_id = ?
	                      AND bot.version = ?
	                      AND bot.deleted_at IS NULL
	                  )
	                  AND boa.form_view_field_id IN (
	                    SELECT DISTINCT boat.form_view_field_id
	                    FROM t_business_object_attributes_temp boat
	                    WHERE boat.form_view_id = ?
	                      AND boat.deleted_at IS NULL
	                  )
	                  AND (boa.business_object_id, boa.form_view_field_id) NOT IN (
	                    SELECT bo.id, boat.form_view_field_id
	                    FROM t_business_object_attributes_temp boat
	                    INNER JOIN t_business_object_temp bot ON boat.business_object_id = bot.id
	                    INNER JOIN t_business_object bo ON bo.form_view_id = bot.form_view_id AND bo.object_name = bot.object_name
	                    WHERE bot.form_view_id = ?
	                      AND bot.version = ?
	                      AND boat.deleted_at IS NULL
	                  )`

	deleteResult, err := m.conn.ExecCtx(ctx, deleteQuery, formViewId, formViewId, version, formViewId, formViewId, version)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("delete business_object_attributes not in temp failed: %w", err)
	}

	deletedCount, err := deleteResult.RowsAffected()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get deleted rows failed: %w", err)
	}

	inserted = 0
	updated = int(totalAffected)
	deleted = int(deletedCount)

	return inserted, updated, deleted, nil
}

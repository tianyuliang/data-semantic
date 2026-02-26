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
	Description       *string `db:"field_description"`
}

// FindByBusinessObjectIdWithFieldInfo 根据business_object_id查询属性列表（包含字段信息）
func (m *BusinessObjectAttributesModelSqlx) FindByBusinessObjectIdWithFieldInfo(ctx context.Context, businessObjectId string) ([]*FieldWithAttrInfo, error) {
	var resp []*FieldWithAttrInfo
	query := `SELECT boa.id, boa.business_object_id, boa.form_view_field_id, boa.attr_name,
	           fvf.technical_name AS field_tech_name, fvf.business_name AS field_business_name,
	           fvf.field_role, fvf.data_type AS field_type, fvf.field_description
	           FROM t_business_object_attributes boa
	           INNER JOIN form_view_field fvf ON boa.form_view_field_id = fvf.id COLLATE utf8mb4_unicode_ci
	           WHERE boa.business_object_id = ? AND boa.deleted_at IS NULL AND fvf.deleted_at = 0
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

// FindUnrecognizedFields 查询未识别字段（business_object_id 和 attr_name 都为空的记录）
func (m *BusinessObjectAttributesModelSqlx) FindUnrecognizedFields(ctx context.Context, formViewId string) ([]*UnrecognizedFieldInfo, error) {
	var resp []*UnrecognizedFieldInfo
	query := `SELECT boa.id, boa.form_view_field_id,
	           fvf.technical_name AS field_tech_name, fvf.business_name AS field_business_name,
	           fvf.field_role, fvf.data_type AS field_type, fvf.field_description
	           FROM t_business_object_attributes boa
	           INNER JOIN form_view_field fvf ON boa.form_view_field_id = fvf.id COLLATE utf8mb4_unicode_ci
	           WHERE boa.form_view_id = ? AND boa.business_object_id = '' AND boa.attr_name = '' AND boa.deleted_at IS NULL
	           ORDER BY boa.id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find unrecognized fields failed: %w", err)
	}
	return resp, nil
}

// UnrecognizedFieldInfo 未识别字段信息
type UnrecognizedFieldInfo struct {
	Id                string  `db:"id"`
	FormViewFieldId   string  `db:"form_view_field_id"`
	FieldTechName     string  `db:"field_tech_name"`
	FieldBusinessName *string `db:"field_business_name"`
	FieldRole         *int8   `db:"field_role"`
	FieldType         string  `db:"field_type"`
	Description       *string `db:"field_description"`
}

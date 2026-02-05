// Package business_object_attributes 业务对象属性正式表Model (SqlConn实现)
package business_object_attributes

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectAttributesModelSqlConn 创建BusinessObjectAttributesModelSqlConn实例
func NewBusinessObjectAttributesModelSqlConn(conn sqlx.SqlConn) *BusinessObjectAttributesModelSqlConn {
	return &BusinessObjectAttributesModelSqlConn{conn: conn}
}

// BusinessObjectAttributesModelSqlConn BusinessObjectAttributesModel实现 (基于 go-zero SqlConn)
type BusinessObjectAttributesModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindByBusinessObjectId 根据business_object_id查询属性列表
func (m *BusinessObjectAttributesModelSqlConn) FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributes, error) {
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
func (m *BusinessObjectAttributesModelSqlConn) FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObjectAttributes, error) {
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
func (m *BusinessObjectAttributesModelSqlConn) FindByBusinessObjectIdWithFieldInfo(ctx context.Context, businessObjectId string) ([]*FieldWithAttrInfo, error) {
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

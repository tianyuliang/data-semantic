// Package business_object_attributes_temp 业务对象属性临时表Model (SqlConn实现)
package business_object_attributes_temp

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewBusinessObjectAttributesTempModelSqlConn 创建BusinessObjectAttributesTempModelSqlConn实例
func NewBusinessObjectAttributesTempModelSqlConn(conn sqlx.SqlConn) *BusinessObjectAttributesTempModelSqlConn {
	return &BusinessObjectAttributesTempModelSqlConn{conn: conn}
}

// BusinessObjectAttributesTempModelSqlConn BusinessObjectAttributesTempModel实现 (基于 go-zero SqlConn)
type BusinessObjectAttributesTempModelSqlConn struct {
	conn sqlx.SqlConn
}

// FindByBusinessObjectId 根据business_object_id查询属性列表
func (m *BusinessObjectAttributesTempModelSqlConn) FindByBusinessObjectId(ctx context.Context, businessObjectId string) ([]*BusinessObjectAttributesTemp, error) {
	var resp []*BusinessObjectAttributesTemp
	query := `SELECT id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes_temp
	           WHERE business_object_id = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp by business_object_id failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewAndVersion 根据form_view_id和version查询所有属性
func (m *BusinessObjectAttributesTempModelSqlConn) FindByFormViewAndVersion(ctx context.Context, formViewId string, version int) ([]*BusinessObjectAttributesTemp, error) {
	var resp []*BusinessObjectAttributesTemp
	query := `SELECT id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes_temp
	           WHERE form_view_id = ? AND version = ? AND deleted_at IS NULL ORDER BY id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// FindOneById 根据id查询属性
func (m *BusinessObjectAttributesTempModelSqlConn) FindOneById(ctx context.Context, id string) (*BusinessObjectAttributesTemp, error) {
	var resp BusinessObjectAttributesTemp
	query := `SELECT id, form_view_id, business_object_id, user_id, version, form_view_field_id, attr_name, created_at, updated_at, deleted_at
	           FROM t_business_object_attributes_temp
	           WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp by id failed: %w", err)
	}
	return &resp, nil
}

// Update 更新属性名称
func (m *BusinessObjectAttributesTempModelSqlConn) Update(ctx context.Context, data *BusinessObjectAttributesTemp) error {
	query := `UPDATE t_business_object_attributes_temp
	           SET attr_name = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.AttrName, data.Id)
	if err != nil {
		return fmt.Errorf("update business_object_attributes_temp failed: %w", err)
	}
	return nil
}

// UpdateBusinessObjectId 更新属性归属的业务对象
func (m *BusinessObjectAttributesTempModelSqlConn) UpdateBusinessObjectId(ctx context.Context, attributeId, businessObjectId string) error {
	query := `UPDATE t_business_object_attributes_temp
	           SET business_object_id = ?
	           WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, businessObjectId, attributeId)
	if err != nil {
		return fmt.Errorf("update business_object_id for attribute failed: %w", err)
	}
	return nil
}

// FieldWithAttrInfoTemp 属性关联字段信息（临时表）
type FieldWithAttrInfoTemp struct {
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
func (m *BusinessObjectAttributesTempModelSqlConn) FindByBusinessObjectIdWithFieldInfo(ctx context.Context, businessObjectId string) ([]*FieldWithAttrInfoTemp, error) {
	var resp []*FieldWithAttrInfoTemp
	query := `SELECT boat.id, boat.business_object_id, boat.form_view_field_id, boat.attr_name,
	           fvf.field_tech_name, fvf.business_name AS field_business_name, fvf.field_role, fvf.field_type
	           FROM t_business_object_attributes_temp boat
	           INNER JOIN t_form_view_field fvf ON boat.form_view_field_id = fvf.id
	           WHERE boat.business_object_id = ? AND boat.deleted_at IS NULL AND fvf.deleted_at IS NULL
	           ORDER BY boat.id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, businessObjectId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp with field info failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewAndVersionWithFieldInfo 根据form_view_id和version查询所有属性（包含字段信息）
func (m *BusinessObjectAttributesTempModelSqlConn) FindByFormViewAndVersionWithFieldInfo(ctx context.Context, formViewId string, version int) ([]*FieldWithAttrInfoTemp, error) {
	var resp []*FieldWithAttrInfoTemp
	query := `SELECT boat.id, boat.business_object_id, boat.form_view_field_id, boat.attr_name,
	           fvf.field_tech_name, fvf.business_name AS field_business_name, fvf.field_role, fvf.field_type
	           FROM t_business_object_attributes_temp boat
	           INNER JOIN t_form_view_field fvf ON boat.form_view_field_id = fvf.id
	           WHERE boat.form_view_id = ? AND boat.version = ? AND boat.deleted_at IS NULL AND fvf.deleted_at IS NULL
	           ORDER BY boat.id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId, version)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp with field info by form_view_id and version failed: %w", err)
	}
	return resp, nil
}

// FindByFormViewIdLatestWithFieldInfo 查询指定form_view_id的最新版本属性列表（包含字段信息）
func (m *BusinessObjectAttributesTempModelSqlConn) FindByFormViewIdLatestWithFieldInfo(ctx context.Context, formViewId string) ([]*FieldWithAttrInfoTemp, error) {
	var resp []*FieldWithAttrInfoTemp
	query := `SELECT boat.id, boat.business_object_id, boat.form_view_field_id, boat.attr_name,
	           fvf.field_tech_name, fvf.business_name AS field_business_name, fvf.field_role, fvf.field_type
	           FROM t_business_object_attributes_temp boat
	           INNER JOIN t_form_view_field fvf ON boat.form_view_field_id = fvf.id
	           WHERE boat.form_view_id = ? AND boat.version = (
	               SELECT MAX(version) FROM t_business_object_attributes_temp
	               WHERE form_view_id = ? AND deleted_at IS NULL
	           ) AND boat.deleted_at IS NULL AND fvf.deleted_at IS NULL
	           ORDER BY boat.id ASC`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, formViewId, formViewId)
	if err != nil {
		return nil, fmt.Errorf("find business_object_attributes_temp with field info by form_view_id latest failed: %w", err)
	}
	return resp, nil
}

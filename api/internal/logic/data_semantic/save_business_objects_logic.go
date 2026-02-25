// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/errorx"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存业务对象及属性
func NewSaveBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveBusinessObjectsLogic {
	return &SaveBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveBusinessObjectsLogic) SaveBusinessObjects(req *types.SaveBusinessObjectsReq) (resp *types.SaveBusinessObjectsResp, err error) {
	logx.Infof("SaveBusinessObjects called with type: %s, id: %s, name: %s",
		req.Type, req.Id, req.Name)

	// 1. 状态校验：只有状态 2（待确认）才能编辑
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("库表视图", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, errorx.NewInvalidUnderstandStatus(formViewData.UnderstandStatus)
	}

	// 2. 根据 type 决定更新业务对象还是属性
	if req.Type == "object" {
		err = l.saveBusinessObjectName(req.Id, req.Name)
		if err != nil {
			return nil, err
		}

	} else if req.Type == "attribute" {
		err = l.saveAttributeName(req.Id, req.Name)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errorx.Newf(errorx.ErrCodeInvalidParam, "无效的 type 参数: %s，必须是 'object' 或 'attribute'", req.Type)
	}

	// 注意：此操作不递增版本号，仅更新当前版本的临时数据

	resp = &types.SaveBusinessObjectsResp{
		Code: 0,
	}

	return resp, nil
}

// saveBusinessObjectName 保存业务对象名称
func (l *SaveBusinessObjectsLogic) saveBusinessObjectName(id, name string) error {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)

	// 先查询记录是否存在
	objData, err := businessObjectTempModel.FindOneById(l.ctx, id)
	if err != nil {
		return errorx.NewQueryFailed("业务对象", err)
	}

	// 名称重复校验：同一库表下不能有重复的业务对象名称
	err = l.checkDuplicateObjectName(objData.FormViewId, name, id)
	if err != nil {
		return err
	}

	// 更新名称
	objData.ObjectName = name
	err = businessObjectTempModel.Update(l.ctx, objData)
	if err != nil {
		return errorx.NewUpdateFailed("业务对象名称", err)
	}

	logx.WithContext(l.ctx).Infof("Updated business object name: id=%s, name=%s", id, name)
	return nil
}

// saveAttributeName 保存属性名称
func (l *SaveBusinessObjectsLogic) saveAttributeName(id, name string) error {
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)

	// 先查询记录是否存在
	attrData, err := businessObjectAttrTempModel.FindOneById(l.ctx, id)
	if err != nil {
		return errorx.NewQueryFailed("业务对象属性", err)
	}

	// 名称重复校验：同一业务对象下，同一字段的属性名称不能重复
	err = l.checkDuplicateAttrName(attrData.BusinessObjectId, attrData.FormViewFieldId, name, id)
	if err != nil {
		return err
	}

	// 更新名称
	attrData.AttrName = name
	err = businessObjectAttrTempModel.Update(l.ctx, attrData)
	if err != nil {
		return errorx.NewUpdateFailed("属性名称", err)
	}

	logx.WithContext(l.ctx).Infof("Updated attribute name: id=%s, name=%s", id, name)
	return nil
}

// checkDuplicateObjectName 检查业务对象名称是否重复
func (l *SaveBusinessObjectsLogic) checkDuplicateObjectName(formViewId, name, excludeId string) error {
	var count int64
	query := `SELECT COUNT(*) FROM t_business_object_temp WHERE form_view_id = ? AND object_name = ? AND id != ? AND deleted_at IS NULL`
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &count, query, formViewId, name, excludeId)
	if err != nil {
		return errorx.NewQueryFailed("业务对象名称重复校验", err)
	}
	if count > 0 {
		return errorx.NewDuplicateName("业务对象", name)
	}
	return nil
}

// checkDuplicateAttrName 检查属性名称是否重复（同一业务对象下，同一字段的属性名称不能重复）
func (l *SaveBusinessObjectsLogic) checkDuplicateAttrName(businessObjectId, formViewFieldId, name, excludeId string) error {
	var count int64
	query := `SELECT COUNT(*) FROM t_business_object_attributes_temp WHERE business_object_id = ? AND form_view_field_id = ? AND attr_name = ? AND id != ? AND deleted_at IS NULL`
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &count, query, businessObjectId, formViewFieldId, name, excludeId)
	if err != nil {
		return errorx.NewQueryFailed("属性名称重复校验", err)
	}
	if count > 0 {
		return errorx.NewDuplicateName("属性", name)
	}
	return nil
}

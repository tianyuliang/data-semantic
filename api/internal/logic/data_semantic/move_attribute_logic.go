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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MoveAttributeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调整属性归属业务对象
func NewMoveAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveAttributeLogic {
	return &MoveAttributeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MoveAttributeLogic) MoveAttribute(req *types.MoveAttributeReq) (resp *types.MoveAttributeResp, err error) {
	logx.Infof("MoveAttribute called with id: %s, attributeId: %s, targetObjectUuid: %s",
		req.Id, req.AttributeId, req.TargetObjectUuid)

	// 1. 状态校验：只有状态 2（待确认）才能编辑
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("库表视图", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, errorx.NewInvalidUnderstandStatus(formViewData.UnderstandStatus)
	}

	// 2. 使用事务执行调整操作（保证原子性）
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
		businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)

		// 3. 验证目标业务对象是否存在
		targetObject, err := businessObjectTempModel.FindOneById(ctx, req.TargetObjectUuid)
		if err != nil {
			return errorx.NewQueryFailed("目标业务对象", err)
		}
		logx.WithContext(ctx).Infof("Found target business object: id=%s, name=%s", targetObject.Id, targetObject.ObjectName)

		// 4. 验证属性是否存在
		attribute, err := businessObjectAttrTempModel.FindOneById(ctx, req.AttributeId)
		if err != nil {
			return errorx.NewQueryFailed("属性", err)
		}
		logx.WithContext(ctx).Infof("Found attribute: id=%s, name=%s, currentObjectId=%s", attribute.Id, attribute.AttrName, attribute.BusinessObjectId)

		// 5. 校验 form_view_id：属性和目标业务对象必须属于同一库表
		if attribute.FormViewId != req.Id {
			return errorx.Newf(errorx.ErrCodeInvalidParam, "属性不属于当前库表: %s", attribute.FormViewId)
		}
		if targetObject.FormViewId != req.Id {
			return errorx.Newf(errorx.ErrCodeInvalidParam, "目标业务对象不属于当前库表: %s", targetObject.FormViewId)
		}

		// 6. 检查是否自移动（将属性移动到当前归属的业务对象）
		if attribute.BusinessObjectId == req.TargetObjectUuid {
			return errorx.Newf(errorx.ErrCodeInvalidParam, "属性已归属到该业务对象，无需移动")
		}

		// 7. 检查目标业务对象下，同一字段是否已存在同名属性
		err = l.checkDuplicateAttrInTarget(businessObjectAttrTempModel, req.TargetObjectUuid, attribute.FormViewFieldId, attribute.AttrName, attribute.Id)
		if err != nil {
			return err
		}

		// 8. 更新属性的 business_object_id
		err = businessObjectAttrTempModel.UpdateBusinessObjectId(ctx, req.AttributeId, req.TargetObjectUuid)
		if err != nil {
			return errorx.NewUpdateFailed("属性归属", err)
		}

		logx.WithContext(ctx).Infof("Moved attribute %s from object %s to object %s",
			req.AttributeId, attribute.BusinessObjectId, req.TargetObjectUuid)

		return nil
	})
	if err != nil {
		return nil, err
	}

	resp = &types.MoveAttributeResp{
		AttributeId:      req.AttributeId,
		BusinessObjectId: req.TargetObjectUuid,
	}

	return resp, nil
}

// checkDuplicateAttrInTarget 检查目标业务对象下，同一字段是否已存在同名属性
func (l *MoveAttributeLogic) checkDuplicateAttrInTarget(model business_object_attributes_temp.BusinessObjectAttributesTempModel, businessObjectId, formViewFieldId, attrName, excludeAttrId string) error {
	targetAttrs, err := model.FindByBusinessObjectId(l.ctx, businessObjectId)
	if err != nil {
		return errorx.NewQueryFailed("目标业务对象的属性列表", err)
	}

	for _, targetAttr := range targetAttrs {
		// 排除自身（如果目标业务对象就是当前归属的业务对象）
		if targetAttr.Id == excludeAttrId {
			continue
		}
		// 同一业务对象下，同一字段的属性名称不能重复
		if targetAttr.FormViewFieldId == formViewFieldId && targetAttr.AttrName == attrName {
			return errorx.NewDuplicateName("目标业务对象下该字段的属性", attrName)
		}
	}
	return nil
}

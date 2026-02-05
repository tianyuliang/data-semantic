// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
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
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, fmt.Errorf("当前状态不允许编辑，当前状态: %d，仅状态 2 (待确认) 可编辑", formViewData.UnderstandStatus)
	}

	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlConn(l.svcCtx.DB)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlConn(l.svcCtx.DB)

	// 2. 验证目标业务对象是否存在
	targetObject, err := businessObjectTempModel.FindOneById(l.ctx, req.TargetObjectUuid)
	if err != nil {
		return nil, fmt.Errorf("查询目标业务对象失败: %w", err)
	}
	logx.WithContext(l.ctx).Infof("Found target business object: id=%s, name=%s", targetObject.Id, targetObject.ObjectName)

	// 3. 验证属性是否存在
	attribute, err := businessObjectAttrTempModel.FindOneById(l.ctx, req.AttributeId)
	if err != nil {
		return nil, fmt.Errorf("查询属性失败: %w", err)
	}
	logx.WithContext(l.ctx).Infof("Found attribute: id=%s, currentObjectId=%s", attribute.Id, attribute.BusinessObjectId)

	// 4. 更新属性的 business_object_id
	err = businessObjectAttrTempModel.UpdateBusinessObjectId(l.ctx, req.AttributeId, req.TargetObjectUuid)
	if err != nil {
		return nil, fmt.Errorf("更新属性归属失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("Moved attribute %s from object %s to object %s",
		req.AttributeId, attribute.BusinessObjectId, req.TargetObjectUuid)

	resp = &types.MoveAttributeResp{
		AttributeId:      req.AttributeId,
		BusinessObjectId: req.TargetObjectUuid,
	}

	return resp, nil
}

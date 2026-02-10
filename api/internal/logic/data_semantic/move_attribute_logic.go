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

	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)

	// 1. 验证目标业务对象是否存在
	targetObject, err := businessObjectTempModel.FindOneById(l.ctx, req.TargetObjectUuid)
	if err != nil {
		return nil, fmt.Errorf("查询目标业务对象失败: %w", err)
	}
	logx.WithContext(l.ctx).Infof("Found target business object: id=%s, name=%s", targetObject.Id, targetObject.ObjectName)

	// 2. 验证属性是否存在
	attribute, err := businessObjectAttrTempModel.FindOneById(l.ctx, req.AttributeId)
	if err != nil {
		return nil, fmt.Errorf("查询属性失败: %w", err)
	}
	logx.WithContext(l.ctx).Infof("Found attribute: id=%s, name=%s, currentObjectId=%s", attribute.Id, attribute.AttrName, attribute.BusinessObjectId)

	// 3. 检查目标业务对象下是否已存在同名属性
	targetAttrs, err := businessObjectAttrTempModel.FindByBusinessObjectId(l.ctx, req.TargetObjectUuid)
	if err != nil {
		return nil, fmt.Errorf("查询目标业务对象的属性列表失败: %w", err)
	}

	for _, targetAttr := range targetAttrs {
		if targetAttr.AttrName == attribute.AttrName {
			return nil, fmt.Errorf("目标业务对象下已存在同名属性: %s", attribute.AttrName)
		}
	}

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

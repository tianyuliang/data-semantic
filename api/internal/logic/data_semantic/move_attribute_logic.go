// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

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

	// 1. 验证目标业务对象是否存在
	// TODO: 查询 t_business_object_temp 表
	// SELECT id FROM t_business_object_temp WHERE id = ? AND deleted_at IS NULL LIMIT 1
	// targetObject, err := findBusinessObject(l.ctx, req.TargetObjectUuid)
	// if err != nil || targetObject == nil {
	//     return nil, errors.New("目标业务对象不存在")
	// }

	// 2. 更新属性的 business_object_id
	// TODO: 更新 t_business_object_attributes_temp 表
	// UPDATE t_business_object_attributes_temp
	// SET business_object_id = ?
	// WHERE id = ? AND deleted_at IS NULL
	logx.Infof("Moving attribute %s to business object %s", req.AttributeId, req.TargetObjectUuid)

	resp = &types.MoveAttributeResp{
		AttributeId:      req.AttributeId,
		BusinessObjectId: req.TargetObjectUuid,
	}

	return resp, nil
}

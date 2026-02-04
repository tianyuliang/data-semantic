// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

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

	// 根据 type 决定更新业务对象还是属性
	if req.Type == "object" {
		// 更新业务对象名称
		// TODO: 更新 t_business_object_temp 表
		// UPDATE t_business_object_temp SET object_name = ? WHERE id = ? AND deleted_at IS NULL
		logx.Infof("Updating business object name: id=%s, name=%s", req.Id, req.Name)
	} else if req.Type == "attribute" {
		// 更新属性名称
		// TODO: 更新 t_business_object_attributes_temp 表
		// UPDATE t_business_object_attributes_temp SET attr_name = ? WHERE id = ? AND deleted_at IS NULL
		logx.Infof("Updating attribute name: id=%s, name=%s", req.Id, req.Name)
	}

	// 注意：此操作不递增版本号，仅更新当前版本的临时数据

	resp = &types.SaveBusinessObjectsResp{
		Code: 0,
	}

	return resp, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSemanticInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存库表信息补全数据
func NewSaveSemanticInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSemanticInfoLogic {
	return &SaveSemanticInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveSemanticInfoLogic) SaveSemanticInfo(req *types.SaveSemanticInfoReq) (resp *types.SaveSemanticInfoResp, err error) {
	logx.Infof("SaveSemanticInfo called with id: %s, tableData: %v, fieldData: %v",
		req.Id, req.TableData != nil, req.FieldData != nil)

	// 1. 如果提供了 TableData，更新 t_form_view_info_temp
	if req.TableData != nil {
		// TODO: 更新库表信息到临时表
		// UPDATE t_form_view_info_temp
		// SET table_business_name = ?, table_description = ?
		// WHERE id = ? AND deleted_at IS NULL
		logx.Infof("Updating table info: id=%s", *req.TableData.Id)
	}

	// 2. 如果提供了 FieldData，更新 t_form_view_field_info_temp
	if req.FieldData != nil {
		// TODO: 更新字段信息到临时表
		// UPDATE t_form_view_field_info_temp
		// SET field_business_name = ?, field_role = ?, field_description = ?
		// WHERE id = ? AND deleted_at IS NULL
		logx.Infof("Updating field info: id=%s", *req.FieldData.Id)
	}

	// 注意：此操作不递增版本号，仅更新当前版本的临时数据

	resp = &types.SaveSemanticInfoResp{
		Code: 0,
	}

	return resp, nil
}

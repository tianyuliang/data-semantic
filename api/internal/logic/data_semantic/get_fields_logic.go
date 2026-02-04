// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询字段语义补全数据
func NewGetFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFieldsLogic {
	return &GetFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFieldsLogic) GetFields(req *types.GetFieldsReq) (resp *types.GetFieldsResp, err error) {
	logx.Infof("GetFields called with id: %s, keyword: %v, only_incomplete: %v",
		req.Id, req.Keyword, req.OnlyIncomplete)

	// 1. 查询 form_view 的 understand_status 和 current_version
	// TODO: 查询 form_view 表
	// SELECT understand_status, current_version FROM form_view WHERE id = ?
	// understandStatus, currentVersion := ...

	// 临时返回值 (用于测试)
	understandStatus := int8(0)
	currentVersion := 0
	tableTechName := "test_table"

	// 2. 根据状态返回不同数据源
	if understandStatus == 0 {
		// 状态 0: 未理解，返回空数据
		resp = &types.GetFieldsResp{
			CurrentVersion:    0,
			TableBusinessName: nil,
			TableTechName:     tableTechName,
			TableDescription:  nil,
			Fields:            []types.FieldSemanticInfo{},
		}
		return resp, nil
	}

	// 3. 状态 2 (待确认) 或 3 (已完成) - 查询临时表或正式表
	// TODO: 根据状态查询 t_form_view_info_temp 或正式表
	// tableInfo, err := queryTableInfo(l.ctx, req.Id, currentVersion)

	// TODO: 查询字段列表 t_form_view_field_info_temp 或正式表
	// fields, err := queryFields(l.ctx, req.Id, currentVersion, req.Keyword, req.OnlyIncomplete)

	// 4. 过滤 only_incomplete (只返回未补全的字段)
	// if req.OnlyIncomplete != nil && *req.OnlyIncomplete {
	//     fields = filterIncompleteFields(fields)
	// }

	// 临时返回值 (用于测试)
	resp = &types.GetFieldsResp{
		CurrentVersion:    currentVersion,
		TableBusinessName: nil,
		TableTechName:     tableTechName,
		TableDescription:  nil,
		Fields:            []types.FieldSemanticInfo{},
	}

	return resp, nil
}

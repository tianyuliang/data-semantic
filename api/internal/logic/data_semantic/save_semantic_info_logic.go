// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_field_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

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

	// 1. 状态校验：只有状态 2（待确认）才能编辑
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, fmt.Errorf("当前状态不允许编辑，当前状态: %d，仅状态 2 (待确认) 可编辑", formViewData.UnderstandStatus)
	}

	// 2. 如果提供了 TableData，更新 t_form_view_info_temp
	if req.TableData != nil {
		formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSqlConn(l.svcCtx.DB)

		// 构建更新数据
		tableInfoTemp := &form_view_info_temp.FormViewInfoTemp{
			Id:                *req.TableData.Id,
			FormViewId:        req.Id,
			TableBusinessName:  req.TableData.TableBusinessName,
			TableDescription:   req.TableData.TableDescription,
		}

		err = formViewInfoTempModel.Update(l.ctx, tableInfoTemp)
		if err != nil {
			return nil, fmt.Errorf("更新库表信息失败: %w", err)
		}
		logx.WithContext(l.ctx).Infof("Updated table info: id=%s", *req.TableData.Id)
	}

	// 3. 如果提供了 FieldData，更新 t_form_view_field_info_temp
	if req.FieldData != nil {
		formViewFieldInfoTempModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSqlConn(l.svcCtx.DB)

		// 构建更新数据
		fieldInfoTemp := &form_view_field_info_temp.FormViewFieldInfoTemp{
			Id:                *req.FieldData.Id,
			FieldBusinessName: req.FieldData.FieldBusinessName,
			FieldRole:        req.FieldData.FieldRole,
			FieldDescription:  req.FieldData.FieldDescription,
		}

		err = formViewFieldInfoTempModel.Update(l.ctx, fieldInfoTemp)
		if err != nil {
			return nil, fmt.Errorf("更新字段信息失败: %w", err)
		}
		logx.WithContext(l.ctx).Infof("Updated field info: id=%s", *req.FieldData.Id)
	}

	resp = &types.SaveSemanticInfoResp{
		Code: 0,
	}

	return resp, nil
}

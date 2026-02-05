// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除识别结果
func NewDeleteBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBusinessObjectsLogic {
	return &DeleteBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBusinessObjectsLogic) DeleteBusinessObjects(req *types.DeleteBusinessObjectsReq) (resp *types.DeleteBusinessObjectsResp, err error) {
	logx.Infof("DeleteBusinessObjects called with id: %s", req.Id)

	// 1. 状态校验 (仅允许状态 2 删除)
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, fmt.Errorf("当前状态不允许删除，当前状态: %d，仅状态 2 (待确认) 可删除", formViewData.UnderstandStatus)
	}

	// 2. 逻辑删除临时表数据
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlConn(l.svcCtx.DB)
	err = businessObjectTempModel.DeleteByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("删除临时表业务对象数据失败: %w", err)
	}

	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlConn(l.svcCtx.DB)
	err = businessObjectAttrTempModel.DeleteByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("删除临时表属性数据失败: %w", err)
	}

	// 3. 检查正式表是否有数据
	businessObjectModel := business_object.NewBusinessObjectModelSqlConn(l.svcCtx.DB)
	formalDataCount, err := businessObjectModel.CountByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询正式表数据失败: %w", err)
	}

	// 4. 根据正式表数据决定最终状态
	var newStatus int8
	if formalDataCount > 0 {
		// 正式表有数据，保持状态 3 (已完成)
		newStatus = form_view.StatusCompleted
	} else {
		// 正式表无数据，回退到状态 0 (未理解)
		newStatus = form_view.StatusNotUnderstanding
	}

	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, newStatus)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("Delete business objects successful: form_view_id=%s, new_status=%d, formal_data_count=%d",
		req.Id, newStatus, formalDataCount)

	resp = &types.DeleteBusinessObjectsResp{
		Success: true,
	}

	return resp, nil
}

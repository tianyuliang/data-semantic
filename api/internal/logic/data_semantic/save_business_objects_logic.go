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
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, fmt.Errorf("当前状态不允许编辑，当前状态: %d，仅状态 2 (待确认) 可编辑", formViewData.UnderstandStatus)
	}

	// 2. 根据 type 决定更新业务对象还是属性
	if req.Type == "object" {
		// 更新业务对象名称
		businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)

		// 先查询记录是否存在
		objData, err := businessObjectTempModel.FindOneById(l.ctx, req.Id)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象失败: %w", err)
		}

		// 更新名称
		objData.ObjectName = req.Name
		err = businessObjectTempModel.Update(l.ctx, objData)
		if err != nil {
			return nil, fmt.Errorf("更新业务对象名称失败: %w", err)
		}

		logx.WithContext(l.ctx).Infof("Updated business object name: id=%s, name=%s", req.Id, req.Name)

	} else if req.Type == "attribute" {
		// 更新属性名称
		businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSqlx(l.svcCtx.DB)

		// 先查询记录是否存在
		attrData, err := businessObjectAttrTempModel.FindOneById(l.ctx, req.Id)
		if err != nil {
			return nil, fmt.Errorf("查询业务对象属性失败: %w", err)
		}

		// 更新名称
		attrData.AttrName = req.Name
		err = businessObjectAttrTempModel.Update(l.ctx, attrData)
		if err != nil {
			return nil, fmt.Errorf("更新属性名称失败: %w", err)
		}

		logx.WithContext(l.ctx).Infof("Updated attribute name: id=%s, name=%s", req.Id, req.Name)
	} else {
		return nil, fmt.Errorf("无效的 type 参数: %s，必须是 'object' 或 'attribute'", req.Type)
	}

	// 注意：此操作不递增版本号，仅更新当前版本的临时数据

	resp = &types.SaveBusinessObjectsResp{
		Code: 0,
	}

	return resp, nil
}

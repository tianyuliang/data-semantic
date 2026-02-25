// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/errorx"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SubmitUnderstandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 提交确认理解数据
func NewSubmitUnderstandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitUnderstandingLogic {
	return &SubmitUnderstandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitUnderstandingLogic) SubmitUnderstanding(req *types.SubmitUnderstandingReq) (resp *types.SubmitUnderstandingResp, err error) {
	logx.Infof("SubmitUnderstanding called with id: %s", req.Id)

	// 1. 状态校验 (仅允许状态 2 提交)
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("库表视图", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, errorx.NewInvalidUnderstandStatus(formViewData.UnderstandStatus)
	}

	// 2. 获取当前版本号
	businessObjectTempModel := business_object.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewQueryFailed("当前版本号", err)
	}
	if latestVersion == 0 {
		return nil, errorx.Newf(errorx.ErrCodeInvalidParam, "没有可提交的数据，版本号为0")
	}

	// 3. 开启事务处理
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 使用事务的 Session 创建正式表 model 实例
		businessObjectModel := business_object.NewBusinessObjectModelSession(session)
		businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSession(session)
		formViewModelSession := form_view.NewFormViewModelSession(session)

		// ========== 合并业务对象（基于业务主键：form_view_id + object_name）==========

		objInserted, objUpdated, objDeleted, err := businessObjectModel.MergeFromTemp(ctx, req.Id, latestVersion)
		if err != nil {
			return errorx.NewUpdateFailed("业务对象", err)
		}
		logx.WithContext(ctx).Infof("Merged business objects: inserted=%d, updated=%d, deleted=%d", objInserted, objUpdated, objDeleted)

		// ========== 合并业务对象属性（基于业务对象匹配 + form_view_field_id）==========

		attrInserted, attrUpdated, attrDeleted, err := businessObjectAttrModel.MergeFromTemp(ctx, req.Id, latestVersion)
		if err != nil {
			return errorx.NewUpdateFailed("属性", err)
		}
		logx.WithContext(ctx).Infof("Merged attributes: inserted=%d, updated=%d, deleted=%d", attrInserted, attrUpdated, attrDeleted)

		// ========== 更新 form_view 状态为 3 (已完成) ==========

		err = formViewModelSession.UpdateUnderstandStatus(ctx, req.Id, form_view.StatusCompleted)
		if err != nil {
			return errorx.NewUpdateFailed("理解状态", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logx.WithContext(l.ctx).Infof("Submit understanding successful: form_view_id=%s, version=%d", req.Id, latestVersion)

	resp = &types.SubmitUnderstandingResp{
		Success: true,
	}

	return resp, nil
}

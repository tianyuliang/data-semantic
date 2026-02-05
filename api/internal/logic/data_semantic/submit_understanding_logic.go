// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm {
		return nil, fmt.Errorf("当前状态不允许提交，当前状态: %d，仅状态 2 (待确认) 可提交", formViewData.UnderstandStatus)
	}

	// 2. 获取当前版本号
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlConn(l.svcCtx.DB)
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询当前版本号失败: %w", err)
	}
	if latestVersion == 0 {
		return nil, fmt.Errorf("没有可提交的数据，版本号为0")
	}

	// 3. 开启事务处理
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 使用事务的 Session 创建正式表 model 实例
		businessObjectModel := business_object.NewBusinessObjectModelSession(session)
		businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSession(session)

		// 4. 删除正式表旧数据
		err = businessObjectModel.DeleteByFormViewId(ctx, req.Id)
		if err != nil {
			return fmt.Errorf("删除旧业务对象数据失败: %w", err)
		}

		err = businessObjectAttrModel.DeleteByFormViewId(ctx, req.Id)
		if err != nil {
			return fmt.Errorf("删除旧属性数据失败: %w", err)
		}

		// 5. 复制业务对象数据 (临时表 → 正式表)
		objectCount, err := businessObjectModel.BatchInsertFromTemp(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("复制业务对象数据失败: %w", err)
		}

		// 6. 复制属性数据 (临时表 → 正式表)
		attributeCount, err := businessObjectAttrModel.BatchInsertFromTemp(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("复制属性数据失败: %w", err)
		}

		logx.WithContext(ctx).Infof("Copied data from temp to formal: objects=%d, attributes=%d", objectCount, attributeCount)

		// 注意：状态更新需要在事务外进行，因为 form_view 表不参与这个事务
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("事务执行失败: %w", err)
	}

	// 7. 更新 form_view 状态为 3 (已完成)
	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, form_view.StatusCompleted)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("Submit understanding successful: form_view_id=%s, version=%d", req.Id, latestVersion)

	resp = &types.SubmitUnderstandingResp{
		Success: true,
	}

	return resp, nil
}

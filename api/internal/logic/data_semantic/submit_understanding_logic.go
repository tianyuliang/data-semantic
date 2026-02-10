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
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
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
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询当前版本号失败: %w", err)
	}
	if latestVersion == 0 {
		return nil, fmt.Errorf("没有可提交的数据，版本号为0")
	}

	// 3. 开启事务处理
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 使用事务的 Session 创建正式表和临时表 model 实例
		businessObjectModel := business_object.NewBusinessObjectModelSession(session)
		businessObjectAttrModel := business_object_attributes.NewBusinessObjectAttributesModelSession(session)
		businessObjectTempModelSession := business_object_temp.NewBusinessObjectTempModelSession(session)
		businessObjectAttrTempModelSession := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)

		// ========== 增量更新业务对象 ==========

		// 3.1 更新已有记录（通过 formal_id 匹配）
		updatedCount, err := businessObjectModel.UpdateByFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("更新业务对象失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Updated business objects: %d", updatedCount)

		// 3.2 新增记录（formal_id 为 NULL 的记录）
		insertedCount, err := businessObjectModel.InsertFromTempWithoutFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("新增业务对象失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Inserted business objects: %d", insertedCount)

		// 3.3 删除不再需要的记录（正式表有但临时表没有的）
		deletedCount, err := businessObjectModel.DeleteNotInFormalIdList(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("删除业务对象失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Deleted business objects: %d", deletedCount)

		// 3.4 回写 formal_id 到临时表（为下次提交准备）
		formalIdUpdatedCount, err := businessObjectTempModelSession.UpdateFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("回写业务对象 formal_id 失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Updated formal_id in temp table: %d", formalIdUpdatedCount)

		// ========== 增量更新业务对象属性 ==========

		// 3.5 更新已有属性（通过 formal_id 匹配）
		attrUpdatedCount, err := businessObjectAttrModel.UpdateByFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("更新属性失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Updated attributes: %d", attrUpdatedCount)

		// 3.6 新增属性（formal_id 为 NULL 的记录）
		attrInsertedCount, err := businessObjectAttrModel.InsertFromTempWithoutFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("新增属性失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Inserted attributes: %d", attrInsertedCount)

		// 3.7 删除不再需要的属性（正式表有但临时表没有的）
		attrDeletedCount, err := businessObjectAttrModel.DeleteNotInFormalIdList(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("删除属性失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Deleted attributes: %d", attrDeletedCount)

		// 3.8 回写 formal_id 到属性临时表
		attrFormalIdUpdatedCount, err := businessObjectAttrTempModelSession.UpdateFormalId(ctx, req.Id, latestVersion)
		if err != nil {
			return fmt.Errorf("回写属性 formal_id 失败: %w", err)
		}
		logx.WithContext(ctx).Infof("Updated formal_id in attributes temp table: %d", attrFormalIdUpdatedCount)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("事务执行失败: %w", err)
	}

	// 4. 更新 form_view 状态为 3 (已完成)
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

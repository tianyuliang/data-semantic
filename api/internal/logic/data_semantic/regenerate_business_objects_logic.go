// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/internal/pkg/aiservice"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateBusinessObjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重新识别业务对象
func NewRegenerateBusinessObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateBusinessObjectsLogic {
	return &RegenerateBusinessObjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegenerateBusinessObjectsLogic) RegenerateBusinessObjects(req *types.RegenerateBusinessObjectsReq) (resp *types.RegenerateBusinessObjectsResp, err error) {
	logx.Infof("RegenerateBusinessObjects called with id: %s", req.Id)

	// 1. 查询 form_view 状态
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	// 2. 状态校验 (仅允许状态 2 或 3)
	if formViewData.UnderstandStatus != form_view.StatusPendingConfirm && formViewData.UnderstandStatus != form_view.StatusCompleted {
		return nil, fmt.Errorf("当前状态不允许重新识别，当前状态: %d，仅状态 2 (待确认) 或 3 (已完成) 可重新识别", formViewData.UnderstandStatus)
	}

	// 3. 限流检查（1秒窗口，防止重复点击）
	limiter := l.svcCtx.GetRateLimiter(req.Id)
	if !limiter.Allow() {
		return nil, fmt.Errorf("操作过于频繁，请稍后再试")
	}

	// 4. 查询字段数据（完整信息，包含业务名称和角色）
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fields, err := formViewFieldModel.FindFullByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询字段列表失败: %w", err)
	}

	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlx(l.svcCtx.DB)

	// 5. 查询当前版本号（用于版本控制，后续扩展使用）
	_, err = businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询当前版本号失败: %w", err)
	}

	// 6. 更新状态为 1（理解中）
	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, form_view.StatusUnderstanding)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	// 7. 调用 AI 服务 HTTP API（同步调用）
	if err := l.callAIService(req.Id, formViewData, fields); err != nil {
		// 调用失败，回退状态
		_ = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, formViewData.UnderstandStatus)
		return nil, fmt.Errorf("调用 AI 服务失败: %w", err)
	}

	resp = &types.RegenerateBusinessObjectsResp{
		ObjectCount:    0, // 实际数量由 AI 识别完成后写入
		AttributeCount: len(fields),
	}

	return resp, nil
}

// callAIService 调用 AI 服务 HTTP API
func (l *RegenerateBusinessObjectsLogic) callAIService(formViewId string, formViewData *form_view.FormView, fields []*form_view_field.FormViewField) error {
	// 使用 builder 构建 FormView
	aiFormView := aiservice.BuildFormView(formViewId, formViewData, fields)

	// 生成 message_id
	messageID := uuid.New().String()

	logx.WithContext(l.ctx).Infof("调用 AI 服务: request_type=%s, field_count=%d", aiservice.RequestTypeRegenerateBusinessObjects, len(fields))

	// 调用 AI 服务
	aiResponse, err := l.svcCtx.AIClient.Call(aiservice.RequestTypeRegenerateBusinessObjects, messageID, aiFormView)
	if err != nil {
		return fmt.Errorf("调用 AI 服务失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("AI 服务调用成功: task_id=%s, status=%s, message=%s",
		aiResponse.TaskID, aiResponse.Status, aiResponse.Message)

	return nil
}

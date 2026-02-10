// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateUnderstandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 一键生成理解数据
func NewGenerateUnderstandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateUnderstandingLogic {
	return &GenerateUnderstandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateUnderstandingLogic) GenerateUnderstanding(req *types.GenerateUnderstandingReq) (resp *types.GenerateUnderstandingResp, err error) {
	logx.Infof("GenerateUnderstanding called with id: %s", req.Id)

	// 1. 状态校验：只有状态 0（未理解）或 3（已完成）才允许生成
	formViewModel := form_view.NewFormViewModel(l.svcCtx.DB)
	formViewData, err := formViewModel.FindOneById(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询库表视图失败: %w", err)
	}

	currentStatus := formViewData.UnderstandStatus
	if currentStatus != form_view.StatusNotUnderstanding && currentStatus != form_view.StatusCompleted {
		return nil, fmt.Errorf("当前状态不允许生成理解数据，当前状态: %d", currentStatus)
	}

	// 2. 限流检查（1秒窗口，防止重复点击）
	limiter := l.svcCtx.GetRateLimiter(req.Id)
	if !limiter.Allow() {
		return nil, fmt.Errorf("操作过于频繁，请稍后再试")
	}

	// 3. 更新状态为 1（理解中）
	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, form_view.StatusUnderstanding)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	// 4. 查询字段数据用于 AI 服务请求（完整信息，包含业务名称和角色）
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fields, err := formViewFieldModel.FindFullByFormViewId(l.ctx, req.Id)
	if err != nil {
		// 查询失败，回退状态
		_ = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, currentStatus)
		return nil, fmt.Errorf("查询字段数据失败: %w", err)
	}

	// 5. 调用 AI 服务 HTTP API（同步调用）
	if err := l.callAIService(req.Id, formViewData, fields); err != nil {
		// 调用失败，回退状态
		_ = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, currentStatus)
		return nil, fmt.Errorf("调用 AI 服务失败: %w", err)
	}

	// 6. 返回新状态
	resp = &types.GenerateUnderstandingResp{
		UnderstandStatus: form_view.StatusUnderstanding,
	}

	return resp, nil
}

// callAIService 调用 AI 服务 HTTP API
func (l *GenerateUnderstandingLogic) callAIService(formViewId string, formViewData *form_view.FormView, fields []*form_view_field.FormViewField) error {
	// 构建字段列表
	formViewFields := make([]map[string]interface{}, 0, len(fields))
	for _, f := range fields {
		fieldRole := ""
		if f.FieldRole != nil {
			// 将 int8 转换为字符串
			fieldRole = fmt.Sprintf("%d", *f.FieldRole)
		}

		fieldDesc := ""
		if f.FieldDescription != nil {
			fieldDesc = *f.FieldDescription
		}

		fieldBusinessName := ""
		if f.FieldBusinessName != nil {
			fieldBusinessName = *f.FieldBusinessName
		}

		formViewFields = append(formViewFields, map[string]interface{}{
			"form_view_field_id":           f.Id,
			"form_view_field_technical_name": f.FieldTechName,
			"form_view_field_business_name":  fieldBusinessName,
			"form_view_field_type":          f.FieldType,
			"form_view_field_role":          fieldRole,
			"form_view_field_desc":          fieldDesc,
		})
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"message_id":   uuid.New().String(),
		"request_type": "full_understanding",
		"form_view": map[string]interface{}{
			"form_view_id":               formViewId,
			"form_view_technical_name":   formViewData.TechnicalName,
			"form_view_business_name":    formViewData.BusinessName,
			"form_view_desc":             formViewData.Description,
			"form_view_fields":           formViewFields,
		},
	}

	// 调用 AI 服务
	aiResponse, err := l.svcCtx.CallAIService("full_understanding", requestBody)
	if err != nil {
		return fmt.Errorf("调用 AI 服务失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("AI 服务调用成功: task_id=%s, status=%s, message=%s",
		aiResponse.TaskID, aiResponse.Status, aiResponse.Message)

	return nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"
	"time"

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

	// 2. Redis 限流检查（1秒窗口，防止重复点击）
	if l.svcCtx.Redis != nil {
		rateLimitKey := fmt.Sprintf("rate_limit:generate:%s", req.Id)
		// 使用 Redis SETNX 实现简单的滑动窗口限流
		allowed, err := l.checkRateLimit(rateLimitKey, time.Second)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("限流检查失败: %v", err)
		}
		if !allowed {
			return nil, fmt.Errorf("操作过于频繁，请稍后再试")
		}
	}

	// 3. 更新状态为 1（理解中）
	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, form_view.StatusUnderstanding)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	// 4. 查询字段数据用于 AI 服务请求
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fields, err := formViewFieldModel.FindByFormViewId(l.ctx, req.Id)
	if err != nil {
		// 记录错误但不中断流程，状态已更新
		logx.WithContext(l.ctx).Errorf("查询字段数据失败: %v", err)
	}

	// 5. 调用 AI 服务 HTTP API
	go func() {
		// 异步调用，避免阻塞主流程
		if err := l.callAIService(req.Id, formViewData, fields); err != nil {
			logx.WithContext(l.ctx).Errorf("调用 AI 服务失败: %v", err)
			// 状态保持为 1-理解中，由 Kafka 消费者处理失败后更新为 5-理解失败
		}
	}()

	// 6. 返回新状态
	resp = &types.GenerateUnderstandingResp{
		UnderstandStatus: form_view.StatusUnderstanding,
	}

	return resp, nil
}

// checkRateLimit 使用 Redis 检查限流
func (l *GenerateUnderstandingLogic) checkRateLimit(key string, window time.Duration) (bool, error) {
	// 使用 SETNX + EXPIRE 实现简单的限流
	// SET key value NX
	ok, err := l.svcCtx.Redis.SetnxCtx(l.ctx, key, "1")
	if err != nil {
		return false, err
	}
	if ok {
		// 设置过期时间
		_ = l.svcCtx.Redis.Expire(key, int(window.Seconds()))
	}
	return ok, nil
}

// callAIService 调用 AI 服务 HTTP API
func (l *GenerateUnderstandingLogic) callAIService(formViewId string, formViewData *form_view.FormViewDataBase, fields []*form_view_field.FormViewFieldBase) error {
	// 构建字段列表
	formViewFields := make([]map[string]interface{}, 0, len(fields))
	for _, f := range fields {
		fieldRole := ""
		if f.FieldRole != nil {
			// 将 int8 转换为字符串
			fieldRole = fmt.Sprintf("%d", *f.FieldRole)
		}

		formViewFields = append(formViewFields, map[string]interface{}{
			"form_view_field_id":           f.Id,
			"form_view_field_technical_name": f.FieldTechName,
			"form_view_field_business_name":  f.FieldBusinessName,
			"form_view_field_type":          f.FieldType,
			"form_view_field_role":          fieldRole,
			"form_view_field_desc":          f.Comment,
		})
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"message_id":   uuid.New().String(),
		"request_type": "full_understanding",
		"form_view": map[string]interface{}{
			"form_view_id":               formViewId,
			"form_view_technical_name":   formViewData.TableTechName,
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

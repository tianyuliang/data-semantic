// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/config"
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

	// 4. 查询字段数据用于 Kafka 消息
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fields, err := formViewFieldModel.FindByFormViewId(l.ctx, req.Id)
	if err != nil {
		// 记录错误但不中断流程，状态已更新
		logx.WithContext(l.ctx).Errorf("查询字段数据失败: %v", err)
	}

	// 5. 生成 Kafka 消息
	kafkaMessage := l.buildKafkaMessage(req.Id, formViewData.TableTechName, fields)

	// 6. 发送 Kafka 消息
	if l.svcCtx.Kafka != nil {
		go func() {
			// 异步发送，避免阻塞主流程
			if err := l.svcCtx.SendKafkaMessage(config.RequestsTopic, kafkaMessage); err != nil {
				logx.WithContext(l.ctx).Errorf("发送 Kafka 消息失败: %v", err)
			}
		}()
	} else {
		logx.WithContext(l.ctx).Infof("Kafka Producer 未初始化，消息未发送")
	}

	// 7. 返回新状态
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

// buildKafkaMessage 构建 Kafka 消息
func (l *GenerateUnderstandingLogic) buildKafkaMessage(formViewId, tableTechName string, fields []*form_view_field.FormViewFieldBase) map[string]interface{} {
	// 构建字段列表
	fieldList := make([]map[string]interface{}, 0, len(fields))
	for i, f := range fields {
		fieldList = append(fieldList, map[string]interface{}{
			"form_view_field_id": f.Id,
			"field_tech_name":    f.FieldTechName,
			"field_type":         f.FieldType,
			"field_index":        i + 1,
		})
	}

	return map[string]interface{}{
		"message_id":   uuid.New().String(),
		"form_view_id": formViewId,
		"request_type": "full_understanding",
		"request_time": time.Now().Format(time.RFC3339),
		"table_info": map[string]interface{}{
			"table_tech_name": tableTechName,
		},
		"fields": fieldList,
	}
}

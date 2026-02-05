// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/api/internal/types"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view_field"

	"github.com/google/uuid"
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

	// 3. 查询字段数量（用于返回统计信息）
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fieldList, err := formViewFieldModel.FindByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询字段列表失败: %w", err)
	}

	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlConn(l.svcCtx.DB)

	// 4. 查询当前版本号
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询当前版本号失败: %w", err)
	}

	// 5. 生成 Kafka 消息并发送
	messageId := uuid.New().String()
	message := map[string]interface{}{
		"message_id":    messageId,
		"form_view_id":  req.Id,
		"type":          "regenerate_business_objects",
		"version":       latestVersion + 1,
		"request_time":  time.Now().Format(time.RFC3339),
	}

	err = l.svcCtx.SendKafkaMessage("data-understanding-requests", message)
	if err != nil {
		return nil, fmt.Errorf("发送 Kafka 消息失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("Sent regenerate business objects message: messageId=%s, formViewId=%s, version=%d",
		messageId, req.Id, latestVersion+1)

	resp = &types.RegenerateBusinessObjectsResp{
		ObjectCount:    0, // 实际数量由 AI 识别完成后写入
		AttributeCount: len(fieldList),
	}

	return resp, nil
}

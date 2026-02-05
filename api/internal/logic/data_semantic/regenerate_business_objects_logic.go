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

	// 3. 查询字段数据
	formViewFieldModel := form_view_field.NewFormViewFieldModel(l.svcCtx.DB)
	fields, err := formViewFieldModel.FindByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询字段列表失败: %w", err)
	}

	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSqlConn(l.svcCtx.DB)

	// 4. 查询当前版本号
	latestVersion, err := businessObjectTempModel.FindLatestVersionByFormViewId(l.ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("查询当前版本号失败: %w", err)
	}

	// 5. 构建完整的 Kafka 消息（包含字段信息）
	kafkaMessage := l.buildKafkaMessage(req.Id, formViewData.TableTechName, fields, latestVersion+1)

	// 6. 发送 Kafka 消息
	if l.svcCtx.Kafka != nil {
		go func() {
			// 异步发送，避免阻塞主流程
			if err := l.svcCtx.SendKafkaMessage(config.RequestsTopic, kafkaMessage); err != nil {
				logx.WithContext(l.ctx).Errorf("发送 Kafka 消息失败: %v", err)
			} else {
				logx.WithContext(l.ctx).Infof("Sent regenerate business objects message: messageId=%s, formViewId=%s, version=%d",
					kafkaMessage["message_id"], req.Id, latestVersion+1)
			}
		}()
	} else {
		logx.WithContext(l.ctx).Infof("Kafka Producer 未初始化，消息未发送")
	}

	resp = &types.RegenerateBusinessObjectsResp{
		ObjectCount:    0, // 实际数量由 AI 识别完成后写入
		AttributeCount: len(fields),
	}

	return resp, nil
}

// buildKafkaMessage 构建 Kafka 消息
func (l *RegenerateBusinessObjectsLogic) buildKafkaMessage(formViewId, tableTechName string, fields []*form_view_field.FormViewFieldBase, version int) map[string]interface{} {
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
		"request_type": "regenerate_business_objects",
		"version":      version,
		"request_time":  time.Now().Format(time.RFC3339),
		"table_info": map[string]interface{}{
			"table_tech_name": tableTechName,
		},
		"fields": fieldList,
	}
}

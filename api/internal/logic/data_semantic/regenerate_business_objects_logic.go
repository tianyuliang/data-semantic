// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package data_semantic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

	// 5. 更新状态为 1（理解中）
	err = formViewModel.UpdateUnderstandStatus(l.ctx, req.Id, form_view.StatusUnderstanding)
	if err != nil {
		return nil, fmt.Errorf("更新理解状态失败: %w", err)
	}

	// 6. 调用 AI 服务 HTTP API
	go func() {
		// 异步调用，避免阻塞主流程
		if err := l.callAIService(req.Id, formViewData, fields); err != nil {
			logx.WithContext(l.ctx).Errorf("调用 AI 服务失败: %v", err)
			// 状态保持为 1-理解中，由 Kafka 消费者处理失败后更新为 5-理解失败
		}
	}()

	resp = &types.RegenerateBusinessObjectsResp{
		ObjectCount:    0, // 实际数量由 AI 识别完成后写入
		AttributeCount: len(fields),
	}

	return resp, nil
}

// callAIService 调用 AI 服务 HTTP API
func (l *RegenerateBusinessObjectsLogic) callAIService(formViewId string, formViewData *form_view.FormViewDataBase, fields []*form_view_field.FormViewFieldBase) error {
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
		"request_type": "regenerate_business_objects",
		"form_view": map[string]interface{}{
			"form_view_id":               formViewId,
			"form_view_technical_name":   formViewData.TableTechName,
			"form_view_business_name":    formViewData.BusinessName,
			"form_view_desc":             formViewData.Description,
			"form_view_fields":           formViewFields,
		},
	}

	// 调用 AI 服务
	aiResponse, err := l.svcCtx.CallAIService("regenerate_business_objects", requestBody)
	if err != nil {
		return fmt.Errorf("调用 AI 服务失败: %w", err)
	}

	logx.WithContext(l.ctx).Infof("AI 服务调用成功: task_id=%s, status=%s, message=%s",
		aiResponse.TaskID, aiResponse.Status, aiResponse.Message)

	return nil
}

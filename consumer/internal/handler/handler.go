// Package handler Kafka消息处理器
package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/consumer/internal/svc"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_attributes_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/business_object_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_field_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/form_view_info_temp"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/data_understanding/kafka_message_log"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/model/form_view"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// DataUnderstandingHandler 数据理解消息处理器
type DataUnderstandingHandler struct {
	svcCtx *svc.ServiceContext
}

// NewDataUnderstandingHandler 创建消息处理器
func NewDataUnderstandingHandler(svcCtx *svc.ServiceContext) *DataUnderstandingHandler {
	return &DataUnderstandingHandler{svcCtx: svcCtx}
}

// AIResponse AI识别结果响应结构
type AIResponse struct {
	MessageId    string `json:"message_id"`
	FormViewId   string `json:"form_view_id"`
	Version      int    `json:"version"`
	RequestTime  string `json:"request_time"`
	ResponseType string `json:"response_type,omitempty"` // 消息类型: full_understanding, regenerate_business_objects
	Status       string `json:"status,omitempty"`        // 消息状态: success, failed
	// 表信息（全量生成时有值）
	TableInfo       *TableInfo  `json:"table_info,omitempty"`
	TableSemantic   *TableInfo  `json:"table_semantic,omitempty"` // 兼容字段
	// 字段列表（全量生成时有值）
	Fields          []FieldInfo `json:"fields,omitempty"`
	FieldsSemantic  []FieldInfo `json:"fields_semantic,omitempty"` // 兼容字段
	// 业务对象列表（全量生成和部分生成都有值）
	BusinessObjects []BusinessObjectInfo `json:"business_objects,omitempty"`
	// 错误信息（失败时有值）
	Error *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TableInfo 表信息
type TableInfo struct {
	TableBusinessName *string `json:"table_business_name,omitempty"`
	TableDescription  *string `json:"table_description,omitempty"`
}

// FieldInfo 字段信息
type FieldInfo struct {
	FormViewFieldId   string  `json:"form_view_field_id"`
	FieldTechName     string  `json:"field_tech_name"`
	FieldBusinessName *string `json:"field_business_name,omitempty"`
	FieldRole         *int8   `json:"field_role,omitempty"`
	FieldDescription  *string `json:"field_description,omitempty"`
}

// BusinessObjectInfo 业务对象信息
type BusinessObjectInfo struct {
	Id         string         `json:"id"`
	ObjectName string         `json:"object_name"`
	Attributes []AttributeInfo `json:"attributes,omitempty"`
}

// AttributeInfo 属性信息
type AttributeInfo struct {
	Id              string `json:"id"`
	AttrName        string `json:"attr_name"`
	FormViewFieldId string `json:"form_view_field_id"`
}

// Handle 处理Kafka消息
func (h *DataUnderstandingHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	logx.Infof("收到Kafka消息: topic=%s partition=%d offset=%d",
		message.Topic, message.Partition, message.Offset)

	// 解析消息为固定结构
	var aiResp AIResponse
	if err := json.Unmarshal(message.Value, &aiResp); err != nil {
		logx.Errorf("解析消息失败: %v", err)
		// 无法解析时记录失败（message_id 为空）
		_ = h.recordFailure(ctx, "", "", fmt.Sprintf("解析消息失败: %v", err))
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 验证必填字段
	if aiResp.MessageId == "" {
		err := fmt.Errorf("消息缺少 message_id 字段")
		_ = h.recordFailure(ctx, "", aiResp.FormViewId, err.Error())
		return err
	}

	// 判断消息类型：失败消息或成功消息
	isFailedMessage := aiResp.Status == "failed"

	// 在事务中处理：去重检查 + 数据保存 + 状态更新
	err := h.svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 去重检查（只检查成功处理的消息，允许失败消息重试）
		kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSession(session)
		if exists, err := kafkaMessageLogModel.ExistsSuccessByMessageId(ctx, aiResp.MessageId); err != nil {
			return fmt.Errorf("检查消息去重失败: %w", err)
		} else if exists {
			logx.Infof("消息已成功处理，跳过: message_id=%s", aiResp.MessageId)
			return nil
		}

		// 2. 根据消息状态处理
		if isFailedMessage {
			// 处理失败消息：记录失败日志 + 更新状态为 5（理解失败）
			return h.processFailedResponseInTx(ctx, session, &aiResp)
		}

		// 3. 处理成功响应（包括数据保存和状态更新）
		if err := h.processSuccessResponseInTx(ctx, session, &aiResp); err != nil {
			return err
		}

		// 4. 记录消息处理日志
		if _, err := kafkaMessageLogModel.InsertSuccess(ctx, aiResp.MessageId, aiResp.FormViewId); err != nil {
			return fmt.Errorf("记录Kafka消息日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		// 记录失败日志
		_ = h.recordFailure(ctx, aiResp.MessageId, aiResp.FormViewId, err.Error())
		return err
	}

	logx.WithContext(ctx).Infof("处理成功响应: message_id=%s, form_view_id=%s, type=%s",
		aiResp.MessageId, aiResp.FormViewId, aiResp.ResponseType)

	return nil
}

// processSuccessResponseInTx 处理成功响应（在事务中执行，包含状态更新）
func (h *DataUnderstandingHandler) processSuccessResponseInTx(ctx context.Context, session sqlx.Session, aiResp *AIResponse) error {
	// 判断消息类型：优先使用 response_type 字段判断
	isFullUnderstanding := h.isFullUnderstanding(aiResp)

	var newVersion int
	var err error

	if isFullUnderstanding {
		// 全量生成：使用行锁查询表信息临时表的版本号
		newVersion, err = h.getNextVersionWithLock(ctx, session, aiResp.FormViewId, "table_info")
		if err != nil {
			return fmt.Errorf("获取表信息版本号失败: %w", err)
		}

		logx.WithContext(ctx).Infof("全量生成: form_view_id=%s, 新版本=%d", aiResp.FormViewId, newVersion)

		// 1.1 保存表信息到临时表（如果有）
		if aiResp.TableInfo != nil {
			if err := h.saveTableInfo(ctx, session, aiResp.FormViewId, newVersion, aiResp.TableInfo); err != nil {
				return fmt.Errorf("保存表信息失败: %w", err)
			}
		}

		// 1.2 保存字段信息到临时表（如果有）
		if len(aiResp.Fields) > 0 {
			if err := h.saveFieldInfo(ctx, session, aiResp.FormViewId, newVersion, aiResp.Fields); err != nil {
				return fmt.Errorf("保存字段信息失败: %w", err)
			}
		}
	} else {
		// 部分生成（重新识别业务对象）：使用行锁查询业务对象临时表的版本号
		newVersion, err = h.getNextVersionWithLock(ctx, session, aiResp.FormViewId, "business_object")
		if err != nil {
			return fmt.Errorf("获取业务对象版本号失败: %w", err)
		}

		logx.WithContext(ctx).Infof("重新识别业务对象: form_view_id=%s, 新版本=%d", aiResp.FormViewId, newVersion)
	}

	// 1.4 保存业务对象到临时表
	if len(aiResp.BusinessObjects) > 0 {
		if err := h.saveBusinessObjects(ctx, session, aiResp.FormViewId, newVersion, aiResp.BusinessObjects); err != nil {
			return fmt.Errorf("保存业务对象失败: %w", err)
		}
	}

	// 1.5 更新 form_view 状态为 2（待确认）- 在事务中执行
	formViewModel := form_view.NewFormViewModelSession(session)
	if err := formViewModel.UpdateUnderstandStatus(ctx, aiResp.FormViewId, form_view.StatusPendingConfirm); err != nil {
		return fmt.Errorf("更新form_view状态失败: %w", err)
	}

	return nil
}

// processFailedResponseInTx 处理失败响应（在事务中执行，包含状态更新）
func (h *DataUnderstandingHandler) processFailedResponseInTx(ctx context.Context, session sqlx.Session, aiResp *AIResponse) error {
	// 构建错误信息
	errMsg := "AI 服务处理失败"
	if aiResp.Error != nil {
		errMsg = fmt.Sprintf("%s: %s", aiResp.Error.Code, aiResp.Error.Message)
	}

	logx.WithContext(ctx).Errorf("处理失败消息: message_id=%s, form_view_id=%s, error=%s",
		aiResp.MessageId, aiResp.FormViewId, errMsg)

	// 记录失败日志到 kafka_message_log
	kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSession(session)
	if _, err := kafkaMessageLogModel.InsertFailure(ctx, aiResp.MessageId, aiResp.FormViewId, errMsg); err != nil {
		return fmt.Errorf("记录Kafka消息失败日志失败: %w", err)
	}

	// 更新 form_view 状态为 5（理解失败）
	formViewModel := form_view.NewFormViewModelSession(session)
	if err := formViewModel.UpdateUnderstandStatus(ctx, aiResp.FormViewId, form_view.StatusFailed); err != nil {
		return fmt.Errorf("更新form_view状态为理解失败失败: %w", err)
	}

	return nil
}

// isFullUnderstanding 判断是否为全量生成
func (h *DataUnderstandingHandler) isFullUnderstanding(aiResp *AIResponse) bool {
	// 优先使用 response_type 字段判断
	switch aiResp.ResponseType {
	case "regenerate_business_objects":
		return false
	case "full_understanding":
		return true
	}

	// 兼容旧逻辑：如果有表信息或字段信息，则认为是全量生成
	return aiResp.TableInfo != nil || len(aiResp.Fields) > 0
}

// getNextVersionWithLock 获取下一个版本号（使用行锁防止并发冲突）
func (h *DataUnderstandingHandler) getNextVersionWithLock(ctx context.Context, session sqlx.Session, formViewId, versionType string) (int, error) {
	var latestVersion int
	var err error

	if versionType == "table_info" {
		formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)
		latestVersion, err = formViewInfoTempModel.FindLatestVersionWithLock(ctx, formViewId)
	} else {
		businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
		latestVersion, err = businessObjectTempModel.FindLatestVersionWithLock(ctx, formViewId)
	}

	if err != nil {
		return 0, err
	}

	return latestVersion + 1, nil
}

// saveTableInfo 保存表信息到临时表（直接插入）
func (h *DataUnderstandingHandler) saveTableInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, tableInfo *TableInfo) error {
	formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)

	// 直接插入新版本数据
	data := &form_view_info_temp.FormViewInfoTemp{
		Id:                uuid.New().String(),
		FormViewId:        formViewId,
		Version:           version,
		TableBusinessName: tableInfo.TableBusinessName,
		TableDescription:  tableInfo.TableDescription,
	}
	if _, err := formViewInfoTempModel.Insert(ctx, data); err != nil {
		return fmt.Errorf("插入表信息失败: %w", err)
	}

	return nil
}

// saveFieldInfo 保存字段信息到临时表（直接插入）
func (h *DataUnderstandingHandler) saveFieldInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, fields []FieldInfo) error {
	formViewFieldInfoTempModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSession(session)

	for _, field := range fields {
		// 直接插入新版本数据
		data := &form_view_field_info_temp.FormViewFieldInfoTemp{
			Id:                uuid.New().String(),
			FormViewId:        formViewId,
			FormViewFieldId:   field.FormViewFieldId,
			Version:           version,
			FieldBusinessName: field.FieldBusinessName,
			FieldRole:         field.FieldRole,
			FieldDescription:  field.FieldDescription,
		}
		if _, err := formViewFieldInfoTempModel.Insert(ctx, data); err != nil {
			return fmt.Errorf("插入字段信息失败: %w", err)
		}
	}

	return nil
}

// saveBusinessObjects 保存业务对象到临时表（直接插入，使用 UUID 作为 ID）
func (h *DataUnderstandingHandler) saveBusinessObjects(ctx context.Context, session sqlx.Session, formViewId string, version int, objects []BusinessObjectInfo) error {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)

	for _, obj := range objects {
		// 1. 生成新的业务对象 ID
		businessObjectId := uuid.New().String()

		// 2. 插入新版本业务对象
		objectData := &business_object_temp.BusinessObjectTemp{
			Id:         businessObjectId,
			FormViewId: formViewId,
			Version:    version,
			ObjectName: obj.ObjectName,
		}
		if _, err := businessObjectTempModel.Insert(ctx, objectData); err != nil {
			return fmt.Errorf("插入业务对象失败: %w", err)
		}

		// 3. 处理属性
		for _, attr := range obj.Attributes {
			// 生成新的属性 ID
			attrId := uuid.New().String()

			// 插入新版本属性
			attrData := &business_object_attributes_temp.BusinessObjectAttributesTemp{
				Id:               attrId,
				FormViewId:       formViewId,
				BusinessObjectId: businessObjectId,
				Version:          version,
				FormViewFieldId:  attr.FormViewFieldId,
				AttrName:         attr.AttrName,
			}
			if _, err := businessObjectAttrTempModel.Insert(ctx, attrData); err != nil {
				return fmt.Errorf("插入属性失败: %w", err)
			}
		}
	}

	return nil
}

// recordFailure 记录处理失败
func (h *DataUnderstandingHandler) recordFailure(ctx context.Context, messageId, formViewId, errMsg string) error {
	// 记录结构化日志
	logx.Errorf("Kafka消息处理失败: message_id=%s, form_view_id=%s, error=%s",
		messageId, formViewId, errMsg)

	// 记录到 t_kafka_message_log
	kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSqlx(h.svcCtx.DB)
	if _, err := kafkaMessageLogModel.InsertFailure(ctx, messageId, formViewId, errMsg); err != nil {
		logx.Errorf("记录Kafka消息失败日志失败: %v", err)
	}

	return nil
}

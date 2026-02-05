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
	MessageId   string `json:"message_id"`
	FormViewId  string `json:"form_view_id"`
	Version     int    `json:"version"`
	RequestTime string `json:"request_time"`
	// 表信息
	TableInfo *TableInfo `json:"table_info,omitempty"`
	// 字段列表
	Fields []FieldInfo `json:"fields,omitempty"`
	// 业务对象列表
	BusinessObjects []BusinessObjectInfo `json:"business_objects,omitempty"`
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

	// 去重检查
	if exists, err := h.checkMessageId(ctx, aiResp.MessageId); err != nil {
		logx.Errorf("检查消息去重失败: %v", err)
		return err
	} else if exists {
		logx.Infof("消息已处理，跳过: message_id=%s", aiResp.MessageId)
		return nil
	}

	// 处理成功响应
	if err := h.processSuccessResponse(ctx, &aiResp); err != nil {
		// 记录失败日志
		_ = h.recordFailure(ctx, aiResp.MessageId, aiResp.FormViewId, err.Error())
		return err
	}

	return nil
}

// checkMessageId 检查消息是否已处理
func (h *DataUnderstandingHandler) checkMessageId(ctx context.Context, messageId string) (bool, error) {
	kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSqlConn(h.svcCtx.DB)
	return kafkaMessageLogModel.ExistsByMessageId(ctx, messageId)
}

// processSuccessResponse 处理成功响应
func (h *DataUnderstandingHandler) processSuccessResponse(ctx context.Context, aiResp *AIResponse) error {
	// 1. 开启事务处理
	err := h.svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1.1 获取当前最新版本号并递增
		formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)
		latestVersion := 0
		latestRecord, err := formViewInfoTempModel.FindLatestByFormViewId(ctx, aiResp.FormViewId)
		// 如果找到记录，使用其版本号；否则从 0 开始
		if err == nil && latestRecord != nil {
			latestVersion = latestRecord.Version
		}
		// 版本号递增
		newVersion := latestVersion + 1

		// 1.2 记录消息处理日志
		kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSession(session)
		if _, err := kafkaMessageLogModel.InsertSuccess(ctx, aiResp.MessageId, aiResp.FormViewId); err != nil {
			return fmt.Errorf("记录Kafka消息日志失败: %w", err)
		}

		// 1.3 保存表信息到临时表
		if aiResp.TableInfo != nil {
			if err := h.saveTableInfo(ctx, session, aiResp.FormViewId, newVersion, aiResp.TableInfo); err != nil {
				return fmt.Errorf("保存表信息失败: %w", err)
			}
		}

		// 1.4 保存字段信息到临时表
		if len(aiResp.Fields) > 0 {
			if err := h.saveFieldInfo(ctx, session, aiResp.FormViewId, newVersion, aiResp.Fields); err != nil {
				return fmt.Errorf("保存字段信息失败: %w", err)
			}
		}

		// 1.5 保存业务对象到临时表
		if len(aiResp.BusinessObjects) > 0 {
			if err := h.saveBusinessObjects(ctx, session, aiResp.FormViewId, newVersion, aiResp.BusinessObjects); err != nil {
				return fmt.Errorf("保存业务对象失败: %w", err)
			}
		}

		logx.WithContext(ctx).Infof("版本递增: form_view_id=%s, 旧版本=%d, 新版本=%d",
			aiResp.FormViewId, latestVersion, newVersion)

		return nil
	})

	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}

	// 2. 更新 form_view 状态为 2（待确认）
	formViewModel := form_view.NewFormViewModel(h.svcCtx.DB)
	if err := formViewModel.UpdateUnderstandStatus(ctx, aiResp.FormViewId, form_view.StatusPendingConfirm); err != nil {
		return fmt.Errorf("更新form_view状态失败: %w", err)
	}

	logx.WithContext(ctx).Infof("处理成功响应: message_id=%s, form_view_id=%s",
		aiResp.MessageId, aiResp.FormViewId)

	return nil
}

// saveTableInfo 保存表信息到临时表（逻辑删除旧版本，插入新版本）
func (h *DataUnderstandingHandler) saveTableInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, tableInfo *TableInfo) error {
	formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)

	// 1. 逻辑删除旧版本数据
	if err := formViewInfoTempModel.DeleteByFormViewId(ctx, formViewId); err != nil {
		return fmt.Errorf("逻辑删除旧版本表信息失败: %w", err)
	}

	// 2. 插入新版本数据
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

// saveFieldInfo 保存字段信息到临时表（逻辑删除旧版本，插入新版本）
func (h *DataUnderstandingHandler) saveFieldInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, fields []FieldInfo) error {
	formViewFieldInfoTempModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSession(session)

	for _, field := range fields {
		// 1. 逻辑删除该字段的旧版本数据
		if err := formViewFieldInfoTempModel.DeleteByFormFieldId(ctx, field.FormViewFieldId); err != nil {
			return fmt.Errorf("逻辑删除旧版本字段信息失败: %w", err)
		}

		// 2. 插入新版本数据
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

// saveBusinessObjects 保存业务对象到临时表（逻辑删除旧版本，插入新版本）
func (h *DataUnderstandingHandler) saveBusinessObjects(ctx context.Context, session sqlx.Session, formViewId string, version int, objects []BusinessObjectInfo) error {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)

	for _, obj := range objects {
		// 1. 逻辑删除该业务对象的旧版本数据
		if err := businessObjectTempModel.DeleteById(ctx, obj.Id); err != nil {
			return fmt.Errorf("逻辑删除旧版本业务对象失败: %w", err)
		}

		// 2. 插入新版本业务对象
		objectData := &business_object_temp.BusinessObjectTemp{
			Id:         obj.Id,
			FormViewId: formViewId,
			Version:    version,
			ObjectName: obj.ObjectName,
		}
		if _, err := businessObjectTempModel.Insert(ctx, objectData); err != nil {
			return fmt.Errorf("插入业务对象失败: %w", err)
		}

		// 3. 处理属性
		for _, attr := range obj.Attributes {
			// 逻辑删除该属性的旧版本数据
			if err := businessObjectAttrTempModel.DeleteById(ctx, attr.Id); err != nil {
				return fmt.Errorf("逻辑删除旧版本属性失败: %w", err)
			}

			// 插入新版本属性
			attrData := &business_object_attributes_temp.BusinessObjectAttributesTemp{
				Id:               attr.Id,
				FormViewId:       formViewId,
				BusinessObjectId: obj.Id,
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
	kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSqlConn(h.svcCtx.DB)
	if _, err := kafkaMessageLogModel.InsertFailure(ctx, messageId, formViewId, errMsg); err != nil {
		logx.Errorf("记录Kafka消息失败日志失败: %v", err)
	}

	return nil
}

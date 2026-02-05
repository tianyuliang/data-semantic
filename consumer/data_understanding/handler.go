// Package data_understanding Kafka消息处理逻辑
package data_understanding

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

	// 解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		logx.Errorf("解析消息失败: %v", err)
		messageId, _ := msg["message_id"].(string)
		formViewId, _ := msg["form_view_id"].(string)
		return h.recordFailure(ctx, messageId, formViewId, fmt.Sprintf("解析消息失败: %v", err))
	}

	// 提取字段
	messageId, _ := msg["message_id"].(string)
	formViewId, _ := msg["form_view_id"].(string)

	// 去重检查
	if exists, err := h.checkMessageId(ctx, messageId); err != nil {
		logx.Errorf("检查消息去重失败: %v", err)
		return err
	} else if exists {
		logx.Infof("消息已处理，跳过: message_id=%s", messageId)
		return nil
	}

	// 处理成功响应
	if err := h.processSuccessResponse(ctx, messageId, formViewId, msg); err != nil {
		// 记录失败日志
		_ = h.recordFailure(ctx, messageId, formViewId, err.Error())
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
func (h *DataUnderstandingHandler) processSuccessResponse(ctx context.Context, messageId, formViewId string, msg map[string]interface{}) error {
	// 1. 解析AI响应
	var aiResp AIResponse
	respBytes, _ := json.Marshal(msg)
	if err := json.Unmarshal(respBytes, &aiResp); err != nil {
		return fmt.Errorf("解析AI响应失败: %w", err)
	}

	// 2. 开启事务处理
	err := h.svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 2.1 记录消息处理日志
		kafkaMessageLogModel := kafka_message_log.NewKafkaMessageLogModelSession(session)
		if _, err := kafkaMessageLogModel.InsertSuccess(ctx, messageId, formViewId); err != nil {
			return fmt.Errorf("记录Kafka消息日志失败: %w", err)
		}

		// 2.2 保存表信息到临时表
		if aiResp.TableInfo != nil {
			if err := h.saveTableInfo(ctx, session, formViewId, aiResp.Version, aiResp.TableInfo); err != nil {
				return fmt.Errorf("保存表信息失败: %w", err)
			}
		}

		// 2.3 保存字段信息到临时表
		if len(aiResp.Fields) > 0 {
			if err := h.saveFieldInfo(ctx, session, formViewId, aiResp.Version, aiResp.Fields); err != nil {
				return fmt.Errorf("保存字段信息失败: %w", err)
			}
		}

		// 2.4 保存业务对象到临时表
		if len(aiResp.BusinessObjects) > 0 {
			if err := h.saveBusinessObjects(ctx, session, formViewId, aiResp.Version, aiResp.BusinessObjects); err != nil {
				return fmt.Errorf("保存业务对象失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}

	// 3. 更新 form_view 状态为 2（待确认）
	formViewModel := form_view.NewFormViewModel(h.svcCtx.DB)
	if err := formViewModel.UpdateUnderstandStatus(ctx, formViewId, form_view.StatusPendingConfirm); err != nil {
		return fmt.Errorf("更新form_view状态失败: %w", err)
	}

	logx.WithContext(ctx).Infof("处理成功响应: message_id=%s, form_view_id=%s, version=%d",
		messageId, formViewId, aiResp.Version)

	return nil
}

// saveTableInfo 保存表信息到临时表
func (h *DataUnderstandingHandler) saveTableInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, tableInfo *TableInfo) error {
	formViewInfoTempModel := form_view_info_temp.NewFormViewInfoTempModelSession(session)

	// 查询是否已有记录
	existing, err := formViewInfoTempModel.FindLatestByFormViewId(ctx, formViewId)
	if err != nil {
		// 记录不存在，创建新记录
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

	// 更新现有记录
	existing.TableBusinessName = tableInfo.TableBusinessName
	existing.TableDescription = tableInfo.TableDescription
	if err := formViewInfoTempModel.Update(ctx, existing); err != nil {
		return fmt.Errorf("更新表信息失败: %w", err)
	}

	return nil
}

// saveFieldInfo 保存字段信息到临时表
func (h *DataUnderstandingHandler) saveFieldInfo(ctx context.Context, session sqlx.Session, formViewId string, version int, fields []FieldInfo) error {
	formViewFieldInfoTempModel := form_view_field_info_temp.NewFormViewFieldInfoTempModelSession(session)

	for _, field := range fields {
		// 查询是否已有记录
		existing, err := formViewFieldInfoTempModel.FindOneByFormFieldId(ctx, field.FormViewFieldId)
		if err != nil {
			// 记录不存在，创建新记录
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
			continue
		}

		// 更新现有记录
		existing.FormViewId = formViewId
		existing.Version = version
		existing.FieldBusinessName = field.FieldBusinessName
		existing.FieldRole = field.FieldRole
		existing.FieldDescription = field.FieldDescription
		if err := formViewFieldInfoTempModel.Update(ctx, existing); err != nil {
			return fmt.Errorf("更新字段信息失败: %w", err)
		}
	}

	return nil
}

// saveBusinessObjects 保存业务对象到临时表
func (h *DataUnderstandingHandler) saveBusinessObjects(ctx context.Context, session sqlx.Session, formViewId string, version int, objects []BusinessObjectInfo) error {
	businessObjectTempModel := business_object_temp.NewBusinessObjectTempModelSession(session)
	businessObjectAttrTempModel := business_object_attributes_temp.NewBusinessObjectAttributesTempModelSession(session)

	for _, obj := range objects {
		// 插入或更新业务对象
		objectData := &business_object_temp.BusinessObjectTemp{
			Id:         obj.Id,
			FormViewId: formViewId,
			Version:    version,
			ObjectName: obj.ObjectName,
		}

		// 尝试查询现有记录
		existing, err := businessObjectTempModel.FindOneById(ctx, obj.Id)
		if err != nil {
			// 记录不存在，插入新记录
			if _, err := businessObjectTempModel.Insert(ctx, objectData); err != nil {
				return fmt.Errorf("插入业务对象失败: %w", err)
			}
		} else {
			// 更新现有记录
			existing.FormViewId = formViewId
			existing.Version = version
			existing.ObjectName = obj.ObjectName
			if err := businessObjectTempModel.Update(ctx, existing); err != nil {
				return fmt.Errorf("更新业务对象失败: %w", err)
			}
		}

		// 保存属性
		for _, attr := range obj.Attributes {
			attrData := &business_object_attributes_temp.BusinessObjectAttributesTemp{
				Id:               attr.Id,
				FormViewId:       formViewId,
				BusinessObjectId: obj.Id,
				Version:          version,
				FormViewFieldId:  attr.FormViewFieldId,
				AttrName:         attr.AttrName,
			}

			// 尝试查询现有记录
			existingAttr, err := businessObjectAttrTempModel.FindOneById(ctx, attr.Id)
			if err != nil {
				// 记录不存在，插入新记录
				if _, err := businessObjectAttrTempModel.Insert(ctx, attrData); err != nil {
					return fmt.Errorf("插入属性失败: %w", err)
				}
			} else {
				// 更新现有记录
				existingAttr.FormViewId = formViewId
				existingAttr.BusinessObjectId = obj.Id
				existingAttr.Version = version
				existingAttr.FormViewFieldId = attr.FormViewFieldId
				existingAttr.AttrName = attr.AttrName
				if err := businessObjectAttrTempModel.Update(ctx, existingAttr); err != nil {
					return fmt.Errorf("更新属性失败: %w", err)
				}
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

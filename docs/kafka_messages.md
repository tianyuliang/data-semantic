# Kafka 消息示例文档

本文档提供 AI 服务返回的 Kafka 消息格式示例，Consumer 会根据这些消息处理数据。

## 消息处理逻辑

Consumer 根据 `response_type` 字段判断消息类型：

```go
isFullUnderstanding := aiResp.ResponseType == "full_understanding" ||
    aiResp.ResponseType == "" ||
    (aiResp.TableInfo != nil || len(aiResp.Fields) > 0)
```

- **全量理解** (`full_understanding`): 包含表信息、字段信息和业务对象
- **部分理解** (`partial_understanding`): 仅包含业务对象
- **重新识别** (`regenerate_business_objects`): 仅包含业务对象

## 1. 全量理解 Kafka 消息

```json
{
  "message_id": "018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "form_view_id": "018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "version": 1,
  "request_time": "2026-02-10T15:00:00Z",
  "response_type": "full_understanding",
  "table_info": {
    "table_business_name": "用户表",
    "table_description": "存储用户基本信息"
  },
  "fields": [
    {
      "form_view_field_id": "field-001",
      "field_tech_name": "id",
      "field_business_name": "用户ID",
      "field_role": 1,
      "field_description": "用户唯一标识"
    },
    {
      "form_view_field_id": "field-002",
      "field_tech_name": "username",
      "field_business_name": "用户名",
      "field_role": 6,
      "field_description": "用户登录名称"
    },
    {
      "form_view_field_id": "field-003",
      "field_tech_name": "email",
      "field_business_name": "邮箱",
      "field_role": 2,
      "field_description": "用户联系邮箱"
    },
    {
      "form_view_field_id": "field-004",
      "field_tech_name": "created_at",
      "field_business_name": "创建时间",
      "field_role": 4,
      "field_description": "记录创建时间"
    }
  ],
  "business_objects": [
    {
      "id": "bo-001",
      "object_name": "用户",
      "attributes": [
        {
          "id": "attr-001",
          "attr_name": "用户ID",
          "form_view_field_id": "field-001"
        },
        {
          "id": "attr-002",
          "attr_name": "用户名",
          "form_view_field_id": "field-002"
        },
        {
          "id": "attr-003",
          "attr_name": "邮箱",
          "form_view_field_id": "field-003"
        }
      ]
    },
    {
      "id": "bo-002",
      "object_name": "用户联系方式",
      "attributes": [
        {
          "id": "attr-004",
          "attr_name": "邮箱",
          "form_view_field_id": "field-003"
        }
      ]
    }
  ]
}
```

**处理流程**:
1. 查询 `t_form_view_info_temp` 表获取当前版本号
2. 保存 `table_info` 到 `t_form_view_info_temp` 表
3. 保存 `fields` 到 `t_form_view_field_info_temp` 表
4. 保存 `business_objects` 到 `t_business_object_temp` 和 `t_business_object_attributes_temp` 表
5. 更新 `form_view` 状态为 2 (待确认)

## 2. 部分理解 Kafka 消息

```json
{
  "message_id": "msg-partial-018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "form_view_id": "018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "version": 1,
  "request_time": "2026-02-10T15:00:00Z",
  "response_type": "partial_understanding",
  "business_objects": [
    {
      "id": "bo-003",
      "object_name": "用户信息",
      "attributes": [
        {
          "id": "attr-005",
          "attr_name": "用户ID",
          "form_view_field_id": "field-001"
        },
        {
          "id": "attr-006",
          "attr_name": "用户名",
          "form_view_field_id": "field-002"
        }
      ]
    }
  ]
}
```

**处理流程**:
1. 查询 `t_business_object_temp` 表获取当前版本号
2. 仅保存 `business_objects` 到临时表
3. 更新 `form_view` 状态为 2 (待确认)

## 3. 重新识别业务对象 Kafka 消息

```json
{
  "message_id": "msg-regen-018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "form_view_id": "018f7d4b-6f8c-7b9a-0c1d-2e3f4a5b6c7d",
  "version": 2,
  "request_time": "2026-02-10T14:00:00Z",
  "response_type": "regenerate_business_objects",
  "business_objects": [
    {
      "id": "bo-004",
      "object_name": "用户基本信息",
      "attributes": [
        {
          "id": "attr-007",
          "attr_name": "用户标识",
          "form_view_field_id": "field-001"
        }
      ]
    },
    {
      "id": "bo-005",
      "object_name": "用户账户信息",
      "attributes": [
        {
          "id": "attr-008",
          "attr_name": "登录名称",
          "form_view_field_id": "field-002"
        },
        {
          "id": "attr-009",
          "attr_name": "邮箱地址",
          "form_view_field_id": "field-003"
        }
      ]
    }
  ]
}
```

**处理流程**: 与部分理解相同

## 字段角色映射 (field_role)

| 值 | 含义 |
|----|----|
| 1  | 业务主键 |
| 2  | 关联标识 |
| 3  | 业务状态 |
| 4  | 时间字段 |
| 5  | 业务指标 |
| 6  | 业务特征 |
| 7  | 审计字段 |
| 8  | 技术字段 |

## 数据库表结构

消息处理涉及的临时表：

- `t_form_view_info_temp` - 表信息临时表
- `t_form_view_field_info_temp` - 字段信息临时表
- `t_business_object_temp` - 业务对象临时表
- `t_business_object_attributes_temp` - 业务对象属性临时表
- `t_kafka_message_log` - Kafka 消息处理日志

## 版本控制

- 全量理解：版本号基于 `t_form_view_info_temp.version`
- 部分/重新识别：版本号基于 `t_business_object_temp.version`
- 每次处理都会创建新版本，旧版本逻辑删除

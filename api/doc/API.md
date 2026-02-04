# Data Understanding API 文档

## 概述

Data Understanding API 提供库表数据语义理解和业务对象识别功能。

**Base URL**: `/api/v1/data-semantic`

**认证方式**: JWT Bearer Token

---

## 接口列表

### 1. 查询库表理解状态
**Endpoint**: `GET /:id/status`

**描述**: 查询指定库表的理解状态和版本信息

**参数**:
- `id` (path): 库表视图ID

**响应**:
```json
{
  "understand_status": 0,
  "current_version": 0
}
```

**状态值说明**:
- `0`: 未理解
- `1`: 理解中
- `2`: 待确认
- `3`: 已完成

---

### 2. 一键生成理解数据
**Endpoint**: `POST /:id/generate`

**描述**: 启动 AI 分析库表和字段的业务语义

**参数**:
- `id` (path): 库表视图ID

**响应**:
```json
{
  "understand_status": 1
}
```

---

### 3. 查询字段语义补全数据
**Endpoint**: `GET /:id/fields`

**描述**: 查询字段语义信息（业务名称、角色、描述）

**参数**:
- `id` (path): 库表视图ID
- `keyword` (query, optional): 按字段名/业务名过滤
- `only_incomplete` (query, optional): 仅返回未完成的字段

**响应**:
```json
{
  "current_version": 10,
  "fields": [
    {
      "id": "field-id",
      "field_tech_name": "user_id",
      "field_business_name": "用户ID",
      "field_role": 1,
      "field_description": "系统唯一标识"
    }
  ]
}
```

**字段角色说明**:
- `0`: 无角色
- `1`: 业务主键
- `2`: 业务属性

---

### 4. 保存语义补全数据
**Endpoint**: `PUT /:id/semantic-info`

**描述**: 保存或编辑库表/字段语义信息

**参数**:
- `id` (path): 库表视图ID
- Body:
```json
{
  "type": "table",
  "id": "field-id",
  "business_name": "用户ID",
  "field_role": 1,
  "description": "系统唯一标识"
}
```

---

### 5. 查询业务对象识别结果
**Endpoint**: `GET /:id/business-objects`

**描述**: 查询 AI 识别的业务对象和属性分组

**参数**:
- `id` (path): 库表视图ID
- `object_id` (query, optional): 按业务对象ID过滤
- `keyword` (query, optional): 按名称过滤

**响应**:
```json
{
  "current_version": 10,
  "list": [
    {
      "id": "object-id",
      "object_name": "用户",
      "attributes": [
        {
          "id": "attr-id",
          "attr_name": "用户ID",
          "form_view_field_id": "field-id"
        }
      ]
    }
  ]
}
```

---

### 6. 保存业务对象及属性
**Endpoint**: `PUT /:id/business-objects`

**描述**: 修改业务对象名称或属性名称

**参数**:
- `id` (path): 库表视图ID
- Body:
```json
{
  "type": "object",
  "id": "object-id",
  "name": "用户信息"
}
```

---

### 7. 调整属性归属业务对象
**Endpoint**: `PUT /:id/business-objects/attributes/move`

**描述**: 将属性移动到其他业务对象

**参数**:
- `id` (path): 库表视图ID
- Body:
```json
{
  "attribute_id": "attr-id",
  "target_object_uuid": "target-object-id"
}
```

---

### 8. 重新识别业务对象
**Endpoint**: `POST /:id/business-objects/regenerate`

**描述**: 基于当前字段语义重新识别业务对象

**参数**:
- `id` (path): 库表视图ID

**响应**:
```json
{
  "object_count": 5,
  "attribute_count": 25
}
```

---

### 9. 提交确认理解数据
**Endpoint**: `POST /:id/submit`

**描述**: 将临时表数据同步到正式表

**参数**:
- `id` (path): 库表视图ID

**响应**:
```json
{
  "success": true
}
```

---

### 10. 删除识别结果
**Endpoint**: `DELETE /:id/business-objects`

**描述**: 删除业务对象临时数据

**参数**:
- `id` (path): 库表视图ID

**响应**:
```json
{
  "success": true
}
```

---

## 错误响应格式

所有错误响应遵循以下格式：

```json
{
  "code": 40001,
  "message": "参数验证失败"
}
```

## 常见错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数验证失败 |
| 40401 | 资源不存在 |
| 50001 | 内部服务错误 |

---

## Swagger 文档

完整的 Swagger 文档请参考: [swagger.json](swagger.json)

可以使用 [Swagger UI](https://petstore.swagger.io/) 在线查看和测试 API。

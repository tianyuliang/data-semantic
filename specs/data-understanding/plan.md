# Data Understanding Technical Plan

> **Branch**: `feature/data-understanding`
> **Spec Path**: `specs/data-understanding/`
> **Created**: 2026-02-03
> **Status**: Draft

---

## Summary

本技术方案实现了库表数据的语义理解和业务对象自动识别功能。核心决策包括：
- **版本控制机制**：通过临时表实现数据的版本管理，支持重新识别和历史追溯
- **异步处理架构**：Kafka 消息队列实现 AI 服务的异步调用
- **状态机设计**：6种理解状态（0-5）的单向流转管理，包含理解失败状态
- **数据源隔离**：正式表与临时表分离，根据状态动态切换数据源

---

## Technical Context

| Item | Value |
|------|-------|
| **Language** | Go 1.24+ |
| **Framework** | Go-Zero v1.9+ |
| **Storage** | MySQL 8.0 |
| **Cache** | Redis 7.0 (限流) |
| **Message Queue** | Kafka 3.0 (AI 集成) |
| **ORM** | SQLx |
| **Testing** | go test |
| **Common Lib** | idrm-go-base v0.1.0+ |

---

## Constitution Check

### 强制约束验证

| 规则 | 状态 | 说明 |
|------|------|------|
| 跳过工作流阶段 | ✅ PASS | 遵循 5 阶段工作流 |
| Handler 写业务逻辑 | ✅ PASS | Handler 仅参数校验，Logic 层处理业务 |
| Logic 直接访问数据库 | ✅ PASS | 通过 Model 层访问 |
| UUID v7 主键 | ✅ PASS | 所有表使用 CHAR(36) UUID v7 |
| datetime(3) 精度 | ✅ PASS | 所有时间字段使用毫秒精度 |
| 软删除 deleted_at | ✅ PASS | 所有表包含 deleted_at 字段 |
| 禁止物理外键 | ✅ PASS | 关联关系在 Logic 层维护 |

### 架构规范验证

| 层级 | 职责 | 文件位置 |
|------|------|----------|
| Handler | 参数校验、响应格式化 | `api/internal/handler/data_semantic/` |
| Logic | 业务逻辑、状态流转、事务管理 | `api/internal/logic/data_semantic/` |
| Model | 数据访问、CRUD 操作 | `model/data_understanding/` |

---

## 通用库 (idrm-go-base)

**安装**:
```bash
go get github.com/jinguoxing/idrm-go-base@latest
```

### 自定义错误码

| 功能 | 范围 | 位置 |
|------|------|------|
| 数据理解 | 600101-600130 | `internal/errorx/codes.go` |

| 错误码范围 | 说明 |
|------------|------|
| 600101-600110 | 状态校验相关错误 |
| 600111-600120 | 数据校验相关错误（重复、不存在等） |
| 600121-600130 | 业务逻辑相关错误（并发冲突、权限不足等） |

### 第三方库确认

| 库 | 原因 | 确认状态 |
|----|------|----------|
| github.com/IBM/sarama (Kafka) | Kafka 消费者生产者 | ✅ 已确认 |
| github.com/google/uuid (v7) | UUID v7 生成 | ✅ 通用库支持 |
| github.com/jinguoxing/idrm-go-base/ratelimiter | Redis 限流器 | ✅ 通用库支持 |

---

## File Structure

### 文件产出清单

| 序号 | 文件 | 生成方式 | 位置 |
|------|------|----------|------|
| 1 | API 文件 | AI 实现 | `api/doc/data_semantic/data_semantic.api` |
| 2 | DDL 文件 | AI 实现 | `migrations/data_understanding/raw/*.sql` |
| 3 | 迁移脚本 | AI 实现 | `migrations/versions/data_understanding/*.sql` |
| 4 | Handler | goctl 生成 | `api/internal/handler/data_semantic/` |
| 5 | Types | goctl 生成 | `api/internal/types/types.go` |
| 6 | Logic | AI 实现 | `api/internal/logic/data_semantic/` |
| 7 | Model | AI 实现 | `model/data_understanding/` |
| 8 | Kafka Consumer | AI 实现 | `consumer/data_understanding/` |

### 代码结构

```
api/internal/
├── handler/data_semantic/
│   ├── get_fields_handler.go
│   ├── get_business_objects_handler.go
│   ├── submit_understanding_handler.go
│   ├── delete_business_objects_handler.go
│   ├── regenerate_business_objects_handler.go
│   ├── generate_understanding_handler.go
│   ├── save_semantic_info_handler.go
│   ├── save_business_objects_handler.go
│   ├── move_attribute_handler.go
│   └── get_status_handler.go
├── logic/data_semantic/
│   ├── get_fields_logic.go
│   ├── get_business_objects_logic.go
│   ├── submit_understanding_logic.go
│   ├── delete_business_objects_logic.go
│   ├── regenerate_business_objects_logic.go
│   ├── generate_understanding_logic.go
│   ├── save_semantic_info_logic.go
│   ├── save_business_objects_logic.go
│   ├── move_attribute_logic.go
│   └── get_status_logic.go
├── types/
│   └── types.go
└── svc/
    └── servicecontext.go

model/data_understanding/
├── business_object/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
├── business_object_attributes/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
├── form_view_info_temp/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
├── form_view_field_info_temp/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
├── business_object_temp/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
├── business_object_attributes_temp/
│   ├── interface.go
│   ├── types.go
│   ├── vars.go
│   ├── factory.go
│   └── sqlx_model.go
└── kafka_message_log/
    ├── interface.go
    ├── types.go
    ├── vars.go
    ├── factory.go
    └── sqlx_model.go

consumer/data_understanding/
├── kafka_consumer.go
└── handler.go

migrations/data_understanding/raw/
├── t_business_object.sql
├── t_business_object_attributes.sql
├── t_business_object_temp.sql
├── t_business_object_attributes_temp.sql
├── t_form_view_info_temp.sql
├── t_form_view_field_info_temp.sql
├── t_kafka_message_log.sql
├── form_view_alter.sql
└── form_view_field_alter.sql

migrations/versions/data_understanding/
├── 20260203000001_init_tables.up.sql
├── 20260203000001_init_tables.down.sql
├── 20260203000002_alter_existing_tables.up.sql
└── 20260203000002_alter_existing_tables.down.sql
```

---

## Architecture Overview

### 分层架构

```
HTTP Request → Handler → Logic → Model → Database
     ↓           ↓        ↓       ↓         ↓
  参数校验    业务逻辑   数据访问  MySQL
  响应格式   事务管理
```

### 状态流转图

```
┌─────────────┐
│ 0 - 未理解   │ ◄─── [删除且无正式数据] ────┐
└──────┬──────┘                            │
       │ [一键生成]                        │
       ▼                                   │
┌─────────────┐                           │
│ 1 - 理解中   │ ──[Kafka消费]────────────► │
└──────┬──────┘                           │
       │ [AI完成]                         │
       ▼                                   │
┌─────────────┐                           │
│ 2 - 待确认   │ ──[提交]────────────────► │
└──────┬──────┘                           │
       │ [删除] ──► 若有正式数据则保持 3  │
       │                                  │
       ├────[重新识别]─────────────────────┘
       │
       ▼
┌─────────────┐              ┌─────────────┐
│ 3 - 已完成   │ ───[建模]──► │ 4 - 已发布   │
└─────────────┘              └─────────────┘
     ▲
     │ [删除且有正式数据时保持]
     └────────────────────────────┘
```

> **删除状态说明**:
> - 状态 2（待确认）删除：仅删除临时数据，若正式表有数据则保持状态 3，否则回退到状态 0
> - 状态 3（已完成）删除：仅删除临时数据，保持状态 3不变（正式表有数据）

### Kafka 集成架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   API 服务   │         │   Kafka      │         │   AI 服务    │
│             │         │              │         │             │
│ /generate   ├────────►│ -requests    ├────────►│  分析处理   │
│ /regenerate  │  发送   │              │  推送   │             │
│             │         └──────┬───────┘         └──────┬──────┘
└─────────────┘                │                        │
                               │                        │
                    ┌──────────▼──────────┐             │
                    │  -responses        │◄────────────┘
                    │                    │    接收结果
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Kafka Consumer   │
                    │                    │
                    │ - 保存临时表       │
                    │ - 更新状态         │
                    │ - 记录日志         │
                    └────────────────────┘
```

---

## Data Model

### 数据表汇总

| 表名 | 类型 | 说明 |
|------|------|------|
| `t_business_object` | 正式表 | 业务对象正式表 |
| `t_business_object_attributes` | 正式表 | 业务对象属性正式表 |
| `t_business_object_temp` | 临时表 | 业务对象临时表（版本控制） |
| `t_business_object_attributes_temp` | 临时表 | 业务对象属性临时表（版本控制） |
| `t_form_view_info_temp` | 临时表 | 库表信息临时表（版本控制） |
| `t_form_view_field_info_temp` | 临时表 | 库表字段信息临时表（版本控制） |
| `t_kafka_message_log` | 辅助表 | Kafka 消息处理记录表 |
| `form_view` | 现有表（扩展） | 数据视图表（增加 understand_status） |
| `form_view_field` | 现有表（扩展） | 字段表（增加 field_role, field_description） |

### DDL 定义

**位置**: `migrations/data_understanding/raw/`

#### 1. 业务对象表 `t_business_object`

```sql
CREATE TABLE IF NOT EXISTS t_business_object (
    id             CHAR(36)     NOT NULL                       COMMENT '业务对象UUID（主键）',
    object_name    VARCHAR(100) NOT NULL                       COMMENT '业务对象名称',
    object_type    TINYINT      NOT NULL DEFAULT 0             COMMENT '对象类型：0-候选业务对象,1-已发布业务对象',
    form_view_id   CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    status         TINYINT      NOT NULL DEFAULT 1             COMMENT '状态：0-禁用,1-启用',
    created_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at     DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (`id`),
    KEY idx_form_view_id (form_view_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象表';
```

#### 2. 业务对象属性表 `t_business_object_attributes`

```sql
CREATE TABLE IF NOT EXISTS t_business_object_attributes (
    id                   CHAR(36)     NOT NULL                       COMMENT '属性UUID（主键）',
    form_view_id         CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    business_object_id   CHAR(36)     NOT NULL                       COMMENT '关联业务对象UUID',
    form_view_field_id   CHAR(36)     NOT NULL                       COMMENT '关联字段UUID',
    attr_name            VARCHAR(100) NOT NULL                       COMMENT '属性名称',
    created_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at           DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (`id`),
    KEY idx_form_view_id (form_view_id, deleted_at),
    KEY idx_business_object_id (business_object_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象属性表';
```

#### 3. 业务对象临时表 `t_business_object_temp`

```sql
CREATE TABLE IF NOT EXISTS t_business_object_temp (
    id             CHAR(36)     NOT NULL                       COMMENT '业务对象UUID（主键）',
    form_view_id   CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    user_id        CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version        INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    object_name    VARCHAR(100) NOT NULL                       COMMENT '业务对象名称',
    created_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at     DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_version (form_view_id, version, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象临时表';
```

#### 4. 业务对象属性临时表 `t_business_object_attributes_temp`

```sql
CREATE TABLE IF NOT EXISTS t_business_object_attributes_temp (
    id                          CHAR(36)     NOT NULL                       COMMENT '属性UUID（主键）',
    form_view_id                CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    business_object_id          CHAR(36)     NOT NULL                       COMMENT '关联业务对象UUID',
    user_id                     CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version                     INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    form_view_field_id          CHAR(36)     NOT NULL                       COMMENT '关联字段UUID',
    attr_name                   VARCHAR(100) NOT NULL                       COMMENT '属性名称',
    created_at                  DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at                  DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at                  DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_object (form_view_id, business_object_id, deleted_at),
    KEY idx_form_view_version (form_view_id, version, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务对象属性临时表';
```

#### 5. 库表信息临时表 `t_form_view_info_temp`

```sql
CREATE TABLE IF NOT EXISTS t_form_view_info_temp (
    id                   CHAR(36)     NOT NULL                       COMMENT '记录UUID（主键）',
    form_view_id         CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    user_id              CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version              INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    table_business_name  VARCHAR(255)        DEFAULT NULL            COMMENT '库表业务名称',
    table_description    VARCHAR(300)        DEFAULT NULL            COMMENT '库表描述',
    created_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at           DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_version (form_view_id, version, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库表信息临时表';
```

#### 6. 库表字段信息临时表 `t_form_view_field_info_temp`

```sql
CREATE TABLE IF NOT EXISTS t_form_view_field_info_temp (
    id                   CHAR(36)     NOT NULL                       COMMENT '记录UUID（主键）',
    form_view_id         CHAR(36)     NOT NULL                       COMMENT '关联数据视图UUID',
    form_view_field_id   CHAR(36)     NOT NULL                       COMMENT '关联字段UUID',
    user_id              CHAR(36)                                         COMMENT '为空代表模型操作，不为空代表某用户操作',
    version              INT          NOT NULL DEFAULT 10            COMMENT '版本号（存储格式：10=1.0，11=1.1，每次递增1表示0.1版本）',
    field_business_name  VARCHAR(255)        DEFAULT NULL            COMMENT '字段业务名称',
    field_role           TINYINT             DEFAULT NULL            COMMENT '字段角色：1-业务主键, 2-关联标识, 3-业务状态, 4-时间字段, 5-业务指标, 6-业务特征, 7-审计字段, 8-技术字段',
    field_description    VARCHAR(300)        DEFAULT NULL            COMMENT '字段描述',
    created_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at           DATETIME(3)          DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at           DATETIME(3)          DEFAULT NULL           COMMENT '删除时间(逻辑删除)',
    PRIMARY KEY (id),
    KEY idx_form_view_version (form_view_id, version, deleted_at),
    KEY idx_form_view_field (form_view_field_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库表字段信息临时表';
```

#### 7. Kafka 消息处理记录表 `t_kafka_message_log`

```sql
CREATE TABLE IF NOT EXISTS t_kafka_message_log (
    id CHAR(36) NOT NULL COMMENT '主键UUID',
    message_id CHAR(36) NOT NULL COMMENT 'Kafka消息ID',
    form_view_id CHAR(36) NOT NULL COMMENT '关联数据视图UUID',
    processed_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '处理时间',
    status TINYINT DEFAULT 1 COMMENT '状态：1-处理成功，2-处理失败',
    error_msg TEXT COMMENT '错误信息',
    PRIMARY KEY (id),
    UNIQUE KEY uk_message_id (message_id),
    KEY idx_form_view_id (form_view_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Kafka消息处理记录表';
```

#### 8. form_view 表扩展

```sql
ALTER TABLE form_view
ADD COLUMN IF NOT EXISTS understand_status TINYINT NOT NULL DEFAULT 0 COMMENT '理解状态：0-未理解,1-理解中,2-待确认,3-已完成,4-已发布,5-理解失败';
```

#### 9. form_view_field 表扩展

```sql
ALTER TABLE form_view_field
ADD COLUMN IF NOT EXISTS field_role TINYINT DEFAULT NULL COMMENT '字段角色：1-业务主键, 2-关联标识, 3-业务状态, 4-时间字段, 5-业务指标, 6-业务特征, 7-审计字段, 8-技术字段' AFTER business_name,
ADD COLUMN IF NOT EXISTS field_description VARCHAR(300) DEFAULT NULL COMMENT '字段描述' AFTER comment;
```

### Go Struct 定义

```go
// BusinessObject 业务对象
type BusinessObject struct {
    Id          string     `db:"id" json:"id"`
    ObjectName  string     `db:"object_name" json:"object_name"`
    ObjectType  int8        `db:"object_type" json:"object_type"`
    FormViewId  string     `db:"form_view_id" json:"form_view_id"`
    Status      int8        `db:"status" json:"status"`
    CreatedAt   time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// BusinessObjectAttribute 业务对象属性
type BusinessObjectAttribute struct {
    Id                 string     `db:"id" json:"id"`
    FormViewId         string     `db:"form_view_id" json:"form_view_id"`
    BusinessObjectId   string     `db:"business_object_id" json:"business_object_id"`
    FormViewFieldId    string     `db:"form_view_field_id" json:"form_view_field_id"`
    AttrName           string     `db:"attr_name" json:"attr_name"`
    CreatedAt          time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt          *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// BusinessObjectTemp 业务对象临时表
type BusinessObjectTemp struct {
    Id          string     `db:"id" json:"id"`
    FormViewId  string     `db:"form_view_id" json:"form_view_id"`
    UserId      *string    `db:"user_id" json:"user_id,omitempty"`
    Version     int        `db:"version" json:"version"`
    ObjectName  string     `db:"object_name" json:"object_name"`
    CreatedAt   time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// BusinessObjectAttributeTemp 业务对象属性临时表
type BusinessObjectAttributeTemp struct {
    Id                 string     `db:"id" json:"id"`
    FormViewId         string     `db:"form_view_id" json:"form_view_id"`
    BusinessObjectId   string     `db:"business_object_id" json:"business_object_id"`
    UserId             *string    `db:"user_id" json:"user_id,omitempty"`
    Version            int        `db:"version" json:"version"`
    FormViewFieldId    string     `db:"form_view_field_id" json:"form_view_field_id"`
    AttrName           string     `db:"attr_name" json:"attr_name"`
    CreatedAt          time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt          *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// FormViewInfoTemp 库表信息临时表
type FormViewInfoTemp struct {
    Id                  string     `db:"id" json:"id"`
    FormViewId          string     `db:"form_view_id" json:"form_view_id"`
    UserId              *string    `db:"user_id" json:"user_id,omitempty"`
    Version             int        `db:"version" json:"version"`
    TableBusinessName   *string    `db:"table_business_name" json:"table_business_name,omitempty"`
    TableDescription    *string    `db:"table_description" json:"table_description,omitempty"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt           *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// FormViewFieldInfoTemp 字段信息临时表
type FormViewFieldInfoTemp struct {
    Id                  string     `db:"id" json:"id"`
    FormViewId          string     `db:"form_view_id" json:"form_view_id"`
    FormViewFieldId     string     `db:"form_view_field_id" json:"form_view_field_id"`
    UserId              *string    `db:"user_id" json:"user_id,omitempty"`
    Version             int        `db:"version" json:"version"`
    FieldBusinessName   *string    `db:"field_business_name" json:"field_business_name,omitempty"`
    FieldRole           *int8      `db:"field_role" json:"field_role,omitempty"`
    FieldDescription    *string    `db:"field_description" json:"field_description,omitempty"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    DeletedAt           *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// KafkaMessageLog Kafka消息处理记录
type KafkaMessageLog struct {
    Id          string     `db:"id" json:"id"`
    MessageId   string     `db:"message_id" json:"message_id"`
    FormViewId  string     `db:"form_view_id" json:"form_view_id"`
    ProcessedAt time.Time  `db:"processed_at" json:"processed_at"`
    Status      int8       `db:"status" json:"status"`
    ErrorMsg    *string    `db:"error_msg" json:"error_msg,omitempty"`
}
```

---

## API Contract

**位置**: `api/doc/data_semantic/data_semantic.api`

```api
syntax = "v1"

import "../base.api"

type (
    // ====== 通用响应 ======
    // 字段语义信息
    FieldSemanticInfo {
        FormViewFieldId   string  `json:"form_view_field_id"`
        FieldBusinessName  *string `json:"field_business_name"`
        FieldTechName      string  `json:"field_tech_name"`
        FieldType          string  `json:"field_type"`
        FieldRole          *int8   `json:"field_role"`
        FieldDescription   *string `json:"field_description"`
    }

    // 业务对象属性
    BusinessObjectAttribute {
        Id                 string  `json:"id"`
        AttrName           string  `json:"attr_name"`
        FormViewFieldId    string  `json:"form_view_field_id"`
        FieldTechName      string  `json:"field_tech_name"`
        FieldBusinessName  *string `json:"field_business_name"`
        FieldRole          *int8   `json:"field_role"`
        FieldType          string  `json:"field_type"`
    }

    // 业务对象
    BusinessObject {
        Id         string                   `json:"id"`
        ObjectName string                   `json:"object_name"`
        Attributes []BusinessObjectAttribute `json:"attributes"`
    }

    // ====== 接口1: 查询字段语义补全数据 ======
    GetFieldsReq {
        Id              string  `path:"id" validate:"required"`
        Keyword         *string `form:"keyword"`
        OnlyIncomplete *bool   `form:"only_incomplete"`
    }
    GetFieldsResp {
        CurrentVersion      int                 `json:"current_version"`
        TableBusinessName   *string             `json:"table_business_name"`
        TableTechName       string              `json:"table_tech_name"`
        TableDescription    *string             `json:"table_description"`
        Fields              []FieldSemanticInfo `json:"fields"`
    }

    // ====== 接口2: 查询业务对象识别结果 ======
    GetBusinessObjectsReq {
        Id        *string `path:"id" validate:"required"`
        ObjectId *string `form:"object_id"`
        Keyword  *string `form:"keyword"`
    }
    GetBusinessObjectsResp {
        CurrentVersion int             `json:"current_version"`
        List           []BusinessObject `json:"list"`
    }

    // ====== 接口3: 提交理解数据 ======
    SubmitUnderstandingReq {
        Id string `path:"id" validate:"required"`
    }
    SubmitUnderstandingResp {
        UnderstandStatus int8 `json:"understand_status"`
    }

    // ====== 接口4: 删除业务对象识别结果 ======
    DeleteBusinessObjectsReq {
        Id string `path:"id" validate:"required"`
    }
    DeleteBusinessObjectsResp {
        UnderstandStatus int8 `json:"understand_status"`
    }

    > **删除后状态说明**:
    > - 状态 2（待确认）：删除临时表数据后，若正式表 `t_business_object` 存在数据则保持状态 3（已完成），否则回退到状态 0（未理解）
    > - 状态 3（已完成）：删除临时表数据后，保持状态 3（已完成）不变（正式表已有数据）
    > - 返回 `understand_status` 为更新后的状态值

    // ====== 接口5: 重新识别业务对象 ======
    RegenerateBusinessObjectsReq {
        Id string `path:"id" validate:"required"`
    }
    RegenerateBusinessObjectsResp {
        ObjectCount    int `json:"object_count"`
        AttributeCount int `json:"attribute_count"`
    }

    // ====== 接口6: 一键生成理解数据 ======
    GenerateUnderstandingReq {
        Id string `path:"id" validate:"required"`
    }
    GenerateUnderstandingResp {
        UnderstandStatus int8 `json:"understand_status"`
    }

    // ====== 接口7: 保存库表信息补全数据 ======
    SaveSemanticInfoReq {
        Id string `path:"id" validate:"required"`
        TableData *struct {
            Id                 *string `json:"id" validate:"required"`    // t_form_view_info_temp.id，用于 upsert 操作
            TableBusinessName *string `json:"table_business_name" validate:"omitempty,max=255"`
            TableDescription  *string `json:"table_description" validate:"omitempty,max=300"`
        } `json:"tableData"`
        FieldData []struct {
            Id                 *string `json:"id" validate:"required"`    // t_form_view_field_info_temp.id，用于 upsert 操作
            FormViewFieldId    *string `json:"form_view_field_id"`       // 关联字段UUID
            FieldBusinessName  *string `json:"field_business_name" validate:"omitempty,max=255"`
            FieldRole          *int8   `json:"field_role" validate:"omitempty,min=1,max=8"`
            FieldDescription   *string `json:"field_description" validate:"omitempty,max=300"`
        } `json:"fieldData"`
    }
    SaveSemanticInfoResp {
        Code int32 `json:"code"`
    }

    // ====== 接口8: 保存业务对象及属性 ======
    SaveBusinessObjectsReq {
        Type string `json:"type" validate:"required,oneof=object attribute"`
        Id   string `json:"id" validate:"required"`
        Name string `json:"name" validate:"required,max=100"`
    }
    SaveBusinessObjectsResp {
        Code int32 `json:"code"`
    }

    // ====== 接口9: 调整属性归属业务对象 ======
    MoveAttributeReq {
        Id                string `path:"id" validate:"required"`
        AttributeId       string `json:"attribute_id" validate:"required"`
        TargetObjectUuid  string `json:"target_object_uuid" validate:"required"`
    }
    MoveAttributeResp {
        AttributeId       string `json:"attribute_id"`
        BusinessObjectId  string `json:"business_object_id"`
    }

    // ====== 接口10: 查询库表理解状态 ======
    GetStatusReq {
        Id string `path:"id" validate:"required"`
    }
    GetStatusResp {
        UnderstandStatus int8 `json:"understand_status"`
        CurrentVersion   int  `json:"current_version"`
    }
)

@server(
    prefix: /api/v1/data-semantic
    group: data_semantic
    middleware: JwtAuth
)
service data-semantic-api {
    @doc "查询字段语义补全数据"
    @handler GetFields
    get /:id/fields (GetFieldsReq) returns (GetFieldsResp)

    @doc "查询业务对象识别结果"
    @handler GetBusinessObjects
    get /:id/business-objects (GetBusinessObjectsReq) returns (GetBusinessObjectsResp)

    @doc "提交理解数据"
    @handler SubmitUnderstanding
    post /:id/submit (SubmitUnderstandingReq) returns (SubmitUnderstandingResp)

    @doc "删除业务对象识别结果"
    @handler DeleteBusinessObjects
    delete /:id/business-objects (DeleteBusinessObjectsReq) returns (DeleteBusinessObjectsResp)

    @doc "重新识别业务对象"
    @handler RegenerateBusinessObjects
    post /:id/business-objects/regenerate (RegenerateBusinessObjectsReq) returns (RegenerateBusinessObjectsResp)

    @doc "一键生成理解数据"
    @handler GenerateUnderstanding
    post /:id/generate (GenerateUnderstandingReq) returns (GenerateUnderstandingResp)

    @doc "保存库表信息补全数据"
    @handler SaveSemanticInfo
    put /:id/semantic-info (SaveSemanticInfoReq) returns (SaveSemanticInfoResp)

    @doc "保存业务对象及属性"
    @handler SaveBusinessObjects
    put /:id/business-objects (SaveBusinessObjectsReq) returns (SaveBusinessObjectsResp)

    @doc "调整属性归属业务对象"
    @handler MoveAttribute
    put /:id/business-objects/attributes/move (MoveAttributeReq) returns (MoveAttributeResp)

    @doc "查询库表理解状态"
    @handler GetStatus
    get /:id/status (GetStatusReq) returns (GetStatusResp)
}
```

---

## AI 服务集成

### HTTP API 调用

**接口地址**：`/api/af-sailor-agent/v1/data_understand/view_semantic_and_business_analysis`

**请求方式**：POST

**请求格式**：

**类型**: `full_understanding` (一键生成)
```json
{
    "message_id": "uuid-request-xxx",
    "request_type": "full_understanding",
    "form_view": {
        "form_view_id": "form-view-uuid",
        "form_view_technical_name": "cowenrr",
        "form_view_business_name": "员工信息表",
        "form_view_desc": "员工基础信息表",
        "form_view_fields": [
            {
                "form_view_field_id": "field-uuid-1",
                "form_view_field_technical_name": "id",
                "form_view_field_business_name": "ID",
                "form_view_field_type": "BIGINT",
                "form_view_field_role": "1",
                "form_view_field_desc": "主键ID"
            },
            {
                "form_view_field_id": "field-uuid-2",
                "form_view_field_technical_name": "name",
                "form_view_field_business_name": "姓名",
                "form_view_field_type": "VARCHAR",
                "form_view_field_role": "2",
                "form_view_field_desc": "员工姓名"
            }
        ]
    }
}
```

**类型**: `regenerate_business_objects` (重新识别)
```json
{
    "message_id": "uuid-xxx",
    "request_type": "regenerate_business_objects",
    "form_view": {
        "form_view_id": "form-view-uuid",
        "form_view_technical_name": "cowenrr",
        "form_view_business_name": "员工信息表",
        "form_view_desc": "员工基础信息表",
        "form_view_fields": [
            {
                "form_view_field_id": "field-uuid-1",
                "form_view_field_technical_name": "name",
                "form_view_field_business_name": "姓名",
                "form_view_field_type": "VARCHAR",
                "form_view_field_role": "2",
                "form_view_field_desc": "人员姓名"
            }
        ]
    }
}
```

**响应**：接口立即返回"任务处理中"（不等待AI处理完成）

**响应体结构**：
```json
{
  "task_id": "string",
  "status": "pending|running|completed|failed|cancelled",
  "message": "string",
  "message_id": "string"
}
```

**异常处理**：
- HTTP 响应码 ≠ 200 → 提示"服务异常，请稍后再试"，状态保持 `1-理解中`
- HTTP 响应码 = 200 且 status = "failed" → 提示"服务异常，请稍后再试"，状态保持 `1-理解中`
- HTTP 响应码 = 200 且 status ≠ "failed" → 正常，AI 服务异步处理中，状态保持 `1-理解中`

**说明**：无论 AI 服务调用成功与否，状态都保持 `1-理解中`，由 Kafka 消费者处理完成后更新为 `2-待确认`。

---

## Kafka Integration

### Topic 配置

| Topic | 用途 | 分区数 |
|-------|------|--------|
| `data-understanding-responses` | AI 返回分析结果 | 3 |

**说明**：AI 服务处理完成后，将结果写入 Kafka 消息到 `data-understanding-responses` 主题，由我们的服务消费处理。

### 响应消息格式

**Topic**: `data-understanding-responses`

**成功时的响应**：
```json
{
    "message_id": "uuid-request-xxx",
    "form_view_id": "form-view-uuid",
    "request_type": "full_understanding",
    "status": "success",
    "process_time": "2026-01-30T10:00:05.000Z",
    "data": {
        "table_semantic": {
            "table_business_name": "员工信息表",
            "table_description": "用于存储员工的基础信息，包括姓名、身份证、联系方式等"
        },
        "fields_semantic": [
            {
                "form_view_field_id": "field-uuid-1",
                "field_business_name": "员工ID",
                "field_role": 1,
                "field_description": "员工唯一标识"
            },
            {
                "form_view_field_id": "field-uuid-2",
                "field_business_name": "员工姓名",
                "field_role": 2,
                "field_description": "员工真实姓名"
            }
        ],
        "no_pattern_fields": [
            {
                "form_view_field_id": "field-uuid-1",
                "field_business_name": "员工ID",
                "field_role": 1,
                "field_description": "员工唯一标识"
            }
        ],
        "business_objects": [
            {
                "object_name": "基础信息",
                "attributes": [
                    {
                        "form_view_field_id": "field-uuid-2",
                        "attr_name": "姓名"
                    },
                    {
                        "form_view_field_id": "field-uuid-3",
                        "attr_name": "身份证"
                    }
                ]
            },
            {
                "object_name": "联系方式",
                "attributes": [
                    {
                        "form_view_field_id": "field-uuid-4",
                        "attr_name": "手机号"
                    }
                ]
            }
        ]
    }
}
```

**`regenerate_business_objects` 类型成功时的响应**：
```json
{
    "message_id": "uuid-xxx",
    "form_view_id": "form-view-uuid",
    "request_type": "regenerate_business_objects",
    "status": "success",
    "process_time": "2026-01-30T10:00:05.000Z",
    "data": {
        "table_semantic": {
            "table_business_name": "员工信息表",
            "table_description": "用于存储员工的基础信息，包括姓名、身份证、联系方式等"
        },
        "fields_semantic": [
            {
                "form_view_field_id": "field-uuid-1",
                "field_business_name": "员工ID",
                "field_role": 1,
                "field_description": "员工唯一标识"
            },
            {
                "form_view_field_id": "field-uuid-2",
                "field_business_name": "员工姓名",
                "field_role": 2,
                "field_description": "员工真实姓名"
            }
        ],
        "no_pattern_fields": [
            {
                "form_view_field_id": "field-uuid-1",
                "field_business_name": "员工ID",
                "field_role": 1,
                "field_description": "员工唯一标识"
            }
        ],
        "business_objects": [
            {
                "object_name": "基础信息",
                "attributes": [
                    {
                        "form_view_field_id": "xxx",
                        "attr_name": "姓名"
                    }
                ]
            }
        ]
    }
}
```

**失败时的响应**：
```json
{
    "message_id": "uuid-request-xxx",
    "form_view_id": "form-view-uuid",
    "request_type": "full_understanding",
    "status": "failed",
    "process_time": "2026-01-30T10:00:05.000Z",
    "error": {
        "code": "AI_SERVICE_ERROR",
        "message": "AI 服务处理失败的具体原因"
    }
}
```

**消费者处理逻辑**：
1. **解析消息**：获取 message_id、form_view_id、request_type、status、data/error
2. **校验 message_id**：检查是否已处理，防止重复消费
3. **校验库表理解状态**：检查当前 understand_status 是否为 `1-理解中`，否则跳过处理
4. **根据 status 处理**：
   - `status = "success"` → 保存数据到临时表，更新 `understand_status` 为 `2-待确认`
   - `status = "failed"` → 记录错误日志到 `t_kafka_message_log`，更新 `understand_status` 为 `5-理解失败`

### 消费者配置

| 配置项 | 值 |
|--------|-----|
| Topic | `data-understanding-responses` |
| Group ID | `data-understanding-consumer-group` |
| Auto Commit | false |
| Initial Offset | earliest |

---

## Error Codes

## 结构化日志格式

根据 AC-27 要求，Kafka 消费者处理失败时需记录结构化日志（JSON 格式）：

```json
{
  "timestamp": "2026-02-03T10:30:45.123Z",
  "level": "error",
  "message": "Kafka message processing failed",
  "context": {
    "message_id": "uuid-request-xxx",
    "form_view_id": "form-view-uuid",
    "topic": "data-understanding-responses",
    "partition": 0,
    "offset": 12345
  },
  "error": {
    "type": "AIAnalysisError",
    "message": "AI service returned invalid response format",
    "details": "Missing 'business_objects' field in response"
  }
}
```

**使用方式** (在 Consumer 中):
```go
import "github.com/jinguoxing/idrm-go-base/telemetry"

// 记录结构化日志
telemetry.ErrorWithContext(ctx, "Kafka message processing failed",
    telemetry.String("message_id", messageID),
    telemetry.String("form_view_id", formViewID),
    telemetry.Any("error", err),
)
```

## 业务逻辑说明

### 删除业务对象状态流转

删除业务对象临时数据时的状态处理逻辑：

```sql
-- 删除临时表数据（逻辑删除）
UPDATE t_business_object_temp
SET deleted_at = NOW()
WHERE form_view_id = ? AND deleted_at IS NULL;

UPDATE t_business_object_attributes_temp
SET deleted_at = NOW()
WHERE form_view_id = ? AND deleted_at IS NULL;

-- 根据是否有正式数据更新状态
UPDATE form_view
SET understand_status = CASE
    WHEN (
        SELECT COUNT(*)
        FROM t_business_object
        WHERE form_view_id = ?
        AND deleted_at IS NULL
    ) > 0 THEN 3  -- 正式表有数据，保持已完成
    ELSE 0        -- 正式表无数据，回退到未理解
END
WHERE id = ?;
```

### 状态流转规则总结

| 当前状态 | 删除后 | 条件 |
|---------|--------|------|
| 2-待确认 | 3-已完成 | 正式表 `t_business_object` 存在数据 |
| 2-待确认 | 0-未理解 | 正式表 `t_business_object` 无数据 |
| 3-已完成 | 3-已完成 | 正式表 `t_business_object` 存在数据（保持） |

---

### 自定义错误码 (600101-600130)

| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 600101 | ErrInvalidStatus | 当前状态不允许操作 |
| 600102 | ErrStatusUnderstandingInProgress | 当前正在理解中 |
| 600103 | ErrStatusNotReady | 数据尚未准备就绪 |
| 600104 | ErrStatusAlreadyPublished | 已发布，无法修改 |
| 600111 | ErrDuplicateObjectName | 业务对象名称重复 |
| 600112 | ErrDuplicateAttrName | 属性名称重复 |
| 600113 | ErrObjectNotFound | 业务对象不存在 |
| 600114 | ErrNoDataToDelete | 没有可删除的数据 |
| 600120 | ErrRateLimitExceeded | 操作过于频繁（限流） |
| 600121 | ErrConcurrentOperation | 并发操作冲突 |
| 600122 | ErrAIAPIFailed | AI 服务 API 调用失败 |
| 600123 | ErrAIServiceUnavailable | AI 服务不可用 |

---

## Testing Strategy

| 类型 | 方法 | 覆盖目标 |
|------|------|----------|
| 单元测试 | 表驱动测试，Mock Model | > 80% |
| 集成测试 | 测试数据库 + Mock Kafka | 核心流程 |

### 测试用例清单

| 接口 | 测试场景 |
|------|----------|
| GetFields | 状态0/2/3查询、keyword过滤、only_incomplete过滤 |
| GetBusinessObjects | 状态0/2/3查询、object_id过滤、keyword过滤 |
| SubmitUnderstanding | 状态2提交、状态3提交、状态校验失败 |
| DeleteBusinessObjects | 状态2删除、状态3删除、有正式数据时保持状态3、无正式数据时回退0、状态校验失败 |
| RegenerateBusinessObjects | 状态2/3重新识别、版本号递增 |
| GenerateUnderstanding | 状态0/3生成、状态校验、HTTP API调用 |
| SaveSemanticInfo | 库表信息保存、字段信息保存、状态校验 |
| SaveBusinessObjects | 业务对象名称保存、属性名称保存、重复校验 |
| MoveAttribute | 属性移动、目标不存在、重复名称 |
| GetStatus | 正常查询、版本号查询 |

---

## Research Findings

### Kafka 消息格式

**决策**: 使用 JSON 格式，message_id 作为去重标识

**理由**:
- JSON 格式易于调试和扩展
- message_id 唯一标识防止重复处理
- 与 AI 团队约定的格式一致

### Redis 限流配置

**决策**: 使用 `ratelimiter` 包实现滑动窗口限流，Key 格式 `data-semantic:{form_view_id}:{user_id}`

**安装**:
```bash
go get github.com/jinguoxing/idrm-go-base/ratelimiter
```

**实现方式**:
```go
import "github.com/jinguoxing/idrm-go-base/ratelimiter"

// 在 Logic 层初始化限流器
limiter := ratelimiter.NewRateLimiter(
    redisClient,
    ratelimiter.WithWindow(time.Second),  // 1秒窗口
    ratelimiter.WithMaxRequests(1),       // 最大1次请求
)

// 限流检查
key := fmt.Sprintf("data-semantic:%s:%s", formViewId, userId)
allowed, err := limiter.Allow(ctx, key)
if err != nil {
    return nil, errorx.NewWithCode(errorx.ErrCodeInternalError)
}
if !allowed {
    return nil, errorx.New(600121, "操作过于频繁，请稍后再试")
}
```

**理由**:
- `ratelimiter` 包是通用库提供的标准限流组件
- 滑动窗口算法精确控制请求频率
- 防止用户快速重复点击
- 精确到用户和库表级别
- 1 秒窗口覆盖人工操作防抖需求

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-03 | - | 初始版本，基于 807707-库表数据理解方案设计.md |
| 1.1 | 2026-02-06 | - | Kafka 响应格式更新：regenerate_business_objects 也返回完整数据；消费者增加库表状态检查 |

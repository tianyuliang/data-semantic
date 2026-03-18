# 批量业务对象匹配接口 Specification

> **Branch**: `feature/batch-object-match`
> **Spec Path**: `specs/batch-object-match/`
> **Created**: 2026-03-18

> **Status**: Completed

---

## Overview

批量业务对象匹配接口 - 给定一批业务对象（有对象名字或视图ID），有视图ID时原样返回，有对象名字时优先从业务对象表匹配，未匹配时从视图表模糊查询并触发理解。

---

## Clarifications

### Session 2026-03-18

- Q: AI推荐的业务对象是否持久化？ → A: 仅作为推荐结果返回，不持久化
- Q: source_object_name 为空字符串时的处理逻辑？ → A: 视为无匹配，返回空 hits，走AI推荐流程
- Q: 接口是否需要权限验证？ → A: 需要登录验证（JWT）
- Q: AI语义理解的匹配方式？ → A: 简单文本模糊匹配（business_name LIKE %keyword% 或 description LIKE %keyword%）
- Q: 二次调用时如何传递"正在理解的视图ID"？ → A: 无需传递，每次调用都自动重新检查所有视图状态直到理解完成

---

## User Stories

### Story 1: 批量业务对象查询匹配 (P1)

AS a 前端用户
I WANT 输入一批业务对象名称，获取对应的视图ID和mdl_id
SO THAT 快速建立业务对象与视图的关联

**独立测试**: 调用接口，输入业务对象名称列表，返回匹配的业务对象信息

### Story 2: 未匹配视图触发理解 (P1)

AS a 前端用户
I WANT 对于无法匹配的视图，触发语义理解
SO THAT 自动识别视图对应的业务对象

**独立测试**: 调用接口，未匹配的视图触发理解，返回正在理解的视图ID列表

### Story 3: 自动等待理解完成 (P2)

AS a 前端用户
I WANT 重复调用接口直到所有视图理解完成
SO THAT 获取最终匹配结果

**独立测试**: 重复调用接口，直到 understanding 为空

---

## Acceptance Criteria (EARS)

### 正常流程

| ID | Scenario | Trigger | Expected Behavior |
|----|----------|---------|-------------------|
| AC-01 | data_source.id已存在 | WHEN 输入包含 data_source.id | THE SYSTEM SHALL 直接追加到结果，返回id、name、object_name |
| AC-02 | 业务对象已匹配 | WHEN 业务对象名称能匹配到 business_object 表 | THE SYSTEM SHALL 返回 mdl_id(视图id)、TechnicalName(视图名)、object_name |
| AC-03 | 视图已理解 | WHEN 业务对象名称匹配 form_view.business_name 且 understand_status=3 | THE SYSTEM SHALL 返回视图对应的业务对象 |
| AC-04 | 视图未理解-记录视图ID | WHEN form_view.understand_status != 3 | THE SYSTEM SHALL 无论什么状态都追加视图到data_source，仅当status!=3时记录视图ID到need_understand数组 |
| AC-05 | 重复调用-理解完成 | WHEN 调用时所有视图已理解完成 | THE SYSTEM SHALL 返回匹配结果，need_understand 为空 |
| AC-06 | 重复调用-理解未完成 | WHEN 调用时仍有视图未理解 | THE SYSTEM SHALL 返回空的data_source，need_understand 包含未理解的视图ID |

### 异常处理

| ID | Scenario | Trigger | Expected Behavior |
|----|----------|---------|-------------------|
| AC-10 | 列表为空 | WHEN 批量列表为空数组 | THE SYSTEM SHALL 返回 400 错误 |
| AC-11 | 完全无匹配 | WHEN 所有输入都无法匹配 | THE SYSTEM SHALL 返回空 data_source 和空的 need_understand 列表 |

---

## Processing Flow

```
输入: []Entry (name 或 data_source)
    │
    ▼
┌─────────────────────────────┐
│ 遍历每个 Entry               │
└─────────────────────────────┘
    │
    ▼
┌─────────────────────────────┐
│ name 为空?                   │
│     Yes → 跳过该条目        │
└─────────────────────────────┘
    │ No
    ▼
┌─────────────────────────────┐
│ 1. data_source 有值?        │
│     Yes → 直接追加到结果    │
└─────────────────────────────┘
    │
    ▼ No
┌─────────────────────────────┐
│ 2. 业务对象表模糊匹配       │
│    business_object.object_name │
│    LIKE %name%              │
└─────────────────────────────┘
    │
    ▼ 找到
┌─────────────────────────────┐
│ 返回: form_view_id, name    │
└─────────────────────────────┘
    │
    ▼ 未找到
┌─────────────────────────────┐
│ 3. 视图表模糊匹配           │
│    form_view.business_name   │
│    LIKE %name%              │
└─────────────────────────────┘
    │
    ▼ 找到
┌─────────────────────────────┐
│ 检查 understand_status       │
│   = 3 (已理解)               │
│     → 查询 business_object   │
│     → 获取 object_name       │
│   ≠ 3 (未理解)               │
│     → 记录 form_view_id     │
│       到 need_understand     │
│   无论什么状态都追加视图     │
│     到 data_source           │
└─────────────────────────────┘
    │
    ▼ 未找到
┌─────────────────────────────┐
│ 返回空 data_source          │
└─────────────────────────────┘
    │
    ▼
输出: []Entry
    - name: 原始输入名称
    - data_source: 匹配的视图列表
    - need_understand (顶层): 需要理解的视图ID数组（去重）
```

---

## Edge Cases

| ID | Case | Expected Behavior |
|----|------|-------------------|
| EC-01 | 批量100条 | 逐条处理，返回100条结果 |
| EC-02 | 多次调用同一批 | 每次调用都重新检查状态，直到理解完成 |
| EC-03 | 部分理解完成 | 返回已完成的hits和未完成的视图ID |

---

## Data Considerations

### 输入数据结构

| Field | Description | Constraints |
|-------|-------------|-------------|
| entries | 业务对象列表 | 必填，最多100条 |
| entries[].name | 业务对象名称 | 必填，非空字符串 |
| entries[].data_source | 给定的视图数据 | 可选，有值时直接追加到结果 |
| entries[].data_source.id | 视图ID | UUID格式 |
| entries[].data_source.name | 视图名称 | 字符串 |

### 输出数据结构

| Field | Description | Constraints |
|-------|-------------|-------------|
| entries | 匹配结果列表 | |
| entries[].name | 原始输入名称 | |
| entries[].data_source | 匹配的视图列表 | 无匹配时为空数组 |
| entries[].data_source[].id | 视图ID (mdl_id) | UUID格式 |
| entries[].data_source[].name | 视图名称 (TechnicalName) | 字符串 |
| entries[].data_source[].object_name | 业务对象名称 | 字符串 |
| need_understand | 需要理解的视图ID列表 | 顶层字段，用于外部监听，已去重 |

---

## Success Metrics

| ID | Metric | Target |
|----|--------|--------|
| SC-01 | 接口响应时间 | < 500ms (P99) |
| SC-02 | 测试覆盖率 | > 80% |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-18 | - | 初始版本 |
| 1.1 | 2026-03-18 | - | 更新协议：list→entries, object_name→name, data_source结构, understanding→need_understand |
| 1.2 | 2026-03-18 | - | 更新响应结构：增加object_name字段，data_source返回mdl_id |
| 1.3 | 2026-03-18 | - | Step4无论状态都追加视图到data_source，输入验证name非空 |

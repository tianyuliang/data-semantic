# 批量业务对象匹配接口 Technical Plan

> **Branch**: `feature/batch-object-match`
> **Spec Path**: `specs/batch-object-match/`
> **Created**: 2026-03-18
> **Status**: Completed

---

## Summary

批量业务对象匹配接口，入参为业务对象名称或视图ID，有视图ID时原样返回，有对象名字时优先从 business_object 表模糊匹配，未匹配时从 form_view 表模糊匹配并根据 understand_status 决定是否触发理解。

---

## Technical Context

| Item | Value |
|------|-------|
| **Language** | Go 1.24+ |
| **Framework** | Go-Zero v1.9+ |
| **Storage** | MySQL 8.0 |
| **ORM** | GORM / SQLx |
| **Testing** | go test |
| **Common Lib** | idrm-go-base v0.1.0+ |

> **重要约束**：不允许修改现有代码，新功能独立实现，不影响原有 API 逻辑

---

## Go-Zero 开发流程

| Step | 任务 | 方式 | 产出 |
|------|------|------|------|
| 1 | 定义 API 文件 | AI 实现 | `api/doc/data_semantic/batch_object_match.api` |
| 2 | 生成 Handler/Types | goctl | `api/internal/handler/`, `types/` |
| 3 | 实现 Logic 层 | AI 实现 | `api/internal/logic/data_semantic/` |

**goctl 命令**:
```bash
goctl api go -api api/doc/api.api -dir api/ --style=go_zero --type-group
```

---

## File Structure

| 序号 | 文件 | 位置 |
|------|------|------|
| 1 | API 文件 | `api/doc/data_semantic/data_semantic.api` |
| 2 | Handler | `api/internal/handler/data_semantic/batch_object_match_handler.go` |
| 3 | Types | `api/internal/types/` |
| 4 | Logic | `api/internal/logic/data_semantic/batch_object_match_logic.go` |

---

## Architecture

```
HTTP Request → Handler → Logic → Model → Database
```

| 层级 | 职责 |
|------|------|
| Handler | 解析参数、格式化响应 |
| Logic | 业务逻辑实现（匹配流程控制） |
| Model | 数据访问（查询 business_object / form_view） |

---

## API Contract

**位置**: `api/doc/data_semantic/data_semantic.api`

```api
type (
    // 批量匹配请求
    BatchObjectMatchReq {
        Entries []SourceObject `json:"entries" validate:"required,min=1,max=100"`
    }

    // 源对象
    SourceObject {
        Name       string             `json:"name" validate:"required,max=100"`    // 业务对象名称
        DataSource *RequestDataSource `json:"data_source,optional"`              // 给定的视图数据
    }

    // 请求中的视图数据
    RequestDataSource {
        Id   string `json:"id"`   // 视图ID（对应form_view表mdl_id）
        Name string `json:"name"` // 视图名称
    }

    // 批量匹配响应
    BatchObjectMatchResp {
        Entries        []MatchResult `json:"entries"`
        NeedUnderstand []string      `json:"need_understand"` // 需要理解的视图ID列表（去重）
    }

    // 匹配结果
    MatchResult {
        Name       string               `json:"name"`                  // 原始输入名称
        DataSource []ResponseDataSource `json:"data_source,optional"` // 匹配的视图列表
    }

    // 响应中的视图数据
    ResponseDataSource {
        Id          string `json:"id"`           // 视图ID（mdl_id）
        Name        string `json:"name"`         // 视图名称（TechnicalName）
        ObjectName  string `json:"object_name"`  // 业务对象名称
    }
)

@server(
    prefix: /api/v1/data-semantic
    group: data_semantic
)
service api {
    @handler BatchObjectMatch
    post /batch-object-match (BatchObjectMatchReq) returns (BatchObjectMatchResp)
}
```

---

## Matching Logic

### 核心处理流程

```go
func (l *BatchObjectMatchLogic) BatchObjectMatch(req *BatchObjectMatchReq) (*BatchObjectMatchResp, error) {
    // 输入验证：过滤空name的条目
    for _, item := range req.Entries {
        if item.Name == "" {
            continue
        }
        // 处理每个条目...
    }
}

func (l *BatchObjectMatchLogic) processObject(item SourceObject) (MatchResult, []string) {
    result := MatchResult{Name: item.Name, DataSource: make([]ResponseDataSource, 0)}
    needUnderstands := make([]string, 0)

    // Step 1: data_source 有值，直接追加到结果
    if item.DataSource != nil && item.DataSource.Id != "" {
        result.DataSource = append(result.DataSource, ResponseDataSource{
            Id:         item.DataSource.Id,
            Name:       item.DataSource.Name,
            ObjectName: item.Name,
        })
        return result, needUnderstands
    }

    // Step 2: 业务对象表模糊匹配
    objects, err := l.businessObjectModel.FuzzyMatchByName(l.ctx, item.Name)
    if err != nil {
        return result, needUnderstands
    }

    if len(objects) > 0 {
        // 找到匹配，组装返回（通过 form_view_id 查询视图信息获取 mdl_id 和 TechnicalName）
        for _, obj := range objects {
            view, _ := l.formViewModel.FindOneById(l.ctx, obj.FormViewId)
            viewName := ""
            mdlId := ""
            if view != nil {
                viewName = view.TechnicalName
                mdlId = view.MdlId
            }
            result.DataSource = append(result.DataSource, ResponseDataSource{
                Id:         mdlId,
                Name:       viewName,
                ObjectName: obj.ObjectName,
            })
        }
        return result, needUnderstands
    }

    // Step 3: 视图表模糊匹配
    views, err := l.formViewModel.FuzzyMatchByName(l.ctx, item.Name)
    if err != nil {
        return result, needUnderstands
    }

    if len(views) == 0 {
        return result, needUnderstands
    }

    // Step 4: 检查每个视图的 understand_status
    for _, view := range views {
        // 无论什么状态都追加视图到结果
        objectName := ""
        if view.UnderstandStatus == 3 {
            // 已理解，查询业务对象表获取 object_name
            objs, _ := l.businessObjectModel.FindByFormViewId(l.ctx, view.Id)
            if len(objs) > 0 {
                objectName = objs[0].ObjectName
            }
        } else {
            // 未理解，记录需要理解的视图ID
            needUnderstands = append(needUnderstands, view.Id)
        }
        result.DataSource = append(result.DataSource, ResponseDataSource{
            Id:         view.MdlId,
            Name:       view.TechnicalName,
            ObjectName: objectName,
        })
    }

    return result, needUnderstands
}
```

### need_understand 逻辑

- 无论什么状态都追加视图到 data_source
- 仅当 status != 3 时记录视图ID到 need_understand
- 不触发实际理解，由外部系统决定何时触发
- 响应顶层包含汇总的 `need_understand` 数组（已去重），供外部监听使用

> **重要**：视图的理解状态**必须**从 `form_view.understand_status` 字段获取，该字段会在理解流程中自动更新流转（0-未理解,1-理解中,2-待确认,3-已完成,4-已发布）

---

## Model 层需要新增的方法

### BusinessObject Model

```go
// FuzzyMatchByName 模糊匹配业务对象名称
FuzzyMatchByName(ctx context.Context, name string) ([]*BusinessObject, error)

// FindByFormViewId 根据视图ID查找业务对象
FindByFormViewId(ctx context.Context, formViewId string) ([]*BusinessObject, error)
```

### FormView Model

```go
// FuzzyMatchByName 模糊匹配视图业务名称
FuzzyMatchByName(ctx context.Context, name string) ([]*FormView, error)

// FindOneById 根据ID查询视图（需要包含 mdl_id 字段）
FindOneById(ctx context.Context, id string) (*FormView, error)
```

> 注意：FormView Model 的 FindOneById 方法需要包含 `mdl_id` 字段

---

## 重复调用逻辑

每次调用接口时：
1. 都会重新检查所有相关视图的 understand_status
2. 如果视图已理解（status=3），返回匹配的业务对象
3. 如果视图未理解，触发理解并返回 understanding 数组
4. 前端重复调用直到 understanding 为空

---

## Testing Strategy

| 类型 | 方法 | 覆盖率 |
|------|------|--------|
| 单元测试 | Mock Model | > 80% |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-18 | - | 初始版本 |
| 1.1 | 2026-03-18 | - | 更新协议：entries列表, name字段, data_source结构, need_understand字段 |
| 1.2 | 2026-03-18 | - | 更新响应结构：增加object_name字段，data_source返回mdl_id |
| 1.3 | 2026-03-18 | - | Step4无论状态都追加视图到data_source，输入验证name非空 |

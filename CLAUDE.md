# data-semantic

基于 Go-Zero 微服务架构的项目，采用 AI 辅助的规范驱动开发 (SDD) 模式。

## ⚠️ 强制工作流 (必读)

**任何功能开发必须遵循 SDD 流程，不允许跳过！**

### 开发请求识别

当用户请求涉及以下内容时，视为"功能开发请求"：
- 新增功能、接口、模块
- 修改现有功能逻辑
- Bug 修复（非简单配置修改）
- 重构代码

### 强制行为

1. **必须使用 speckit 命令启动**
   ```
   /speckit.start <功能描述>    # 智能场景匹配
   /speckit.specify <功能描述>  # 创建规格文档
   ```

2. **禁止跳过 SDD 阶段**
   - ❌ 直接编写代码
   - ❌ 未创建 spec.md 就开始实现
   - ❌ 未经用户确认就进入下一阶段

3. **如果用户要求直接编码**
   - 停止并提醒："此请求涉及功能开发，请使用 `/speckit.start <功能描述>` 启动 SDD 流程"
   - 解释 SDD 流程的价值（可追溯、可测试、减少返工）

4. **specs 文件格式**
   - 所有 `specs/` 下的文件必须使用 `.specify/templates/` 中的模板
   - spec.md 使用 EARS 格式 (WHEN...THE SYSTEM SHALL)
   - plan.md 包含 API/DDL/Model 设计

5. **测试先行 (Test-First) 🧪**
   - ❌ 没有测试的代码不允许提交
   - ✅ 实现代码必须同时编写测试用例
   - ✅ 测试覆盖率 > 80%
   - ✅ 使用 `[TEST]` 标记测试任务

### 例外情况

以下请求可以直接处理，无需 SDD 流程：
- 配置文件修改
- 依赖版本更新
- 代码格式化
- 简单问答和解释

---

## 技术栈

- **语言**: Go 1.24+
- **框架**: Go-Zero v1.9+
- **数据库**: MySQL 8.0
- **ORM**: GORM (复杂查询) + SQLx (高性能)
- **架构**: 微服务 (API/RPC/Job/Consumer)
- **通用库**: idrm-go-base v0.1.0+

## 通用库规范 (idrm-go-base)

### 必须使用

| 场景 | 使用模块 | 禁止行为 |
|------|----------|----------|
| 错误处理 | `errorx` | ❌ 自定义 error struct |
| HTTP 响应 | `response` | ❌ 自定义响应格式 |
| API 中间件 | `middleware` | ❌ 重复实现认证/日志 |
| 参数校验 | `validator` | ❌ 手写校验逻辑 |
| 日志追踪 | `telemetry` | ❌ 直接使用 fmt/log |

### Import 路径

```go
import "github.com/jinguoxing/idrm-go-base/{module}"
```

### 引入其他库规则

如需使用通用库以外的第三方库：
- **停止** 并询问：该库是否可以使用？
- 等待确认后再继续

### 主键规范 (UUID v7)

所有表使用 UUID v7 作为主键：

```sql
`id` CHAR(36) NOT NULL COMMENT 'ID (UUID v7)'
```

```go
Id string `gorm:"primaryKey;size:36"`  // UUID v7
```

## 项目结构

```
data-semantic/
├── api/                      # API 服务
│   ├── doc/                  # API 定义 (.api 文件)
│   ├── etc/                  # 配置文件
│   └── internal/             # 内部实现
│       ├── handler/          # 请求处理 (参数校验)
│       ├── logic/            # 业务逻辑
│       ├── svc/              # 服务上下文
│       └── types/            # 类型定义
├── model/                    # 数据模型
├── migrations/               # DDL 迁移
├── specs/                    # SDD 规格文档
├── deploy/                   # 部署配置
└── .specify/                 # Spec Kit 配置
```

## 快速命令

以下命令可直接使用:

```bash
# 开发
make api           # 生成 API 代码
make swagger       # 生成 Swagger 文档
make run           # 运行服务
make test          # 运行测试

# 部署
make docker-build  # 构建镜像
make k8s-deploy    # 部署到 K8s
```

## SDD 工作流程

本项目遵循 Spec-Driven Development 5 阶段工作流:

1. **Context** - 阅读 `.specify/memory/constitution.md` 理解项目规范
2. **Specify** - 使用 EARS 格式定义需求 → `specs/<feature>/spec.md`
3. **Design** - 创建技术方案 → `specs/<feature>/plan.md`
4. **Tasks** - 拆分任务 (每个 <50 行) → `specs/<feature>/tasks.md`
5. **Implement** - 按任务顺序编码实现

**重要**: 每个阶段完成后等待用户确认，再进入下一阶段。

## 架构规范

### 分层职责 (严格遵守)

| 层 | 职责 | 禁止 |
|---|------|------|
| Handler | 参数绑定、校验、调用 Logic | 包含业务逻辑 |
| Logic | 业务逻辑、事务管理 | 直接操作 HTTP |
| Model | 数据访问 (GORM/SQLx) | 包含业务逻辑 |

### API 设计

- 入口文件: `api/doc/api.api`
- 基础类型: `api/doc/base.api`
- 模块 API: `api/doc/<module>/<module>.api`
- 使用 `goctl api go` 生成代码

## 编码约定

### 命名规范

```
文件名: snake_case.go
包名:   lowercase
结构体: PascalCase
方法:   PascalCase
变量:   camelCase
常量:   UPPER_SNAKE_CASE
```

### 错误处理

```go
import "github.com/jinguoxing/idrm-go-base/errorx"

// 使用预定义错误码
if user == nil {
    return nil, errorx.NewWithCode(errorx.ErrCodeNotFound)
}

// 自定义业务错误码 (在 internal/errorx/codes.go 定义)
if user.Status == 0 {
    return nil, errorx.New(30102, "用户已禁用")
}
```

### 日志规范

```go
// 使用 logx，包含 traceId
logx.WithContext(ctx).Infof("user login: %s", phone)
```

## 重要约束

### 必须

- ✅ Handler 使用 validator 校验参数
- ✅ Logic 层管理事务边界
- ✅ 使用配置文件管理环境变量
- ✅ 错误信息使用 errors.Wrapf 包装

### 禁止

- ❌ Handler 直接操作数据库
- ❌ Model 层包含业务判断
- ❌ 硬编码配置值
- ❌ 使用 fmt.Println 替代 logx

## 相关文档

- 项目宪法: `.specify/memory/constitution.md`
- SDD 模板: `.specify/templates/`
- Spec Kit 命令: `.cursor/commands/` 或 `.claude/commands/`

## 常见操作

### 新增 API 接口

1. 在 `api/doc/<module>/` 创建 `.api` 文件
2. 运行 `make api` 生成代码
3. 在 `api/internal/logic/` 实现业务逻辑
4. 运行 `make swagger` 更新文档

### 新增数据表

1. 在 `migrations/` 创建 DDL 文件
2. 执行 DDL 创建表
3. 使用 goctl 生成 Model 或手写 GORM Model
4. 在 Logic 层调用 Model

### 部署服务

```bash
make docker-build    # 构建镜像
make docker-push     # 推送镜像
make k8s-deploy      # 部署到 K8s
make k8s-status      # 查看状态
```

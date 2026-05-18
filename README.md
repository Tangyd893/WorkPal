# WorkPal

[English](README-en.md) | 中文说明

WorkPal 是一个基于 Go 和 React 的办公协作平台。当前版本已经按真实微服务形态运行：前端只访问 API Gateway，后端各领域服务拥有各自的运行边界和存储边界。

## 项目能力

**WorkPal 2.0** — 以项目为中心的一站式团队工作空间。

- 预置验收账号：`admin`、`emma.chen`、`liam.wang`、`sofia.zhao`
- `English / 简体中文` 双语界面 + 深浅色主题 + 界面密度
- **11 个微服务**：项目空间 / 即时通讯 / 协作文档 / 日程日历 / 审批中心 / AI 助手
- 看板拖拽排序 + 自定义工作流引擎 (Jira 式状态转换)
- RBAC 三层权限模型 (全局角色 → 项目角色 → 权限中间件)
- 自定义字段系统 (EAV 模型，文本/数字/日期/选项)
- 频道消息 ↔ Issue 双向关联
- Kafka/Elasticsearch 事件驱动全文本搜索
- 消息幂等保护 + 冷热分离 + Saga 编排
- OpenTelemetry 全链路追踪 + Prometheus/Grafana/Jaeger 可观测性
- Kubernetes 部署 + 响应式移动端适配 + Electron 桌面应用配置

## 技术栈

- **后端：** Go 1.22、Gin、GORM、PostgreSQL 16、Redis 7、Kafka/Redpanda、Elasticsearch 8
- **前端：** React 18、Vite 5、TypeScript 5.4、Zustand 4.5
- **协同编辑：** TipTap/ProseMirror 编辑器 (预留 Yjs CRDT)
- **文件存储：** MinIO，带本地文件回退
- **实时通信：** WebSocket，IM 服务通过 Redis Pub/Sub 做多实例消息扇出
- **可观测性：** OpenTelemetry SDK（Go + JS）、Jaeger、Prometheus、Grafana、zerolog 结构化日志
- **容器编排：** Docker Compose、Kubernetes（Deployment + Service + 探针 + 滚动更新）
- **性能测试：** k6 / wrk 压测脚本

## 默认端口

| 组件 | 地址 | 说明 |
| --- | --- | --- |
| 前端 | `http://localhost:3000` | Vite 开发服务器 |
| API Gateway | `http://localhost:8080` | 前端唯一后端入口 |
| User Service | `http://localhost:8081` | 认证、用户、RBAC 权限 |
| IM Service | `http://localhost:8082` | 会话、消息、频道、WebSocket |
| File Service | `http://localhost:8083` | 个人文件、群文件 |
| Search Service | `http://localhost:8084` | Elasticsearch 全文搜索 |
| Workspace Service | `http://localhost:8085` | 任务管理 |
| **Project Service** | `http://localhost:8086` | 项目/看板/工作流/自定义字段/效能报表 |
| **Docs Service** | `http://localhost:8087` | 协作文档/版本历史/知识库 |
| **Calendar Service** | `http://localhost:8088` | 日程日历/视频会议室 |
| **Approval Service** | `http://localhost:8089` | 审批模板/实例/审批流 |
| Notification Service | `http://localhost:8090` | 多通道消息通知 |
| **AI Service** | `http://localhost:8091` | 智能搜索/任务摘要 |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |
| Prometheus | `http://localhost:9090` | 指标采集与查询 |
| Grafana | `http://localhost:3001` | 监控看板 |
| Jaeger | `http://localhost:16686` | 分布式追踪 UI |

## 微服务边界

| 服务 | 存储边界 | 主要职责 |
| --- | --- | --- |
| Gateway | 无状态 | 统一入口、路由目录、服务发现、限流、重试、熔断 |
| User Service | `workpal_user` | 登录、用户、部门、RBAC 权限模型、审计日志 |
| IM Service | `workpal_im` | 频道、私聊、消息、WebSocket、outbox、幂等去重 |
| File Service | `workpal_file` | 文件元数据、上传、分享、审计日志 |
| Search Service | ES + Kafka | 全文搜索、事件驱动索引 |
| Workspace Service | `workpal_workspace` | 任务管理、Saga 编排 |
| **Project Service** | `workpal_project` | 项目/看板/Issue/工作流引擎/自定义字段/效能报表 |
| **Docs Service** | `workpal_docs` | 协作文档/版本历史/知识库 |
| **Calendar Service** | `workpal_calendar` | 日程日历/视频会议室 |
| **Approval Service** | `workpal_approval` | 审批模板/实例/审批流 |
| Notification Service | `workpal_notification` | 多通道消息通知 |
| **AI Service** | 无状态 | 智能搜索/任务摘要 (LLM 网关预留) |

## 数据库迁移

每个服务在 `backend/migrations/<service>/` 下配有版本化 SQL 迁移：

| 服务 | 迁移 | 核心表 |
| --- | --- | --- |
| user-service | `001_init`, `002_rbac` | users, departments, employees, roles, permissions, user_roles, project_roles, project_members |
| im-service | `001-004` | conversations, messages, channels, threads, audit_logs |
| file-service | `001-002` | files, audit_logs |
| workspace-service | `001-003` | tasks, schedule_events, task_sagas |
| **project-service** | `001-002` | projects, issues, workflows, boards, versions, custom_field_defs, custom_field_values, associations, issue_changelogs |
| **docs-service** | `001` | documents, document_revisions |
| **calendar-service** | `001` | calendar_events, calendar_attendees |
| **approval-service** | `001` | approval_templates, approval_instances, approval_actions |
| notification-service | `001` | notifications |

手动运行迁移：

```powershell
cd backend
make migrate-install
make migrate-up SERVICE=user-service
make migrate-down SERVICE=user-service
```

创建新迁移：

```powershell
make migrate-create SERVICE=im-service NAME=add_message_attachments
```

## 快速开始

### 环境要求

> 以下命令在 Windows 和 Linux 环境下均可使用。Windows 需安装 Docker Desktop，Linux 安装 Docker Engine 配合 Compose 插件即可。

| 工具    | 最低版本 | 用途                                 |
| ------- | -------- | ------------------------------------ |
| Docker  | 20.10+   | PostgreSQL、Redis、MinIO 容器运行    |
| Go      | 1.22+    | 后端服务                             |
| Node.js | 18.x+    | 前端构建与开发服务器                 |
| npm     | 9.x+     | 前端包管理                           |

一条命令检查所有必需工具：

```bash
docker --version && go version && node --version && npm --version
```

### 启动完整技术栈

```bash
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
```

### 启动前端

```bash
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

浏览器访问 `http://localhost:3000`，验收账号见下方。

### 逐个调试服务（可选）

先启动基础设施，再分别在独立终端运行各服务：

```bash
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

```bash
cd backend && go run ./cmd/user-service
cd backend && go run ./cmd/im-service
cd backend && go run ./cmd/file-service
cd backend && go run ./cmd/search-service
cd backend && go run ./cmd/workspace-service
cd backend && go run ./cmd/gateway
```

## 默认验收账号

| 角色 | 用户名 | 密码 |
| --- | --- | --- |
| 管理员 | `admin` | `admin123` |
| 员工 | `emma.chen` | `workpal123` |
| 员工 | `liam.wang` | `workpal123` |
| 员工 | `sofia.zhao` | `workpal123` |

## 网关与服务发现自检

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/gateway/routes
Invoke-RestMethod http://localhost:8080/gateway/services
```

你应该能看到：

- 网关存活结果
- 网关及下游服务就绪结果
- 显式路由目录
- 带 `discovery_mode`、实例信息、超时、重试和熔断元数据的服务目录

## 当前后端链路的关键变化

- 服务启动后会把实例注册到 Redis，Gateway 会优先基于注册表发现下游服务
- IM 服务使用 Redis Pub/Sub 做跨实例消息广播
- IM 写消息、编辑消息、撤回消息时，会把待发布事件写入 `message_outbox`
- 后台 worker 再把 outbox 事件发布到 Redis Streams（消费者组模式），Search Service 订阅后更新 Bleve 索引
- 消息发送携带幂等令牌（`idempotency_key`），后端按 `(conv_id, sender_id, key)` 去重，有效期 5 分钟
- 消息冷热分离：近期消息写扩散至成员收件箱，历史消息走读扩散（按会话 + 时间范围查询）
- Gateway 层对搜索路由启用熔断，连续失败 5 次后半开 30 秒，搜索不可用时消息收发不受影响
- 各服务统一接入 OpenTelemetry，Gateway 注入 `trace_id` 并通过 Header 传递至下游
- 敏感操作（删除任务/消息/群成员、下载文件）写入 `audit_logs` 表
- Workspace Service 支持"创建带提醒的任务"Saga 编排：正向写任务 → 写日程提醒，失败则调用补偿接口回滚
- 各服务暴露 `/metrics`（Prometheus 格式，含 QPS/错误率/P99 延迟/goroutine 数）
- Docker Compose 内嵌 Prometheus + Grafana + Jaeger，启动后即可使用

## 关于当前前端数据形态

- 任务、日程、文件、沟通、通讯录都走后端
- 消息支持编辑与撤回，编辑采用行内模式
- 文件模块主列表不再混入前端写死的演示文档
- 登录页仍保留预置账号提示，方便验收和调试

## 测试命令

### 后端

```powershell
cd backend
go vet ./...
go test -race ./...
```

### 前端

```powershell
cd frontend
npm run lint
npm test
npm run build
```

### 持续集成

GitHub Actions 会在推送到 `main` 和向 `main` 发起 Pull Request 时运行。流水线包括：

- **后端：** 构建、`go vet`、`golangci-lint`、带 race 检测的 Go 测试
- **前端：** TypeScript 类型检查、ESLint、Vitest 组件测试、生产构建
- **端到端：** 启动 Compose 服务，运行 Playwright API 烟雾测试（健康检查、登录、聊天）
- **Compose：** Docker Compose 配置校验

### 性能压测

```powershell
cd testing/performance
npm install
k6 run 群发消息压测.js
```

压测脚本模拟 2000 人大群同时发消息场景，输出 P50/P99 延迟、吞吐量等关键指标。

### 端到端烟测

确保前后端都已启动，再执行：

```powershell
cd testing/e2e
npm install
npx playwright install chromium
node playwright.mjs
```

## 相关文档

- [backend/README.md](backend/README.md)
- [frontend/README.md](frontend/README.md)
- [学习手册](docs/学习手册.md)
- [测试手册](docs/测试手册.md)
- [技术选型文档](docs/技术选型文档.md)
- [架构设计](docs/架构设计.md)
- [微服务架构核心逻辑说明](docs/微服务架构核心逻辑说明.md)
- [可观测性设计](docs/可观测性设计.md)
- [消息推送性能优化报告](docs/消息推送性能优化报告.md)
- [生产部署指南](docs/生产部署指南.md)
- [韧性设计与降级策略](docs/韧性设计与降级策略.md)
- [安全设计](docs/安全设计.md)
- [分布式事务设计](docs/分布式事务设计.md)
- [工作上下文智能检索与洞察平台设计](docs/工作上下文智能检索与洞察平台设计.md)
- [分布式任务调度系统设计](docs/分布式任务调度系统设计.md)

## 生产部署注意事项

- 生产环境必须通过环境变量或 Kubernetes Secret 覆盖 `SERVER_JWTSECRET`、`SERVER_INTERNALTOKEN`、`DATABASE_PASSWORD`、`FILE_MINIO_ACCESSKEY`、`FILE_MINIO_SECRETKEY` 等敏感配置。
- 不要提交本地 `.env`、`backend/configs/config.yaml`、数据库密码、MinIO 密钥或真实令牌；仓库 `.gitignore` 已覆盖这些本地文件。
- `backend/configs/生产配置模板.yaml` 仅作为模板，不应直接携带真实密钥。
- 首次上线前请替换所有默认开发账号密码，并检查 Grafana、MinIO、数据库等控制台默认口令。
- 敏感操作会写入 `audit_logs` 表，建议生产环境定期归档并限制查询权限。

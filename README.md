# WorkPal

[English](README-en.md) | 中文说明

WorkPal 是一个基于 Go 和 React 的办公协作平台。当前版本已经按真实微服务形态运行：前端只访问 API Gateway，后端各领域服务拥有各自的运行边界和存储边界。

## 项目能力

- 预置验收账号：`admin`、`emma.chen`、`liam.wang`、`sofia.zhao`
- `English / 简体中文` 双语界面
- 深浅色主题、消息提示音开关、界面密度切换
- 总览、沟通、任务、日程、文件、通讯录六大模块
- 私聊、群聊、群公告、群文件
- 消息编辑、消息撤回、行内编辑
- 消息发送幂等保护（幂等令牌去重）
- 后端驱动的任务、日程、文件、通讯录搜索
- 网关治理（路由目录、限流、重试、熔断、健康检查）
- Redis 服务注册发现、IM 跨实例 Pub/Sub 广播
- 基于 outbox + Redis Streams 的搜索索引链路
- 消息冷热分离存储（近期写扩散 / 历史读扩散）
- OpenTelemetry 全链路追踪 + 结构化 JSON 日志
- Prometheus `/metrics` 端点 + Grafana 看板 + Jaeger 追踪
- Kubernetes 部署清单（Deployment + Service + 探针 + 滚动更新）
- 六个微服务的版本化数据库迁移体系（含审计日志、幂等表、Saga 编排表）
- 审计日志记录（用户操作、IP、时间戳）
- 搜索服务熔断降级（连续失败后半开探测）
- 任务创建 Saga 编排（带补偿回滚的事务状态机）

## 技术栈

- **后端：** Go 1.22、Gin、GORM、PostgreSQL 16、Redis 7、Redis Streams、Bleve、golang-migrate
- **前端：** React 18、Vite 5、TypeScript 5.4、Zustand 4.5
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
| User Service | `http://localhost:8081` | 认证、用户、部门、员工档案 |
| IM Service | `http://localhost:8082` | 会话、消息、群公告、WebSocket |
| File Service | `http://localhost:8083` | 个人文件、群文件 |
| Search Service | `http://localhost:8084` | 消息搜索与索引 |
| Workspace Service | `http://localhost:8085` | 任务、日程 |
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
| Gateway | 无状态 | 统一入口、路由目录、服务目录、注册发现回退、限流、重试、熔断、搜索降级、健康检查、trace 注入 |
| User Service | `workpal_user` | 登录、用户、部门、员工档案、开发种子数据、审计日志 |
| IM Service | `workpal_im` | 私聊、群聊、群公告、消息、消息编辑/撤回、幂等去重、消息冷热分离、WebSocket、消息 outbox、审计日志 |
| File Service | `workpal_file` | 文件元数据、上传、分享、删除、审计日志 |
| Search Service | Bleve + Redis Streams | 消息索引与搜索、熔断降级 |
| Workspace Service | `workpal_workspace` | 任务、日程、Saga 事务编排、审计日志 |

## 数据库迁移

每个服务在 `backend/migrations/<service>/` 下配有版本化 SQL 迁移：

| 服务 | 迁移 | 表 |
| --- | --- | --- |
| user-service | `001_init` | `users`, `departments`, `employees` |
| im-service | `001_init` | `conversations`, `conversation_members`, `messages`, `message_outbox`, `message_reads` |
| im-service | `002_message_idempotency_and_tier` | 消息幂等令牌、消息冷热分区索引 |
| im-service | `003_audit_logs` | `audit_logs` |
| file-service | `001_init` | `files` |
| file-service | `002_audit_logs` | `audit_logs` |
| workspace-service | `001_init` | `tasks`, `schedule_events` |
| workspace-service | `002_audit_logs` | `audit_logs` |
| workspace-service | `003_task_sagas` | `task_sagas`（Saga 事务状态表） |

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

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
- 后端驱动的任务、日程、文件、通讯录搜索
- 网关治理、Redis 服务注册发现、IM 跨实例广播、基于 outbox 的 Redis Streams 搜索索引链路
- 四个领域服务的版本化数据库迁移体系

## 技术栈

- **后端：** Go 1.22、Gin、GORM、PostgreSQL 16、Redis 7、Redis Streams、Bleve、golang-migrate
- **前端：** React 18、Vite 5、TypeScript 5.4、Zustand 4.5
- **文件存储：** MinIO，带本地文件回退
- **实时通信：** WebSocket，IM 服务通过 Redis Pub/Sub 做多实例消息扇出

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

## 微服务边界

| 服务 | 存储边界 | 主要职责 |
| --- | --- | --- |
| Gateway | 无状态 | 统一入口、路由目录、服务目录、注册发现回退、限流、重试、熔断、健康检查 |
| User Service | `workpal_user` | 登录、用户、部门、员工档案、开发种子数据 |
| IM Service | `workpal_im` | 私聊、群聊、群公告、消息、消息编辑/撤回、WebSocket、消息 outbox |
| File Service | `workpal_file` | 文件元数据、上传、分享、删除 |
| Search Service | Bleve + Redis Streams | 消息索引与搜索 |
| Workspace Service | `workpal_workspace` | 任务、日程 |

## 数据库迁移

每个服务在 `backend/migrations/<service>/` 下配有版本化 SQL 迁移：

| 服务 | 迁移 | 表 |
| --- | --- | --- |
| user-service | `001_init` | `users`, `departments`, `employees` |
| im-service | `001_init` | `conversations`, `conversation_members`, `messages`, `message_outbox`, `message_reads` |
| file-service | `001_init` | `files` |
| workspace-service | `001_init` | `tasks`, `schedule_events` |

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
- 后台 worker 再把 outbox 事件发布到 Redis Streams，Search Service 订阅后更新 Bleve 索引

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
- [测试手册](docs/测试手册.md)
- [技术选型文档](docs/技术选型文档.md)
- [架构设计](docs/架构设计.md)
- [学习手册](docs/学习手册.md)

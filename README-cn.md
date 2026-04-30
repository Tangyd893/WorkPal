# WorkPal

中文说明 | [English](README.md)

WorkPal 是一个基于 Go 和 React 的办公协作项目。当前版本已经按真实微服务形态运行：前端只访问 API Gateway，后端各领域服务拥有各自的运行边界和存储边界。

## 项目能力

- 预置验收账号：`admin`、`emma.chen`、`liam.wang`、`sofia.zhao`
- `English / 简体中文` 双语界面
- 深浅色主题、消息提示音开关、界面密度切换
- 总览、沟通、任务、日程、文件、通讯录六大模块
- 私聊、群聊、群公告、群文件
- 后端驱动的任务、日程、文件、通讯录搜索
- 网关治理、Redis 服务注册发现、IM 跨实例广播、基于 outbox 的 Redis Streams 搜索索引链路

## 技术栈

- 后端：Go、Gin、GORM、PostgreSQL、Redis、Redis Streams、Bleve
- 前端：React 18、Vite、TypeScript、Zustand
- 文件存储：MinIO，带本地文件回退
- 实时通信：WebSocket，IM 服务通过 Redis Pub/Sub 做多实例消息扇出

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
| IM Service | `workpal_im` | 私聊、群聊、群公告、消息、WebSocket、消息 outbox |
| File Service | `workpal_file` | 文件元数据、上传、分享、删除 |
| Search Service | Bleve + Redis Streams | 消息索引与搜索 |
| Workspace Service | `workpal_workspace` | 任务、日程 |

## 快速开始

### 1. 先确认 Docker 正在运行

```powershell
docker version
```

只有输出同时包含 `Client` 和 `Server` 时，再继续下面步骤。

### 2. 使用 Docker Compose 启动整套服务

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

预期结果：

- `postgres`、`redis`、`minio` 为 `Up` 或 `healthy`
- `gateway`、`user-service`、`im-service`、`file-service`、`search-service`、`workspace-service` 为 `Up`

Compose 会等待 Redis 健康后再启动开启注册发现的后端服务，这样 `/gateway/services` 可以优先看到注册发现实例，而不是一启动就退回静态地址。

后端服务会自动确保自己拥有的数据存储存在：

- `workpal_user`
- `workpal_im`
- `workpal_file`
- `workpal_workspace`

### 3. 如果你要逐个调试服务

先只启动基础设施：

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

然后在多个终端分别运行：

```powershell
cd backend
go run ./cmd/user-service
```

```powershell
cd backend
go run ./cmd/im-service
```

```powershell
cd backend
go run ./cmd/file-service
```

```powershell
cd backend
go run ./cmd/search-service
```

```powershell
cd backend
go run ./cmd/workspace-service
```

```powershell
cd backend
go run ./cmd/gateway
```

### 4. 启动前端

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

浏览器访问：`http://localhost:3000`

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

这意味着搜索索引链路不再是“请求成功后顺手发一下事件”，而是具备了更稳定的最终一致性补偿路径。

## 关于当前前端数据形态

- 任务、日程、文件、通讯录都走后端
- 文件模块主列表不再混入前端写死的演示文档
- 登录页仍保留预置账号提示，方便验收和调试

## 测试命令

### 后端

```powershell
cd backend
go test ./...
```

### 前端

```powershell
cd frontend
npm test
npm run build
```

### 端到端烟测

确保前后端都已启动，再执行：

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

## 相关文档

- [backend/README.md](backend/README.md)
- [frontend/README.md](frontend/README.md)
- [测试手册](docs/测试手册.md)
- [技术选型文档](docs/技术选型文档.md)
- [架构设计](docs/架构设计.md)
- [学习手册](docs/学习手册.md)

# WorkPal

中文说明 | [English](README.md)

WorkPal 是一个基于 Go 与 React 的办公协作平台示例项目。当前版本的后端已经彻底收敛为微服务形态，不再保留单体兼容入口；前端通过 API Gateway 访问后端，各领域服务拥有各自的数据边界。

## 项目能力

- 预置管理员与员工验收账号
- `English / 简体中文` 双语界面
- 浅色 / 深色主题、消息提示音开关、密度切换
- 总览、沟通、任务、日程、文件、通讯录六大板块
- 私聊、群聊、群公告、群文件
- 后端驱动的通讯录搜索、任务、日程、个人文件
- API Gateway、领域微服务、Redis Streams、Bleve 搜索链路

## 技术栈

- 后端：Go、Gin、GORM、PostgreSQL、Redis、Redis Streams、Bleve
- 前端：React 18、Vite、TypeScript、Zustand
- 文件存储：MinIO，带本地文件回退
- 实时通信：WebSocket

## 环境要求

- Go `1.22+`
- Node.js `18+`
- npm
- Docker Desktop 或 Docker Engine

开始前请先确认 Docker 正在运行：

```powershell
docker version
```

只有当输出同时包含 `Client` 和 `Server` 时，才继续后续步骤。

## 默认端口

| 组件 | 地址 | 说明 |
| --- | --- | --- |
| 前端 | `http://localhost:3000` | Vite 开发服务器 |
| API Gateway | `http://localhost:8080` | 前端唯一后端入口 |
| User Service | `http://localhost:8081` | 认证、用户、部门、员工档案 |
| IM Service | `http://localhost:8082` | 会话、消息、WebSocket |
| File Service | `http://localhost:8083` | 个人文件、群文件 |
| Search Service | `http://localhost:8084` | 消息搜索与索引 |
| Workspace Service | `http://localhost:8085` | 任务、日程 |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |

## 后端微服务边界

| 服务 | 存储边界 | 关键职责 |
| --- | --- | --- |
| Gateway | 无状态 | 统一入口、路由目录、服务目录、限流、重试、熔断、健康检查 |
| User Service | `workpal_user` | 登录、用户、部门、员工档案、开发种子数据 |
| IM Service | `workpal_im` | 私聊、群聊、消息、群公告、WebSocket |
| File Service | `workpal_file` | 文件元数据、上传、分享、删除 |
| Search Service | Bleve + Redis Streams | 消息搜索与索引消费 |
| Workspace Service | `workpal_workspace` | 任务、日程 |

## Gateway 学习亮点

Gateway 现在不只是一个转发入口，而是微服务入口层的学习样本。它已经具备：

- `/gateway/routes`：查看显式路由目录
- `/gateway/services`：查看下游服务目录与治理状态
- `/health/live`：存活检查
- `/health/ready`：就绪检查
- `/health`：聚合健康检查
- 请求 ID 注入
- 限流
- 服务级超时
- 只对幂等读请求启用重试
- 熔断器

如果你熟悉 Spring Cloud Alibaba，可以把它理解为：

- Gateway：入口层
- 静态配置驱动的服务目录：轻量版 Nacos 视角
- 限流 / 重试 / 熔断：轻量版 Sentinel 视角

## 快速开始

### 1. 用 Docker Compose 启动整套服务

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

预期结果：

- `postgres`、`redis`、`minio` 为 `Up` 或 `healthy`
- `gateway`、`user-service`、`im-service`、`file-service`、`search-service`、`workspace-service` 为 `Up`

首次启动时，后端服务会自动创建自己的数据库：

- `workpal_user`
- `workpal_im`
- `workpal_file`
- `workpal_workspace`

### 2. 如果要逐个调试服务

先只启动基础设施：

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

然后在多个终端中分别启动：

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

### 3. 启动前端

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

打开：

```text
http://localhost:3000
```

## 默认验收账号

默认开发模式下，User Service 会自动确保以下账号存在：

| 角色 | 用户名 | 密码 |
| --- | --- | --- |
| 管理员 | `admin` | `admin123` |
| 员工 | `emma.chen` | `workpal123` |
| 员工 | `liam.wang` | `workpal123` |
| 员工 | `sofia.zhao` | `workpal123` |

同时还会补齐部门与员工档案数据，因此通讯录筛选和搜索可以直接使用。

## 快速自检

### Gateway 管理面与健康检查

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/gateway/routes
Invoke-RestMethod http://localhost:8080/gateway/services
```

预期：

- `live` 返回网关存活
- `ready` 返回网关和下游服务可接流量
- `health` 返回聚合健康状态
- `routes` 返回显式路由目录
- `services` 返回 5 个下游服务及其超时、重试、熔断信息

### 登录接口

```powershell
$body = @{
  username = "admin"
  password = "admin123"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

预期：

- `code` 为 `0`
- `data.token` 存在
- 响应头中可看到 `X-Upstream-Service: user-service`

### 前端登录

打开 `http://localhost:3000`，使用 `admin / admin123` 登录，预期会跳转到 `/workspace/overview`。

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
- [docs/测试手册.md](docs/测试手册.md)
- [docs/技术选型文档.md](docs/技术选型文档.md)
- [docs/架构设计.md](docs/架构设计.md)
- [docs/学习手册.md](docs/学习手册.md)

## License

MIT

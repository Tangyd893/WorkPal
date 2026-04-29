# WorkPal

中文说明 | [English](README.md)

WorkPal 是一个基于 Go 微服务 + React 的办公协作演示项目。当前版本已经不只是一个聊天外壳，包含以下能力：

- 用于验收测试的内置管理员和员工账号
- 双语界面：`English / 简体中文`
- 浅色/深色主题、消息提示音开关、紧凑密度开关
- 总览、沟通、任务、日程、文件和通讯录模块
- 后端数据库内置部门和员工种子数据
- 私聊、群聊、群公告和群文件

本文档基于当前仓库代码编写，用于按步骤完成本地启动和调试。

## 技术栈

- 后端：Go 微服务、Gin、GORM、PostgreSQL、Redis Streams、Bleve
- 前端：React、Vite、Zustand
- 文件存储：默认本地存储，同时支持 MinIO
- 实时通信：通过 IM 服务和 API Gateway 转发 WebSocket

## 环境要求

- Go `1.22+`
- Node.js `18+`
- npm
- Docker Desktop 或 Docker Engine

开始前请先确认 Docker 正在运行：

```powershell
docker version
```

只有当输出同时包含 `Client` 和 `Server` 时，再继续后续步骤。

## 默认端口

| 服务 | URL | 说明 |
|---|---|---|
| 前端 | `http://localhost:3000` | Vite 开发服务器 |
| API Gateway | `http://localhost:8080` | HTTP 和 WebSocket 的前端统一入口 |
| User Service | `http://localhost:8081` | 认证、用户、部门 |
| IM Service | `http://localhost:8082` | 会话、消息、WebSocket |
| File Service | `http://localhost:8083` | 个人文件和群文件 |
| Search Service | `http://localhost:8084` | Bleve 搜索 API 和消息索引消费者 |
| Workspace Service | `http://localhost:8085` | 后端持久化任务和日程 |
| 健康检查 | `http://localhost:8080/health` | 网关健康检查端点 |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO 控制台 | `http://localhost:9001` | `workpal / workpal123456` |

## 快速开始

### 1. 启动 Docker Compose 服务栈

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

预期结果：

- `postgres` 状态为 `Up` 或 `healthy`
- `redis` 状态为 `Up`
- `minio` 状态为 `Up`
- `gateway`、`user-service`、`im-service`、`file-service`、`search-service` 和 `workspace-service` 状态为 `Up`

如果想从源码手动运行 Go 服务，可以只先启动基础设施：

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

### 2. 启动后端微服务

如果已经启动完整 Docker Compose 服务栈，可以跳过本节。否则分别打开多个终端启动以下服务：

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

前端只需要访问 Gateway。Gateway 会把请求转发到对应领域服务：

| Gateway 路径 | 上游服务 |
|---|---|
| `/api/v1/auth/*`、`/api/v1/users*`、`/api/v1/departments*` | User Service |
| `/api/v1/conversations*`、`/api/v1/messages*`、`/ws` | IM Service |
| `/api/v1/files*`、`/api/v1/conversations/:id/files` | File Service |
| `/api/v1/search*` | Search Service |
| `/api/v1/tasks*`、`/api/v1/schedule*` | Workspace Service |

为了兼容快速本地调试，保留了一体化后端启动方式：

```powershell
cd backend
go run ./cmd/server
```

微服务模式下的重要启动行为：

1. 每个服务启动时迁移自己负责的数据表。
2. User Service 在非 `release` 模式下会自动确保部门、员工和验收账号等种子数据存在。
3. IM Service 将消息写入 PostgreSQL，并向 Redis Streams 发布消息索引事件。
4. Search Service 消费 Redis Streams 事件，并更新 Bleve 索引。
5. Workspace Service 持久化过去只存在于前端本地演示状态的任务和日程。
6. 除非需要覆盖样例配置，否则不需要创建 `backend/configs/config.yaml`。

后端配置查找顺序：

1. `CONFIG_PATH`
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

快速验证：

```powershell
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
Invoke-WebRequest http://localhost:8080/ -UseBasicParsing
```

预期结果：

- `/health` 返回 HTTP `200`
- `/` 返回类似下面的 JSON：

```json
{"name":"WorkPal","status":"running","version":"0.2.0"}
```

### 3. 启动前端

打开另一个终端：

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

然后打开：

```text
http://localhost:3000
```

前端代理规则：

- `/api/*` -> `http://localhost:8080` -> API Gateway -> 目标服务
- `/ws` -> `ws://localhost:8080` -> API Gateway -> IM Service

## 微服务消息流

聊天消息使用 HTTP 保证持久化，使用 Redis Streams 做跨服务索引解耦：

```text
Frontend
  -> API Gateway
  -> IM Service
  -> PostgreSQL
  -> Redis Streams: message.upserted / message.deleted
  -> Search Service
  -> Bleve index
```

这样即使搜索索引服务短暂不可用，也不会影响消息发送。PostgreSQL 仍是消息事实源，Redis Streams 用来解耦 IM 写入和搜索索引更新。

## 验收账号

后端以默认开发模式启动时，会自动确保以下账号存在：

| 角色 | 用户名 | 密码 | 建议用途 |
|---|---|---|---|
| 管理员 | `admin` | `admin123` | 完整验收和设置检查 |
| 员工 | `emma.chen` | `workpal123` | 运营和协作流程 |
| 员工 | `liam.wang` | `workpal123` | 工程和群聊测试 |
| 员工 | `sofia.zhao` | `workpal123` | 设计和发布就绪测试 |

种子组织数据还包含部门和员工档案，因此通讯录搜索和部门筛选可以开箱即用。

如果想先通过 API 验证登录：

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

预期结果：

- `code` 为 `0`
- `data.token` 存在

## 推荐验收路径

在前端登录后，建议按以下顺序检查：

1. `Overview / 总览`
   - 确认总览页面可以加载
   - 点击指标卡片或模块按钮，确认能跳转到对应模块

2. `Preferences / 偏好设置`
   - 切换 `English / 简体中文`
   - 切换浅色和深色主题
   - 开关消息提示音
   - 切换舒适和紧凑密度

3. `Directory / 通讯录`
   - 确认种子用户可见
   - 使用部门筛选
   - 按职位、电话或部门搜索，而不仅是按用户名搜索
   - 示例：筛选 `Engineering`，搜索 `Platform Engineer`

4. `Chat / 沟通`
   - 与 `emma.chen` 创建私聊
   - 发送一条消息
   - 创建包含 `emma.chen` 和 `liam.wang` 的群聊
   - 发送一条群消息
   - 更新群公告
   - 上传一个群文件

5. `Tasks / 任务`
   - 创建任务
   - 在不同列之间移动任务
   - 分享任务
   - 删除任务

6. `Schedule / 日程`
   - 创建日程
   - 分享日程
   - 删除日程

7. `Files / 文件`
   - 上传个人文件
   - 打开文件
   - 分享文件
   - 删除文件

## 后端支撑数据与前端本地演示状态

这部分对调试预期非常重要。

### 当前由后端支撑

- 登录
- 当前用户
- 用户列表和部门列表
- 通讯录搜索和部门筛选
- 私聊和群聊
- 消息发送和消息搜索
- WebSocket 连接
- 群公告
- 群文件
- 个人文件上传、列表、分享、删除
- 任务创建、状态更新、分享、删除
- 日程创建、分享、删除

### 当前仍是前端本地演示状态

- 总览摘要组合
- 文件模块中的种子知识卡片

这意味着：

- 任务和日程现在通过 Workspace Service 持久化
- 通过文件服务上传的文件是真实后端数据
- 总览模块会结合当前前端状态和后端加载的人员数据展示

## 验证命令

### 后端和前端测试

```powershell
cd backend
go test ./...
make build-services

cd ..\frontend
npm test
npm run build
```

### 端到端冒烟测试

确保后端和前端已经运行，然后执行：

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

Playwright 冒烟测试覆盖：

- `/health`
- `/metrics`
- 种子账号登录 API
- 私聊和群聊 API 流程
- 群公告和群文件 API 流程
- 前端登录
- 总览跳转动作
- 通讯录筛选和搜索
- 任务创建
- 日程创建
- 文件上传
- 私聊创建和消息发送

## 停止服务

在前端和后端终端中使用 `Ctrl + C` 停止服务。

在仓库根目录停止 Docker 依赖：

```powershell
docker compose -f docker/docker-compose.yaml down
```

## 相关文档

- [README.md](README.md)
- [frontend/README.md](frontend/README.md)
- [docs/acceptance-testing.md](docs/acceptance-testing.md)
- [docs/项目技术特点学习笔记.md](docs/项目技术特点学习笔记.md)

## 许可证

MIT

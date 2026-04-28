# WorkPal

WorkPal 是一个基于 Go + React 的协作/即时通讯学习项目，包含登录鉴权、会话、消息、文件、全文搜索和基础监控。

这份 README 的启动步骤已按仓库当前代码实际验证过一次：在 **2026 年 4 月 28 日**，我实际启动了 Docker 依赖、后端、前端，并完成了“注册 -> 登录 -> 获取当前用户”的联调烟测。

## 运行环境

- Go 1.22+
- Node.js 18+
- npm
- Docker Desktop / Docker Engine

建议先确认 Docker 已经真正启动完成：

```powershell
docker version
```

如果输出里只有 `Client` 没有 `Server`，说明 Docker 还没准备好。

## 默认端口

| 服务 | 地址 | 说明 |
|---|---|---|
| 前端 | `http://localhost:3000` | Vite dev server |
| 后端 | `http://localhost:8080` | Gin API |
| 健康检查 | `http://localhost:8080/health` | 检查 PostgreSQL/Redis 连通性 |
| PostgreSQL | `localhost:5432` | 用户/密码：`workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO Console | `http://localhost:9001` | 用户/密码：`workpal / workpal123456` |

## 快速开始

下面默认使用 **Windows PowerShell**。如果你在 macOS/Linux 上运行，把 `Copy-Item` 换成 `cp` 即可。

### 1. 启动基础依赖

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

你应该能看到 `postgres`、`redis`、`minio` 三个服务。

如果这里失败，先不要继续启动后端；先解决 Docker 本身的问题。

### 2. 准备后端配置

后端会按下面的顺序找配置文件：

1. `CONFIG_PATH` 指定的文件
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

也就是说，**不复制配置文件也能直接跑**，因为仓库里已经有 `config.example.yaml`，并且我就是按这个路径实际启动成功的。

如果你想改本地配置，先复制一份：

```powershell
Copy-Item backend\configs\config.example.yaml backend\configs\config.yaml
```

常见需要改的地方：

- `server.jwtSecret`
- `database.*`
- `redis.*`
- `file.storeType`
- `search.bleve.indexPath`

默认样例配置使用本地 Docker 启动出来的 PostgreSQL / Redis / MinIO。

### 3. 启动后端

打开一个终端窗口：

```powershell
cd backend
go run ./cmd/server
```

保持这个终端不要关。

后端启动后，用另一个终端验证：

```powershell
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
Invoke-WebRequest http://localhost:8080/ -UseBasicParsing
```

预期：

- `/health` 返回 `200`
- `/` 返回类似下面的 JSON：

```json
{"name":"WorkPal","status":"running","version":"0.2.0"}
```

### 4. 启动前端

再开一个终端窗口：

```powershell
cd frontend
npm ci
npm run dev
```

启动后打开：

```text
http://localhost:3000
```

前端通过 Vite 代理把下面两类请求转发到后端：

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

所以本地开发时不需要单独处理跨域。

### 5. 默认登录账号

在默认开发配置下，后端启动时会自动确保一个默认管理员账号存在：

- 用户名：`admin`
- 密码：`admin123`

这里的“默认开发配置”指的是：

- `server.mode` 不是 `release`
- 你按仓库默认的本地启动方式运行后端

你可以先直接验证这组账号能否登录：

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

成功时返回体里的 `code` 应该是 `0`，并带有 `token`。

### 6. 登录并开始调试

浏览器打开：

```text
http://localhost:3000
```

优先使用默认管理员账号登录：

- 用户名：`admin`
- 密码：`admin123`

登录成功后会进入聊天页。

### 7. 可选：额外创建一个普通测试账号

当前前端只有登录页，**没有注册页**。  
如果你需要第二个账号做会话、搜索或联调测试，可以手动创建：

```powershell
$suffix = Get-Date -Format 'MMddHHmmss'
$username = "debug$suffix"

$body = @{
  username = $username
  password = "pass123456"
  nickname = "Debug User"
  email = "$username@example.com"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

这个写法每次都会生成新的用户名和 email，适合重复调试。

## 一次性联调自检

如果你想快速确认“前端代理 + 后端 API + 数据库 + JWT”整条链路都通，可以在仓库根目录执行：

```powershell
$suffix = Get-Date -Format 'MMddHHmmss'
$username = "codex$suffix"
$password = "pass123456"
$base = "http://localhost:3000/api/v1"

$registerBody = @{
  username = $username
  password = $password
  nickname = "Smoke Test"
  email = "$username@example.com"
} | ConvertTo-Json

$register = Invoke-RestMethod `
  -Uri "$base/auth/register" `
  -Method Post `
  -ContentType "application/json" `
  -Body $registerBody

$loginBody = @{
  username = $username
  password = $password
} | ConvertTo-Json

$login = Invoke-RestMethod `
  -Uri "$base/auth/login" `
  -Method Post `
  -ContentType "application/json" `
  -Body $loginBody

$headers = @{
  Authorization = "Bearer $($login.data.token)"
}

Invoke-RestMethod -Uri "$base/users/me" -Headers $headers
```

这条命令链我已经实际跑过，返回正常。

## 常用开发命令

### 后端

```powershell
cd backend
go test ./...
go test -race ./...
go build ./cmd/server
```

### 前端

```powershell
cd frontend
npm test
npm run build
```

### E2E 脚本

要求前后端都已经启动：

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

说明：

- `npx playwright install chromium` 第一次运行时执行一次即可
- 这个脚本现在会自己创建临时测试账号，不再依赖不存在的默认管理员账号

补充：

- 当前数据库里的 `users.email` 是唯一索引
- 如果你重复跑注册示例，**不要省略 email，也不要反复用同一个 email**
- README 里的注册示例都已经改成“自动生成唯一 email”的写法
- 默认管理员账号用于本地开发/验收最方便；如果你把 `server.mode` 改成 `release`，就不要再依赖它自动创建

## 停止服务

### 停止前后端

在各自终端里按 `Ctrl + C`。

### 停止 Docker 依赖

回到仓库根目录：

```powershell
docker compose -f docker/docker-compose.yaml down
```

## 常见问题

### 1. `docker compose` 报错连不上 daemon

先确认 Docker Desktop 已经完全启动，再执行：

```powershell
docker version
```

只有出现 `Server` 段后，再继续后续步骤。

### 2. 后端起不来

优先检查：

- `5432` 是否被别的 PostgreSQL 占用
- `6379` 是否被别的 Redis 占用
- `backend/configs/config.yaml` 是否写错
- Docker 里的 `postgres` / `redis` 是否真的 healthy

可以先看：

```powershell
docker compose -f docker/docker-compose.yaml ps
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
```

### 3. 前端能打开，但登录失败

优先检查：

- 你是否先通过注册 API 创建了账号
- 后端是否还在运行
- 浏览器里请求是否打到了 `http://localhost:3000/api/v1/...`

### 4. 文件上传依赖 MinIO 吗

默认样例配置是 `minio` 模式。  
如果 MinIO 初始化失败，后端代码会回退到本地文件存储，但开发联调时还是建议把 MinIO 一起跑起来，避免行为和线上预期不一致。

## 项目结构

```text
WorkPal/
├─ backend/                 Go 后端
│  ├─ cmd/server/           程序入口
│  ├─ configs/              配置
│  ├─ internal/             业务代码
│  └─ pkg/                  公共工具
├─ frontend/                React + Vite 前端
│  ├─ src/api/              API 封装
│  ├─ src/components/       组件
│  ├─ src/hooks/            hooks / store
│  ├─ src/pages/            页面
│  └─ src/styles/           样式
├─ docker/                  Docker Compose
├─ testing/                 脚本和测试资源
└─ docs/                    设计与分析文档
```

## 补充说明

- 根 README 现在以“**从零启动并本地调试**”为目标编写。
- 如果你只看某个子项目，请同时参考：
  - [backend/README.md](backend/README.md)
  - [frontend/README.md](frontend/README.md)

## License

MIT

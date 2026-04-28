# WorkPal

WorkPal 是一个基于 Go + React 的办公协作平台示例项目。当前版本除了实时沟通，还补齐了更完整的工作台形态：

- 登录鉴权、用户资料、用户目录
- 私聊 / 群聊 / 消息搜索 / WebSocket 实时通信
- 工作台总览、任务看板、日程、文件与知识面板
- 中英文切换、深浅色主题、消息提示音、紧凑密度设置
- 本地开发环境的预置验收账号与员工账号

这份 README 的快速开始步骤，已按仓库当前代码重新校对过，目标就是让你可以直接照着启动并验收。

## 运行环境

- Go 1.22+
- Node.js 18+
- npm
- Docker Desktop / Docker Engine

建议先确认 Docker 已真正启动：

```powershell
docker version
```

只有在输出里同时看到 `Client` 和 `Server` 时，再继续后面的步骤。

## 默认端口

| 服务 | 地址 | 说明 |
|---|---|---|
| 前端 | `http://localhost:3000` | Vite dev server |
| 后端 | `http://localhost:8080` | Gin API |
| 健康检查 | `http://localhost:8080/health` | 检查 PostgreSQL / Redis |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |

## 快速开始

下面默认使用 Windows PowerShell。macOS / Linux 下把命令语法换成对应 shell 即可。

### 1. 启动依赖服务

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

预期看到 `postgres`、`redis`、`minio` 都是 `Up` / `healthy`。

### 2. 准备后端配置

后端按下面顺序读取配置：

1. `CONFIG_PATH` 指定的文件
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

因此，**即使你不复制 `config.yaml`，项目也能直接按样例配置跑起来**。

如果你想改本地配置，再复制一份：

```powershell
Copy-Item backend\configs\config.example.yaml backend\configs\config.yaml
```

### 3. 启动后端

新开一个终端：

```powershell
cd backend
go run ./cmd/server
```

后端启动后，项目会自动做两件事：

1. 执行 GORM `AutoMigrate`
2. 在 `server.mode != release` 时自动确保开发验收账号存在

先验证后端是否可用：

```powershell
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
Invoke-WebRequest http://localhost:8080/ -UseBasicParsing
```

预期：

- `/health` 返回 `200`
- `/` 返回类似：

```json
{"name":"WorkPal","status":"running","version":"0.2.0"}
```

### 4. 启动前端

再开一个终端：

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

然后打开：

```text
http://localhost:3000
```

前端本地代理规则：

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

### 5. 使用预置验收账号登录

只要后端是按默认开发模式启动，就会自动确保以下账号存在：

| 角色 | 用户名 | 密码 | 用途 |
|---|---|---|---|
| 管理员 | `admin` | `admin123` | 总览、设置、全局验收 |
| 员工 | `emma.chen` | `workpal123` | 运营协作测试 |
| 员工 | `liam.wang` | `workpal123` | 工程协作测试 |
| 员工 | `sofia.zhao` | `workpal123` | 设计 / 验收测试 |

你可以先直接验证其中一个账号是否能登录：

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

返回体里的 `code` 应为 `0`，并带有 `token`。

### 6. 建议的验收路径

登录前端后，建议按下面顺序验收：

1. `Overview / 总览`
   - 看工作台摘要卡片是否正常渲染
2. `Preferences / 偏好设置`
   - 切换 `English / 简体中文`
   - 切换浅色 / 深色主题
   - 切换消息提示音
   - 切换舒适 / 紧凑密度
3. `Directory / 通讯录`
   - 确认 `admin`、`emma.chen`、`liam.wang`、`sofia.zhao` 都可见
4. `Chat / 沟通`
   - 用“新建会话”直接选择员工账号创建私聊或群组
5. `Tasks / 任务`
   - 推进任务状态列，验证前端交互
6. `Schedule / 日程`
   - 查看会议与协作节奏
7. `Files / 文件`
   - 查看共享文档与状态标签

## 当前模块说明

### 后端实时数据

下面这些能力直接走后端 API：

- 登录
- 用户列表 / 当前用户
- 私聊 / 群聊
- 消息发送
- 消息搜索
- WebSocket 实时连接

### 前端预置协作演示数据

下面这些模块当前是为了补齐“办公协作平台”的产品完整性，在前端内置了演示数据与交互：

- 工作台总览摘要
- 任务看板
- 日程面板
- 文件与知识面板

这部分不会影响聊天和用户相关的真实联调，文档里也明确区分了。

## 一次性烟测

### 后端与前端单元 / 构建

```powershell
cd backend
go test ./...

cd ..\frontend
npm test
npm run build
```

### E2E 冒烟

要求前后端都已启动：

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

当前脚本会验证：

- `/health`
- `/metrics`
- `admin` 与 `emma.chen` 的登录 API
- 前端登录后的工作台导航
- 语言切换
- 通讯录中的预置成员
- 聊天模块里的新建会话弹窗

## 停止服务

### 停止前后端

在对应终端按 `Ctrl + C`。

### 停止 Docker 依赖

```powershell
docker compose -f docker/docker-compose.yaml down
```

## 相关文档

- [backend/README.md](backend/README.md)
- [frontend/README.md](frontend/README.md)
- [docs/acceptance-testing.md](docs/acceptance-testing.md)

## License

MIT

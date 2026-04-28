# WorkPal

WorkPal 是一个基于 Go + React 的办公协作演示项目。当前版本已经不只是一个聊天外壳，包含以下能力：

- 用于验收测试的内置管理员和员工账号
- 双语界面：`English / 简体中文`
- 浅色/深色主题、消息提示音开关、紧凑密度开关
- 总览、沟通、任务、日程、文件和通讯录模块
- 后端数据库内置部门和员工种子数据
- 私聊、群聊、群公告和群文件

本文档基于当前仓库代码编写，用于按步骤完成本地启动和调试。

## 技术栈

- 后端：Go、Gin、GORM、PostgreSQL、Redis、Bleve
- 前端：React、Vite、Zustand
- 文件存储：默认本地存储，同时支持 MinIO
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

只有当输出同时包含 `Client` 和 `Server` 时，再继续后续步骤。

## 默认端口

| 服务 | URL | 说明 |
|---|---|---|
| 前端 | `http://localhost:3000` | Vite 开发服务器 |
| 后端 | `http://localhost:8080` | Gin API |
| 健康检查 | `http://localhost:8080/health` | 检查 PostgreSQL 和 Redis |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | 默认无密码 |
| MinIO API | `http://localhost:9000` | 对象存储 |
| MinIO 控制台 | `http://localhost:9001` | `workpal / workpal123456` |

## 快速开始

### 1. 启动基础设施依赖

在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

预期结果：

- `postgres` 状态为 `Up` 或 `healthy`
- `redis` 状态为 `Up`
- `minio` 状态为 `Up`

### 2. 启动后端

打开一个新的终端：

```powershell
cd backend
go run ./cmd/server
```

重要启动行为：

1. 后端启动时会执行 `AutoMigrate`。
2. 在非 `release` 模式下，会自动确保部门、员工和验收账号等种子数据存在。
3. 除非需要覆盖样例配置，否则不需要创建 `backend/configs/config.yaml`。

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

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

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

### 当前仍是前端本地演示状态

- 总览摘要组合
- 任务看板条目
- 日程条目
- 文件模块中的种子知识卡片

这意味着：

- 任务和日程在 UI 中可用，但尚未持久化到后端
- 通过文件服务上传的文件是真实后端数据
- 总览模块会结合当前前端状态和后端加载的人员数据展示

## 验证命令

### 后端和前端测试

```powershell
cd backend
go test ./...

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

- [frontend/README.md](frontend/README.md)
- [docs/acceptance-testing.md](docs/acceptance-testing.md)
- [docs/项目技术特点学习笔记.md](docs/项目技术特点学习笔记.md)

## 许可证

MIT

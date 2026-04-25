# WorkPal

基于 Go + React 的企业协作平台（仿飞书/钉钉），适合作为学习项目。

## 项目进度总览

| Phase | 内容 | 状态 |
|-------|------|------|
| Phase 1 | 基础骨架（注册登录 + JWT + PostgreSQL + Redis） | ✅ 完成 |
| Phase 2 | IM 核心（WebSocket 私聊/群聊 + 已读回执 + 在线状态） | ✅ 完成 |
| Phase 3 | 生产级扩展（Docker + CI/CD + 单测 + Bleve 搜索 + MinIO） | ✅ 完成 |
| Phase 4 | 高级特性（云文档 Yjs · 音视频 WebRTC · 多因素认证） | 📋 待开发 |

---

## 项目结构

```
WorkPal/
├── backend/                 # Go 后端
│   ├── cmd/server/         # 程序入口
│   ├── configs/            # 配置文件
│   ├── deployments/        # Docker 部署配置
│   ├── internal/           # 私有业务代码
│   │   ├── common/        # 公共组件（errors/middleware/response/pagination）
│   │   ├── user/           # 用户模块（注册/登录/个人资料）
│   │   ├── im/            # 即时通讯（WebSocket/会话/消息）
│   │   ├── file/          # 文件存储（MinIO + 本地双模式）
│   │   └── search/        # 全文搜索（Bleve 索引）
│   ├── pkg/               # 公共工具（auth/JWT）
│   ├── Makefile
│   └── go.mod
│
├── frontend/              # React 18 前端
│   ├── src/
│   │   ├── api/          # Axios 封装 + 搜索 API
│   │   ├── components/   # 公共组件
│   │   ├── hooks/        # 自定义 Hook（WebSocket/Zustand stores）
│   │   ├── pages/        # 页面（Login/Register/Chat）
│   │   └── stores/       # Zustand 状态管理
│   ├── testing/          # E2E 测试（Playwright）
│   ├── package.json
│   └── vite.config.ts
│
├── AI-DEVELOPMENT.md     # AI 开发踩坑记录
└── README.md
```

---

## 技术栈

### 后端

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.22（测试通过） |
| HTTP 框架 | Gin |
| WebSocket | gorilla/websocket |
| 数据库 | PostgreSQL 16 |
| 缓存/消息队列 | Redis 7 + Redis Streams |
| 全文搜索 | Bleve（嵌入，无需额外部署） |
| 对象存储 | MinIO + 本地文件系统（双模式） |
| ORM | GORM |
| 配置 | Viper |
| 单元测试 | testify + miniredis v2 |

### 前端

| 类别 | 技术 |
|------|------|
| 框架 | React 18 + TypeScript |
| 构建 | Vite 5 |
| 路由 | React Router v6 |
| 状态管理 | Zustand |
| E2E 测试 | Playwright |

### DevOps / 基础设施

| 类别 | 技术 |
|------|------|
| 容器化 | Docker Compose（PostgreSQL + Redis + MinIO） |
| CI/CD | GitHub Actions（全流程自动化） |
| 监控 | Prometheus（`/metrics`） |

---

## 开发进度

### Phase 1 - 基础骨架 ✅

- [x] 用户注册/登录（bcrypt + JWT）
- [x] 个人资料管理
- [x] PostgreSQL + Redis 连接
- [x] 统一错误码体系（5位数 ABCDE 格式，HTTPStatus 内置）

### Phase 2 - IM 核心 ✅

- [x] WebSocket 长连接（gorilla/websocket）
- [x] 私聊/群聊会话管理
- [x] 消息发送/接收/历史记录
- [x] 消息已读回执（WebSocket TypeRead/TypeReadAll 广播）
- [x] 聊天页面全文搜索（头部搜索框 + Bleve 后端）
- [x] Redis 在线状态管理

### Phase 3 - 生产级扩展 ✅

- [x] Docker Compose 一键部署（PostgreSQL + Redis + MinIO）
- [x] GitHub Actions CI（go build + golangci-lint + `go test -race` + tsc + vitest + playwright）
- [x] 服务层单元测试（auth/message/conversation/presence + Hub 并发）
- [x] golangci-lint 代码质量检查（0 warnings）
- [x] Bleve 全文搜索（消息索引 + 搜索 API + 前端搜索框）
- [x] MinIO 对象存储（文件上传/下载/列表 + 本地文件双模式）
- [x] Prometheus 指标监控（`/metrics`）
- [x] 用户模糊搜索（PostgreSQL ILIKE）

### Phase 4 - 高级特性（待开发）

- [ ] 云文档协作编辑（Yjs）
- [ ] 音视频通话（WebRTC）
- [ ] 表情回复/线程消息
- [ ] 多因素认证

---

## 快速开始

### 环境要求

- Go 1.22+ | Node.js 18+ | Docker

### 1. 启动基础设施

```bash
cd WorkPal
docker compose -f docker/docker-compose.yaml up -d

# 验证服务
docker compose -f docker/docker-compose.yaml ps
```

### 2. 启动后端

```bash
cd backend
GOTOOLCHAIN=local go run cmd/server/main.go

# 服务启动在 http://localhost:8080
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev

# 前端启动在 http://localhost:3000
```

---

## 测试

### 后端单元测试

```bash
cd backend

# 全部测试（含数据竞争检测）
go test -race ./...

# 按模块运行
go test -race ./internal/im/service/...
go test -race ./internal/user/service/...
go test -race ./internal/im/ws/...
```

**当前覆盖率**：auth_svc · message_svc · conversation_svc · presence_svc · Hub（并发 race 测试）

### 前端类型检查

```bash
cd frontend
npx tsc --noEmit       # 类型检查
npm run build          # 生产构建
npm run test           # Vitest 单元测试
```

### E2E 测试（Playwright）

```bash
cd frontend
npx playwright install chromium
node testing/e2e/playwright.mjs
```

### CI

每次 PR/Push 自动运行：Go build · golangci-lint · `go test -race` · tsc · vitest · playwright e2e

---

## API 路由

### 认证

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录（返回 JWT） | ❌ |
| GET | `/api/v1/users/me` | 当前用户信息 | ✅ |
| PUT | `/api/v1/users/me` | 更新个人资料 | ✅ |

### 用户

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/users` | 用户列表（分页） | ✅ |
| GET | `/api/v1/users/search?q=` | 模糊搜索用户 | ✅ |

### IM 会话

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/conversations` | 创建私聊/群聊 | ✅ |
| GET | `/api/v1/conversations` | 会话列表 | ✅ |
| GET | `/api/v1/conversations/:id` | 会话详情 | ✅ |
| PUT | `/api/v1/conversations/:id` | 更新会话（群名） | ✅ |
| DELETE | `/api/v1/conversations/:id` | 解散会话（群主） | ✅ |
| POST | `/api/v1/conversations/:id/members` | 添加成员 | ✅ |
| DELETE | `/api/v1/conversations/:id/members/:uid` | 移除成员 | ✅ |

### 消息

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/conversations/:id/messages` | 历史消息（分页） | ✅ |
| POST | `/api/v1/conversations/:id/messages` | 发送消息（自动索引） | ✅ |
| PUT | `/api/v1/messages/:id` | 编辑消息 | ✅ |
| DELETE | `/api/v1/messages/:id` | 撤回消息 | ✅ |
| POST | `/api/v1/conversations/:id/read-all` | 全部已读 | ✅ |

### 文件

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/files/upload` | 上传文件（MinIO/本地） | ✅ |
| GET | `/api/v1/files` | 用户文件列表 | ✅ |
| GET | `/api/v1/files/:id` | 下载文件 | ✅ |
| GET | `/api/v1/conversations/:id/files` | 会话文件列表 | ✅ |

### 搜索

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/search/messages?q=&conv_id=&page=&page_size=` | 搜索消息（Bleve） | ✅ |

### 其他

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | ❌ |
| GET | `/metrics` | Prometheus 指标 | ❌ |
| WS | `/ws?token=` | WebSocket（JWT 放入 query 参数） | ✅ |

---

## Docker Compose 服务

```
PostgreSQL:    localhost:5432  (DB: workpal, User/Pass: workpal/workpal123)
Redis:         localhost:6379
MinIO API:     localhost:9000  (Console: localhost:9001, User/Pass: workpal/workpal123)
后端 API:      localhost:8080
前端:          localhost:3000
```

---

## 配置说明

配置文件：`backend/configs/config.yaml`

```yaml
server:
  port: 8080
  jwtSecret: "your-secret-key"       # 注意：YAML key 须为 camelCase
  jwtExpiryHours: 72

database:
  host: "localhost"
  port: 5432
  user: "workpal"
  password: "workpal123"
  dbname: "workpal"

redis:
  host: "localhost"
  port: 6379

file:
  store_type: "minio"   # minio / local
  minio:
    endpoint: "localhost:9000"
    access_key: "workpal"
    secret_key: "workpal123456"
    bucket: "workpal"
    use_ssl: false

search:
  engine: "bleve"
  bleve:
    index_path: "/tmp/workpal-search"
```

---

## Makefile 命令

```bash
make deps        # 下载 Go 依赖
make run         # 运行后端
make build       # 编译二进制
make docker-up   # 启动 Docker 服务
make docker-down # 停止 Docker 服务
```

---

## 编译验证

```bash
# 后端
cd backend
GOTOOLCHAIN=local go build ./cmd/server/ && echo "✅ 后端编译通过"

# 前端
cd frontend
npm run build && echo "✅ 前端编译通过"
```

---

## 踩坑记录（AI-DEVELOPMENT.md）

项目真实 AI 开发踩坑案例：

- **并发代码 race**：AI 生成的 Hub，`clients` map 读写没有锁，`go test -race` 爆红 → 手工加 RWMutex
- **viper YAML 映射**：YAML 用下划线（`jwt_secret`），Go struct 用 camelCase（`JWTSecret`），`mapstructure` 标签没加 → token 全部失效
- **Zustand persist**：AI 把整个 store persist 导致 hydration race → 去掉 persist 中间件
- **WS token 位置**：Gin 框架 Authorization header 被拦截，token 改放 query 参数 `?token=`
- **errcheck 忽略**：搜索索引失败被 `_ =` 吞掉，测试跑不过

详见 `AI-DEVELOPMENT.md`。

---

## License

MIT
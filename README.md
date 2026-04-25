# WorkPal

基于 Go + React 的企业协作平台（仿飞书/钉钉），适合作为学习项目。

## 项目结构

```
WorkPal/
├── backend/                 # Go 后端
│   ├── cmd/server/         # 程序入口
│   ├── configs/            # 配置文件
│   ├── deployments/        # Docker 部署配置
│   ├── internal/           # 私有业务代码
│   │   ├── common/        # 公共组件（错误/中间件/响应）
│   │   ├── user/          # 用户模块
│   │   ├── im/            # IM 即时通讯模块（WebSocket）
│   │   ├── file/          # 文件存储模块（MinIO + 本地）
│   │   ├── search/        # 搜索模块（Bleve 全文索引）
│   │   └── org/           # 组织架构模块
│   ├── pkg/                # 公共工具包（消息队列）
│   ├── Makefile
│   └── go.mod
│
├── frontend/               # React 18 前端
│   ├── src/
│   │   ├── api/           # Axios 封装
│   │   ├── components/    # 公共组件
│   │   ├── hooks/         # 自定义 Hook
│   │   ├── pages/         # 页面
│   │   └── styles/        # 全局样式
│   ├── package.json
│   └── vite.config.ts
│
└── docs/                   # 项目文档
```

## 技术栈

### 后端

- **语言**: Go 1.21+（Go 1.22.2 测试通过）
- **框架**: Gin (HTTP) + gorilla/websocket
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **消息队列**: Redis Streams
- **全文搜索**: Bleve 全文索引（v1.0.14）
- **文件存储**: MinIO + 本地文件系统双模式
- **ORM**: GORM
- **配置**: Viper
- **监控**: Prometheus client_golang

### 前端

- **框架**: React 18 + TypeScript
- **构建**: Vite 5
- **路由**: React Router v6
- **状态**: Zustand

## 开发阶段

### Phase 1 - 基础骨架 ✅
- 用户注册/登录（bcrypt + JWT）
- 个人资料管理
- PostgreSQL + Redis 连接

### Phase 2 - IM 核心 ✅
- WebSocket 长连接（gorilla/websocket）
- 私聊/群聊会话管理
- 消息发送/接收/历史
- 已读回执
- Redis 在线状态

### Phase 3 - 生产级扩展 ✅
- Redis Streams 消息队列
- Bleve 全文搜索（消息索引 + 搜索）
- MinIO 对象存储 + 本地文件双模式
- Prometheus 指标监控
- 用户模糊搜索（PostgreSQL ILIKE）
- Docker Compose 一键部署
- GitHub Actions CI（lint + test -race + 前端类型检查）
- 服务层单元测试覆盖（auth/message/conversation/presence + Hub 并发）
- golangci-lint 代码质量检查

### Phase 4 - 高级特性（待开发）
- 云文档协作编辑（OT 算法 / Yjs）
- 音视频通话（WebRTC）
- 表情回复/线程消息
- 多因素认证

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 16
- Redis 7
- Node.js 18+

### 1. 启动基础设施（Docker）

```bash
docker compose -f docker-compose.yaml up -d

# 验证
docker compose -f docker-compose.yaml ps
```

### 2. 启动后端

支持 `CONFIG_PATH` 环境变量指定配置文件路径（默认从 server 二进制所在目录查找）：

```bash
cd backend
GOTOOLCHAIN=local go build ./cmd/server/
./server

# 或直接运行（自动查找 configs/config.yaml）
GOTOOLCHAIN=local go run cmd/server/main.go

# 服务启动在 http://localhost:8080
```

### 3. 启动前端

```bash
cd frontend
pnpm install
pnpm dev

# 前端启动在 http://localhost:3000
```

## 测试

### 单元测试（Go）

```bash
cd backend
go test -race ./...           # 所有测试
go test -race ./internal/im/service/...   # IM 模块
go test -race ./internal/user/service/...  # 用户/认证模块
```

### E2E 测试（Playwright）

```bash
# 安装浏览器
cd frontend && npx playwright install chromium

# 运行测试
node testing/e2e/playwright.mjs
```

### 测试覆盖率

```bash
cd backend
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

CI 已在每次 PR/Push 时自动运行（见 `.github/workflows/ci.yml`）。

## API 路由

### 认证
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
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
| POST | `/api/v1/conversations` | 创建会话 | ✅ |
| GET | `/api/v1/conversations` | 会话列表 | ✅ |
| GET | `/api/v1/conversations/:id` | 会话详情 | ✅ |
| PUT | `/api/v1/conversations/:id` | 更新会话 | ✅ |
| DELETE | `/api/v1/conversations/:id` | 解散会话 | ✅ |
| POST | `/api/v1/conversations/:id/members` | 添加成员 | ✅ |
| DELETE | `/api/v1/conversations/:id/members/:uid` | 移除成员 | ✅ |

### 消息
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/conversations/:id/messages` | 历史消息 | ✅ |
| POST | `/api/v1/conversations/:id/messages` | 发送消息（自动索引） | ✅ |
| PUT | `/api/v1/messages/:id` | 编辑消息 | ✅ |
| DELETE | `/api/v1/messages/:id` | 撤回消息 | ✅ |
| POST | `/api/v1/conversations/:id/read-all` | 全部已读 | ✅ |

### 文件
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/files/upload` | 上传文件 | ✅ |
| GET | `/api/v1/files` | 用户文件列表 | ✅ |
| GET | `/api/v1/files/:id` | 下载文件 | ✅ |
| GET | `/api/v1/conversations/:id/files` | 会话文件 | ✅ |

### 搜索
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/search/messages?q=` | 搜索消息（Bleve） | ✅ |

### 其他
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | ❌ |
| GET | `/metrics` | Prometheus 指标 | ❌ |
| WSS | `/ws` | WebSocket | ✅ |

## Docker Compose 服务

```bash
# 启动所有服务
docker compose -f docker-compose.yaml up -d

# 服务端口
PostgreSQL:  localhost:5432
Redis:       localhost:6379
MinIO API:   localhost:9000
MinIO Console: localhost:9001
后端 API:    localhost:8080
前端:       localhost:3000 (开发模式)
```

## 配置说明

主要配置文件：`backend/configs/config.yaml`

```yaml
server:
  port: 8080
  jwtSecret: "your-secret-key"
  jwtExpiryHours: 72

database:
  host: "localhost"
  port: 5432

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

## 监控系统

Prometheus 指标端点：`GET /metrics`

当前指标：
- `http_requests_total` - HTTP 请求计数
- `http_request_duration_seconds` - 请求延迟
- `websocket_connections` - WebSocket 连接数
- `messages_total` - 消息收发计数

## Makefile 命令

```bash
make deps      # 下载 Go 依赖
make run       # 运行后端
make build     # 编译二进制
make docker-up   # 启动 Docker 服务
make docker-down # 停止 Docker 服务
```

## 编译验证

```bash
# 后端
cd backend
GOTOOLCHAIN=local go build ./cmd/server/  && echo "✅ 后端编译通过"

# 前端
cd frontend
npm run build  && echo "✅ 前端编译通过"
```

## License

MIT

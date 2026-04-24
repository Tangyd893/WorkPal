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
│   │   ├── file/          # 文件存储模块
│   │   ├── search/         # 搜索模块（PostgreSQL ILIKE）
│   │   └── org/           # 组织架构模块
│   ├── pkg/                # 公共工具包
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

- **语言**: Go 1.21+
- **框架**: Gin (HTTP) + gorilla/websocket
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **消息队列**: Redis Streams（Phase 3）
- **全文搜索**: PostgreSQL ILIKE（Phase 3）
- **文件存储**: 本地文件系统（Phase 3，MinIO 可选）
- **ORM**: GORM
- **配置**: Viper
- **监控**: Prometheus client_golang（Phase 3）

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
- PostgreSQL ILIKE 全文搜索
- 文件上传/下载（本地存储）
- Prometheus 指标监控
- Docker Compose 一键部署

### Phase 4 - 高级特性（待开发）
- 云文档协作编辑（OT 算法）
- 音视频通话（WebRTC）
- 表情回复/线程消息
- 多因素认证

## 快速开始

### 1. 启动基础设施（Docker）

```bash
cd backend/deployments/docker
docker compose up -d

# 验证
docker compose ps
```

### 2. 启动后端

```bash
cd backend
go mod download
go run cmd/server/main.go

# 服务启动在 http://localhost:8080
```

### 3. 启动前端

```bash
cd frontend
pnpm install
pnpm dev

# 前端启动在 http://localhost:3000
```

## API 路由

### 认证
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
| GET | `/api/v1/users/me` | 当前用户信息 | ✅ |
| PUT | `/api/v1/users/me` | 更新个人资料 | ✅ |

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
| POST | `/api/v1/conversations/:id/messages` | 发送消息 | ✅ |
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
| GET | `/api/v1/search/messages?q=keyword` | 搜索消息 | ✅ |

### 其他
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | ❌ |
| GET | `/metrics` | Prometheus 指标 | ❌ |
| WSS | `/ws` | WebSocket | ✅ |

## Docker Compose 服务

```bash
# 启动所有服务
docker compose up -d

# 服务端口
PostgreSQL:  localhost:5432
Redis:       localhost:6379
后端 API:    localhost:8080
前端:       localhost:3000 (开发模式)
Prometheus: localhost:9090 (可选)
```

## 配置说明

主要配置文件：`backend/configs/config.yaml`

```yaml
server:
  port: 8080
  jwt_secret: "your-secret-key"
  jwt_expiry_hours: 72

database:
  host: "localhost"
  port: 5432

redis:
  host: "localhost"
  port: 6379

file:
  store_type: "local"  # local 或 minio
  local_base_path: "/tmp/workpal-files"
  max_file_size_mb: 100
```

## Makefile 命令

```bash
make deps      # 下载 Go 依赖
make run       # 运行后端
make build     # 编译二进制
make docker-up   # 启动 Docker 服务
make docker-down # 停止 Docker 服务
make test      # 运行测试
make lint      # 代码检查
```

## 监控系统（Phase 3）

Prometheus 指标端点：`GET /metrics`

可配合 Grafana 使用，配置 Prometheus 数据源即可。

## License

MIT

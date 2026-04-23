# WorkPal

基于 Go 的企业协作平台（仿飞书/钉钉），适合作为学习项目。

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin (Web)
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **WebSocket**: gorilla/websocket
- **ORM**: GORM
- **配置**: Viper
- **监控**: Prometheus client_golang

## 项目结构

```
WorkPal/
├── cmd/
│   └── server/main.go          # 程序入口
├── configs/
│   ├── config.go               # 配置加载器
│   └── config.yaml             # 配置文件
├── internal/
│   ├── common/                 # 公共组件
│   │   ├── errors/             # 统一错误定义
│   │   ├── middleware/          # Gin 中间件（JWT/CORS/RequestID）
│   │   ├── pagination/          # 分页工具
│   │   └── response/           # 统一响应格式
│   ├── user/                   # 用户模块
│   │   ├── handler/            # HTTP Handler
│   │   ├── service/            # 业务逻辑
│   │   ├── repo/               # 数据访问层
│   │   └── model/              # 数据模型
│   └── im/                     # IM 即时通讯模块（后续开发）
├── pkg/
│   └── auth/                   # JWT 认证工具
├── deployments/
│   └── docker/docker-compose.yaml
├── Makefile
└── README.md
```

## 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose

### 1. 启动基础设施（数据库 + Redis）

```bash
make docker-up
```

### 2. 安装依赖

```bash
make deps
```

### 3. 运行服务

```bash
make run
```

服务启动在 `http://localhost:8080`

### 4. 运行测试

```bash
make test
```

## API 路由

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
| GET | `/api/v1/users/me` | 获取当前用户 | ✅ |
| PUT | `/api/v1/users/me` | 更新个人资料 | ✅ |
| GET | `/api/v1/users` | 用户列表（分页） | ✅ |
| GET | `/health` | 健康检查 | ❌ |
| GET | `/metrics` | Prometheus 指标 | ❌ |

## API 示例

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","nickname":"测试用户"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 获取当前用户（需要 Token）
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>"
```

## 开发指南

### 添加新的模块

参考 `internal/user` 的四层结构：

```
handler  →  service  →  repo  →  model
 (接口层)   (业务层)   (数据层)  (数据模型)
```

1. 在 `internal/` 下创建模块目录
2. 定义 `model.go`（对应数据表结构）
3. 定义 `repo.go`（数据库操作）
4. 定义 `service.go`（业务逻辑）
5. 定义 `handler.go`（HTTP 接口）
6. 在 `main.go` 中初始化并注册路由

### 代码规范

```bash
make lint   # 运行 golangci-lint
make swag   # 生成 Swagger 文档
```

## License

MIT

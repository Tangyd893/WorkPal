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
│   │   └── user/          # 用户模块
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
    ├── 飞书类产品技术分析与Go仿品架构设计.md
    ├── 技术选型文档.md
    └── 阶段四-高级特性详细说明.md
```

## 技术栈

### 后端

- **语言**: Go 1.22+
- **框架**: Gin (HTTP) + gorilla/websocket
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **ORM**: GORM
- **配置**: Viper
- **监控**: Prometheus client_golang

### 前端

- **框架**: React 18 + TypeScript
- **构建**: Vite 5
- **路由**: React Router v6
- **状态**: Zustand

## 快速开始

### 1. 启动后端

```bash
# 安装依赖
make deps

# 启动数据库（需要 Docker）
make docker-up

# 运行
make run
# 服务启动在 http://localhost:8080
```

### 2. 启动前端

```bash
cd frontend

pnpm install
pnpm dev
# 前端启动在 http://localhost:3000
```

## API 路由

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
| GET | `/api/v1/users/me` | 当前用户信息 | ✅ |
| PUT | `/api/v1/users/me` | 更新个人资料 | ✅ |
| GET | `/api/v1/users` | 用户列表 | ✅ |
| GET | `/health` | 健康检查 | ❌ |
| GET | `/metrics` | Prometheus 指标 | ❌ |

## Makefile 常用命令

```bash
make deps         # 下载 Go 依赖
make run          # 运行后端
make build        # 编译二进制
make docker-up    # 启动数据库容器
make docker-down  # 停止数据库容器
make test         # 运行测试
make lint         # 代码检查
```

## 开发指南

### 添加新模块

参考 `backend/internal/user/` 的四层结构：

```
handler  →  service  →  repo  →  model
  接口层     业务层     数据层    数据模型
```

1. 在 `backend/internal/` 下创建模块目录
2. `model.go` — 定义数据表结构
3. `repo.go` — 数据库操作
4. `service.go` — 业务逻辑
5. `handler.go` — HTTP 接口
6. 在 `backend/cmd/server/main.go` 中初始化并注册路由

# WorkPal Backend

这是 WorkPal 的 Go 后端。

如果你的目标是从零启动整套项目，请优先看仓库根目录的 [README.md](../README.md)。这份文档只补充后端独有的信息。

## 环境要求

- Go 1.22+
- Docker Desktop / Docker Engine

## 启动依赖

后端依赖：

- PostgreSQL
- Redis
- MinIO

推荐在仓库根目录执行：

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

## 配置文件

后端按下面顺序读取配置：

1. `CONFIG_PATH`
2. `configs/config.yaml`
3. `configs/config.example.yaml`

因此，默认情况下不复制 `config.yaml` 也能启动。

如果你想改本地配置：

```powershell
Copy-Item configs\config.example.yaml configs\config.yaml
```

## 启动后端

```powershell
cd backend
go run ./cmd/server
```

启动后可验证：

```powershell
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
Invoke-WebRequest http://localhost:8080/ -UseBasicParsing
```

## 开发模式预置账号

当 `server.mode != release` 时，后端启动会自动确保这些账号存在：

| 角色 | 用户名 | 密码 |
|---|---|---|
| 管理员 | `admin` | `admin123` |
| 员工 | `emma.chen` | `workpal123` |
| 员工 | `liam.wang` | `workpal123` |
| 员工 | `sofia.zhao` | `workpal123` |

这意味着你每次本地重启后端，都可以直接用这些账号做验收与联调，不必再手动注册测试用户。

## 重要行为

- 启动时自动执行 GORM `AutoMigrate`
- `/health` 同时检查 PostgreSQL 与 Redis
- `/metrics` 提供 Prometheus 指标
- WebSocket 地址为 `/ws?token=...`

## 常用接口

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/health` | 健康检查 |
| `GET` | `/metrics` | Prometheus 指标 |
| `POST` | `/api/v1/auth/register` | 注册 |
| `POST` | `/api/v1/auth/login` | 登录 |
| `GET` | `/api/v1/users/me` | 当前用户 |
| `GET` | `/api/v1/users` | 用户列表 |
| `GET` | `/api/v1/users/search` | 用户搜索 |
| `WS` | `/ws?token=...` | 实时消息 |

## 登录示例

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

## 常用命令

```powershell
cd backend
go test ./...
go test -race ./...
go build ./cmd/server
```

## Makefile

如果你的环境里有 `make`，也可以使用：

```bash
cd backend
make deps
make docker-up
make run
make test
make lint
```

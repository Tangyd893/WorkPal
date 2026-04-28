# WorkPal Backend

这是 WorkPal 的 Go 后端。

如果你的目标是从零把整套项目跑起来，请优先看仓库根目录的 [README.md](../README.md)。  
这份文档只补充后端独有的信息。

## 环境要求

- Go 1.22+
- Docker Desktop / Docker Engine

## 启动前依赖

后端依赖：

- PostgreSQL
- Redis
- MinIO

推荐直接在仓库根目录启动：

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

## 配置文件

后端按下面顺序读取配置：

1. `CONFIG_PATH` 指向的文件
2. `configs/config.yaml`
3. `configs/config.example.yaml`

所以直接运行时，即使你没有复制 `config.yaml`，也会回退到样例配置。

如果要自定义本地配置：

```powershell
Copy-Item configs\config.example.yaml configs\config.yaml
```

默认样例配置适配仓库里的 Docker Compose：

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MinIO: `localhost:9000`

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

## 重要事实

- 后端会在启动时自动执行 GORM `AutoMigrate`
- `/health` 会同时检查 PostgreSQL 和 Redis
- WebSocket 地址是 `/ws?token=...`
- 在默认开发配置下，后端启动时会自动确保默认管理员账号存在：
  - 用户名：`admin`
  - 密码：`admin123`
- 这条默认管理员逻辑仅适用于 `server.mode != release` 的本地开发/验收场景

## 常用命令

```powershell
cd backend
go test ./...
go test -race ./...
go build ./cmd/server
```

## 可选的 Makefile

如果你的环境里有 `make`，也可以用：

```bash
cd backend
make deps
make docker-up
make run
make test
make lint
```

注意：

- `make docker-up` 依赖 Docker
- 这个 Makefile 主要偏 Unix 风格；Windows 上更推荐直接用上面的 PowerShell 命令

## 常见接口

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/health` | 健康检查 |
| `GET` | `/metrics` | Prometheus 指标 |
| `POST` | `/api/v1/auth/register` | 注册 |
| `POST` | `/api/v1/auth/login` | 登录 |
| `GET` | `/api/v1/users/me` | 当前用户 |
| `WS` | `/ws?token=...` | WebSocket |

## 默认管理员登录示例

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

## 可选：创建额外测试账号

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

建议每次调试都带一个新的 `email`，因为当前 `users.email` 是唯一索引。
